package keychain

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/maybetheresloop/keychain/internal/data"
	art "github.com/plar/go-adaptive-radix-tree"
)

// Conf represents the configuration options for a Keychain store.
type Conf struct {
	Sync  bool
	clock data.Clock
}

type defaultClock struct{}

func (d defaultClock) Now() time.Time {
	return time.Now()
}

// Keychain represents an instance of a Keychain store.
type Keychain struct {
	mtx         sync.RWMutex
	readHandle  *os.File
	writeHandle *os.File
	writeBuffer *data.Writer
	entries     art.Tree
	counter     uint64
	offset      int64
	sync        bool
}

// Opens a Keychain store using the specified file path and configuration. If the file does not exist,
// then it is created.
func OpenConf(name string, conf *Conf) (*Keychain, error) {

	// Two handles to the file: one is used for reading, the other is used for writing.
	writeHandle, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	readHandle, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	// Also want to get the size of the file, so that we can keep track of the offsets
	// of values in the file.
	stat, err := writeHandle.Stat()
	if err != nil {
		return nil, err
	}
	offset := stat.Size()

	entries := art.New()

	r := data.NewEntryReader(readHandle, 1)

	// Populate radix tree with entries from the database file.
	var entry *data.Entry
	var key []byte
	for key, entry, err = r.ReadEntry(); err == nil; key, entry, err = r.ReadEntry() {
		entries.Insert(key, art.Value(entry))
	}

	var clock data.Clock = defaultClock{}
	if conf != nil && conf.clock != nil {
		clock = conf.clock
	}

	keys := &Keychain{
		readHandle:  readHandle,
		writeHandle: writeHandle,
		writeBuffer: data.NewWriter(writeHandle, clock),
		entries:     entries,
		offset:      offset,
		sync:        false,
	}

	if conf != nil {
		keys.sync = conf.Sync
	}

	return keys, nil
}

// Opens a Keychain store using the specified file path. If the file does not exist, then
// it is created.
func Open(name string) (*Keychain, error) {
	return OpenConf(name, nil)
}

// append is used internally by appendItem* and does the actual appending and flushing of
// the underlying buffer. Additionally, this will call Sync() on the underlying file
// so that the new item is synchronized to disk.
func (k *Keychain) append(item *data.Item) error {
	if err := k.writeBuffer.WriteItem(item); err != nil {
		return err
	}

	if err := k.writeBuffer.Flush(); err != nil {
		return err
	}

	// Sync the new items to the underlying storage.
	if err := k.writeHandle.Sync(); err != nil {
		return err
	}

	k.offset += valueOffset(len(item.Key))
	if len(item.Value) > 0 {
		k.offset += int64(len(item.Value))
	}

	return nil
}

// appendItem appends a key-value pair to the end of the store file's log.
func (k *Keychain) appendItem(key []byte, value []byte) error {
	return k.append(data.NewItem(key, value))
}

// appendItemDelete appends a special delete marker for the specified key.
func (k *Keychain) appendItemDelete(key []byte) error {
	return k.append(data.NewItemDeleteMarker(key))
}

// Set inserts a key-value pair into the store. If the key already exists in the store, then
// the previous value is overwritten.
func (k *Keychain) Set(key []byte, value []byte) error {
	k.mtx.Lock()

	v, found := k.entries.Search(key)
	valuePos := k.offset + valueOffset(len(key))
	valueSize := int64(len(value))

	// We insert the new value unconditionally, even if the key was already present
	// in the database with the same value. Otherwise, we would have to do a disk seek
	// to check the current value, and in this case we have decided to optimize for performance
	// and not for space.
	if err := k.appendItem(key, value); err != nil {
		k.mtx.Unlock()
		return err
	}

	// If the trie already contains the entry, simply update the existing entry. Otherwise,
	// insert the new entry into the trie.
	if found {
		entry := v.(*data.Entry)
		entry.ValuePos = valuePos
		entry.ValueSize = valueSize

		k.mtx.Unlock()
		return nil
	}

	entry := data.NewEntry(0, valueSize, valuePos)
	k.entries.Insert(key, art.Value(entry))

	k.mtx.Unlock()
	return nil
}

// Reads a value of the given size at the offset.
func (k *Keychain) readValue(offset int64, size int64) ([]byte, error) {
	value := make([]byte, size)
	n, err := k.readHandle.ReadAt(value, offset)
	if err != nil {
		return nil, err
	}

	if int64(n) != size {
		return nil, errors.New("could not read full value")
	}

	return value, nil
}

// Get retrieves from the store the value corresponding to the specified key. If the key does not
// exist, then nil is returned.
func (k *Keychain) Get(key []byte) ([]byte, error) {
	k.mtx.RLock()
	v, ok := k.entries.Search(key)
	if !ok {
		k.mtx.RUnlock()
		return nil, nil
	}

	defer k.mtx.RUnlock()

	entry := v.(*data.Entry)
	if entry.ValueSize == -1 {
		return nil, nil
	}

	return k.readValue(entry.ValuePos, entry.ValueSize)
}

// Removes a key-value pair from the store. Returns true only if an item was removed.
func (k *Keychain) Remove(key []byte) (bool, error) {
	k.mtx.Lock()

	v, found := k.entries.Search(key)
	if found {
		entry := v.(*data.Entry)
		if entry.ValueSize != -1 {
			defer k.mtx.Unlock()

			if err := k.appendItemDelete(key); err != nil {
				return false, err
			}

			entry.ValueSize = -1
			return true, nil
		}
	}

	k.mtx.Unlock()
	return false, nil
}

// Flushes the underlying write buffer.
func (k *Keychain) Flush() error {
	return k.writeBuffer.Flush()
}

// Closes the store.
func (k *Keychain) Close() error {
	if err := k.Flush(); err != nil {
		return err
	}

	if err := k.readHandle.Close(); err != nil {
		return err
	}

	if err := k.writeHandle.Close(); err != nil {
		return err
	}

	return nil
}

// Returns the offset of the value in an item with a key of the specified len.
// This works out to the following:
//
//   timestamp (8 bytes) + key size (8 bytes) + value size (8 bytes) + key (key size bytes)
func valueOffset(keyLen int) int64 {
	return int64(3*8 + keyLen)
}
