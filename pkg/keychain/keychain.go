package keychain

import (
	"bytes"
	"errors"
	"os"
	"sync"

	"github.com/maybetheresloop/keychain/internal/proto"
	art "github.com/plar/go-adaptive-radix-tree"
)

// Keychain represents an instance of a Keychain store.
type Keychain struct {
	mtx         sync.RWMutex
	readHandle  *os.File
	writeHandle *os.File
	writeBuffer *proto.Writer
	entries     art.Tree
	counter     uint64
	offset      int64
}

// Opens a Keychain store using the specified file path. If the file does not exist, then
// it is created.
func Open(name string) (*Keychain, error) {

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

	r := proto.NewEntryReader(readHandle, 1)

	// Populate radix tree with entries from the database file.
	var entry *proto.Entry
	var key []byte
	for key, entry, err = r.ReadEntry(); err == nil; key, entry, err = r.ReadEntry() {
		entries.Insert(key, art.Value(entry))
	}

	return &Keychain{
		readHandle:  readHandle,
		writeHandle: writeHandle,
		writeBuffer: proto.NewWriter(writeHandle),
		entries:     entries,
		offset:      offset,
	}, nil
}

// append is used internally by appendItem* and does the actual appending and flushing of
// the underlying buffer.
func (k *Keychain) append(item *proto.Item) error {
	if err := k.writeBuffer.WriteItem(item); err != nil {
		return err
	}

	if err := k.writeBuffer.Flush(); err != nil {
		return err
	}

	k.offset += int64(2*8 + len(item.Key))
	if len(item.Value) > 0 {
		k.offset += int64(len(item.Value))
	}

	return nil
}

// appendItem appends a key-value pair to the end of the store file's log.
// The underlying writer is flushed after the append is done.
func (k *Keychain) appendItem(key []byte, value []byte) error {
	return k.append(proto.NewItem(key, value))
}

// appendItemDelete appends a special delete marker for the specified key.
func (k *Keychain) appendItemDelete(key []byte) error {
	return k.append(proto.NewItemDeleteMarker(key))
}

// Set inserts a key-value pair into the store. If the key already exists in the store, then
// the previous value is overwritten.
func (k *Keychain) Set(key []byte, value []byte) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	v, found := k.entries.Search(key)
	if !found {
		valuePos := k.offset + valueOffset(len(key))
		if err := k.appendItem(key, value); err != nil {
			return err
		}

		entry := &proto.Entry{
			FileID:    0,
			ValueSize: int64(len(value)),
			ValuePos:  valuePos,
		}

		k.entries.Insert(key, art.Value(entry))

		return nil
	}

	entry := v.(*proto.Entry)

	if entry.ValueSize == -1 {
		goto insert
	} else {
		oldValue, err := k.readValue(entry.ValuePos, entry.ValueSize)
		if err != nil {
			return err
		}

		if bytes.Compare(value, oldValue) != 0 {
			goto insert
		}
	}

	return nil

insert:
	valuePos := k.offset + valueOffset(len(key))
	if err := k.appendItem(key, value); err != nil {
		return err
	}

	entry.ValuePos = valuePos
	entry.ValueSize = int64(len(value))

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

	entry := v.(*proto.Entry)
	if entry.ValueSize == -1 {
		return nil, nil
	}

	return k.readValue(entry.ValuePos, entry.ValueSize)
}

func (k *Keychain) Remove(key []byte) (bool, error) {
	k.mtx.Lock()

	v, found := k.entries.Search(key)
	if found {
		entry := v.(*proto.Entry)
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

func (k *Keychain) Close() error {
	_ = k.Flush()

	_ = k.readHandle.Close()
	_ = k.writeHandle.Close()

	return nil
}

func valueOffset(keyLen int) int64 {
	return int64(2*8 + keyLen)
}
