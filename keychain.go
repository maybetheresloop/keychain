package keychain

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/maybetheresloop/keychain/internal/data/handle"

	"github.com/maybetheresloop/keychain/internal/data"
	art "github.com/plar/go-adaptive-radix-tree"
	log "github.com/sirupsen/logrus"
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
	idCounter   uint64
	clock       data.Clock
	mtx         sync.RWMutex
	readHandle  *os.File
	writeHandle *os.File
	writeBuffer *data.Writer

	handles map[int64]*handle.Handle
	active  *handle.Handle

	entries art.Tree
	counter uint64
	offset  int64
	sync    bool
}

// Gets the list of paths that make up the database files from the specified directory.
// A database file name is of the form <timestamp>.keychain.dat, where <timestamp> is a
// Unix timestamp in nanoseconds. Also considered are hint files, which may accompany a
// data file. A hint file has a name of the form <timestamp>.keychain.hint.
func scanDatabaseFilePaths(dirname string) ([]string, []string, error) {
	hintPaths := make([]string, 0)
	dataPaths := make([]string, 0)

	if err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && path != dirname {
			return filepath.SkipDir
		}

		if strings.HasSuffix(path, ".data") {
			dataPaths = append(dataPaths, path)
		} else if strings.HasSuffix(path, ".hint") {
			hintPaths = append(hintPaths, path)
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	return hintPaths, dataPaths, nil
}

// Opens a Keychain store using the specified file path and configuration. If the file does not exist,
// then it is created.
func OpenConf(dirname string, conf *Conf) (*Keychain, error) {

	entries := art.New()
	handles := make(map[int64]*handle.Handle)

	hasHintFile := make(map[int64]bool)

	stat, err := os.Stat(dirname)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		return nil, errors.New("must refer to a directory")
	}

	// Get the lists of database files in the database directory.
	hintPaths, dataPaths, err := scanDatabaseFilePaths(dirname)
	if err != nil {
		return nil, err
	}

	// Populate the in-memory index, processing .hint files first.
	for _, path := range hintPaths {
		fileId, err := strconv.ParseInt(path[:strings.IndexByte(path, '.')], 10, 64)
		if err != nil {
			log.Warnf("Skipping hint file w/ malformed path: %s", path)
		}

		if err := addHintFileEntries(entries, path, fileId); err != nil {
			return nil, err
		}

		// Add the file ID to the file IDs we have seen.
		hasHintFile[fileId] = true
	}

	// Process .data files next. If the .data file has a corresponding .hint file that we already
	// processed, then skip adding its entries.
	for _, path := range dataPaths {
		fileId, err := strconv.ParseInt(path[:strings.IndexByte(path, '.')], 10, 64)
		if err != nil {
			log.Warnf("Skipping data file w/ malformed path: %s", path)
		}

		// If we haven't seen this file ID yet, open the file and add its entries to our key
		if !hasHintFile[fileId] {
			if err := addDataFileEntries(entries, path, fileId); err != nil {
				return nil, err
			}
		}

		// Create a read-only handle for the file.
		h, err := handle.Open(path, true)
		if err != nil {
			return nil, err
		}

		handles[fileId] = h
	}

	// Create the active database file.
	activeId := time.Now().UnixNano()
	activeName := fmt.Sprintf("%d.keychain.data", activeId)

	h, err := handle.Open(activeName, false)
	if err != nil {
		return nil, err
	}

	handles[activeId] = h

	//// Two handles to the file: one is used for reading, the other is used for writing.
	//writeHandle, err := os.OpenFile(dirname, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	//if err != nil {
	//	return nil, err
	//}
	//
	//readHandle, err := os.Open(dirname)
	//if err != nil {
	//	return nil, err
	//}
	//
	//// Also want to get the size of the file, so that we can keep track of the offsets
	//// of values in the file.
	//stat, err = writeHandle.Stat()
	//if err != nil {
	//	return nil, err
	//}
	//offset := stat.Size()
	//
	//entries := art.New()
	//
	//r := data.NewEntryReader(readHandle, 1)
	//
	//// Populate radix tree with entries from the database file.
	//var entry *data.Entry
	//var key []byte
	//
	//for key, entry, err = r.ReadEntry(); err == nil; key, entry, err = r.ReadEntry() {
	//
	//	// If an entry with the same key does not yet exist in the radix tree, insert it.
	//	oldEntry, ok := entries.Search(key)
	//	if !ok {
	//		entries.Insert(key, art.Value(entry))
	//	}
	//
	//	// Otherwise, check if the entry currently in the tree is older than the entry just
	//	// read. If the existing entry is older, overwrite it with the fields of the new
	//	// entry.
	//	v := oldEntry.(*data.Entry)
	//	if v.Timestamp <= entry.Timestamp {
	//		*v = *entry
	//	}
	//}
	//
	//var clock data.Clock = defaultClock{}
	//if conf != nil && conf.clock != nil {
	//	clock = conf.clock
	//}
	//
	//keys := &Keychain{
	//	clock:       clock,
	//	readHandle:  readHandle,
	//	writeHandle: writeHandle,
	//	writeBuffer: data.NewWriter(writeHandle, clock),
	//	entries:     entries,
	//	offset:      offset,
	//	sync:        false,
	//}
	//
	//if conf != nil {
	//	keys.sync = conf.Sync
	//}
	//
	//return keys, nil
	return nil, nil
}

func addHintFileEntries(entries art.Tree, path string, fileId int64) error {
	// Open the file, and add it's entries to our key map.
	r, err := os.Open(path)
	if err != nil {
		return err
	}
	defer r.Close()

	rd := data.NewHintFileEntryReader(r, fileId)
	for {
		key, entry, err := rd.ReadEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		insertEntry(entries, key, entry)
	}

	return nil
}

func addDataFileEntries(entries art.Tree, path string, fileId int64) error {
	r, err := os.Open(path)
	if err != nil {
		return err
	}

	rd := data.NewEntryReader(r, uint64(fileId))
	for {
		key, entry, err := rd.ReadEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			if err2 := r.Close(); err2 != nil {
				return err2
			}
			return err
		}
		insertEntry(entries, key, entry)
	}

	if err := r.Close(); err != nil {
		return err
	}

	return nil
}

// Inserts a key-entry pair into the provided entry map. If an entry already
// exists for the given key, it will be overwritten if the entry passed in is
// newer than the existing entry.
func insertEntry(entries art.Tree, key []byte, entry *data.Entry) {
	old, ok := entries.Search(key)
	if !ok {
		entries.Insert(key, art.Value(entry))
	}

	if old.(*data.Entry).Timestamp <= entry.Timestamp {
		*old.(*data.Entry) = *entry
	}
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
func (k *Keychain) appendItem(key []byte, value []byte, timestamp int64) error {
	return k.append(data.NewItem(key, value, timestamp))
}

// appendItemDelete appends a special delete marker for the specified key.
func (k *Keychain) appendItemDelete(key []byte, timestamp int64) error {
	return k.append(data.NewItemDeleteMarker(key, timestamp))
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
	ts := k.clock.Now().UnixNano()

	if err := k.appendItem(key, value, ts); err != nil {
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

	entry := data.NewEntry(0, valueSize, valuePos, ts)
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

			if err := k.appendItemDelete(key, k.clock.Now().UnixNano()); err != nil {
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
