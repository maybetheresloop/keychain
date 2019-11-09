package internal

import (
	"bytes"
	"errors"
	"os"
	"sync"

	"github.com/maybetheresloop/keychain/internal/proto"
	art "github.com/plar/go-adaptive-radix-tree"
)

type Keychain struct {
	sync.RWMutex

	readHandle  *os.File
	writeHandle *os.File
	writeBuffer *proto.Writer
	entries     art.Tree
	counter     uint64
	offset      int64
}

func valueOffset(keyLen int) int64 {
	return int64(2*8 + keyLen)
}

func Open(name string) (*Keychain, error) {
	writeHandle, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	readHandle, err := os.Open(name)

	stat, err := writeHandle.Stat()
	if err != nil {
		return nil, err
	}

	offset := stat.Size()

	if err != nil {
		return nil, err
	}

	// Populate entries
	entries := art.New()

	r := proto.NewEntryReader(readHandle, 1)

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

func (k *Keychain) appendItem(key []byte, value []byte) error {
	if err := k.writeBuffer.WriteItem(proto.NewItem(key, value)); err != nil {
		return err
	}

	if err := k.writeBuffer.Flush(); err != nil {
		return err
	}

	k.offset += int64(2*8 + len(key) + len(value))

	return nil
}

func (k *Keychain) appendItemDelete(key []byte) error {
	if err := k.writeBuffer.WriteItem(proto.NewItemDeleteMarker(key)); err != nil {
		return err
	}

	if err := k.writeBuffer.Flush(); err != nil {
		return err
	}

	k.offset += int64(2*8 + len(key))

	return nil
}

func (k *Keychain) Set(key []byte, value []byte) error {
	k.Lock()
	defer k.Unlock()
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

func (k *Keychain) Get(key []byte) ([]byte, error) {
	k.RLock()
	v, ok := k.entries.Search(key)
	if !ok {
		k.RUnlock()
		return nil, nil
	}

	defer k.RUnlock()

	entry := v.(*proto.Entry)
	if entry.ValueSize == -1 {
		return nil, nil
	}

	return k.readValue(entry.ValuePos, entry.ValueSize)
}

func (k *Keychain) Remove(key []byte) (bool, error) {
	k.Lock()

	v, found := k.entries.Search(key)
	if found {
		entry := v.(*proto.Entry)
		if entry.ValueSize != -1 {
			defer k.Unlock()

			if err := k.appendItemDelete(key); err != nil {
				return false, err
			}

			entry.ValueSize = -1
			return true, nil
		}
	}

	k.Unlock()
	return false, nil
}

func (k *Keychain) Flush() error {
	return k.writeBuffer.Flush()
}

func (k *Keychain) Close() error {
	_ = k.Flush()

	_ = k.readHandle.Close()
	_ = k.writeHandle.Close()

	return nil
}
