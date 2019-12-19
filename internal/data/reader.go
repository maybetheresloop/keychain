package data

import (
	"encoding/binary"
	"io"
)

type Reader struct {
	rd     io.Reader
	offset uint64
}

func NewReader(rd io.Reader) *Reader {
	return &Reader{
		rd: rd,
	}
}

// Reads a Keychain item from the reader.
func (r *Reader) ReadItem() (item *Item, err error) {

	// Read the timestamp.
	var timestamp uint64
	if err := binary.Read(r.rd, binary.BigEndian, &timestamp); err != nil {
		return nil, err
	}

	r.offset += 8

	// Read the key size.
	var keySize int64
	if err := binary.Read(r.rd, binary.BigEndian, &keySize); err != nil {
		return nil, err
	}

	r.offset += 8

	var valueSize int64
	if err := binary.Read(r.rd, binary.BigEndian, &valueSize); err != nil {
		return nil, err
	}

	r.offset += 8

	key := make([]byte, keySize)
	_, err = io.ReadFull(r.rd, key)
	if err != nil {
		return nil, err
	}

	value := make([]byte, valueSize)
	if _, err := io.ReadFull(r.rd, value); err != nil {
		return nil, err
	}

	item = &Item{
		KeySize:   keySize,
		ValueSize: valueSize,
		Key:       key,
		Value:     value,
	}

	return
}
