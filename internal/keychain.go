package internal

import (
	"bufio"
	"errors"
	"os"

	"github.com/maybetheresloop/keychain/internal/proto"
	art "github.com/plar/go-adaptive-radix-tree"
)

type Keychain struct {
	readHandle  *os.File
	writeHandle *os.File
	writeBuffer *bufio.Writer
	entries     art.Tree
	counter     uint64
}

func Open(name string) (*Keychain, error) {
	readHandle, err := os.Open(name)
	writeHandle, err := os.OpenFile(name, os.O_APPEND, 0644)
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
		writeBuffer: bufio.NewWriter(writeHandle),
		entries:     entries,
	}, nil
}

func (k *Keychain) Set(key []byte) ([]byte, error) {

	return nil, nil
}

func (k *Keychain) Get(key []byte) ([]byte, error) {
	v, ok := k.entries.Search(key)
	if !ok {
		return nil, nil
	}

	entry := v.(proto.Entry)
	if entry.ValueSize == -1 {
		return nil, nil
	}

	value := make([]byte, entry.ValueSize)
	n, err := k.readHandle.ReadAt(value, entry.ValuePos)
	if err != nil {
		return nil, err
	}

	if int64(n) != entry.ValueSize {
		return nil, errors.New("could not read full value")
	}

	return value, nil
}

func (k *Keychain) Flush() error {
	return k.writeBuffer.Flush()
}

func (k *Keychain) Close() error {
	_ = k.Flush()

	_ = k.readHandle.Close()
	_ = k.writeHandle.Close()
}
