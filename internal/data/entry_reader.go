package data

import (
	"bufio"
	"encoding/binary"
	"io"
	"io/ioutil"
)

// Reader for .hint files, which consist of only keys and value file offsets.
type HintFileEntryReader struct {

	// Internal buffered reader.
	rd io.Reader

	fileId int64 // File ID to assign to the deserialized entries.
}

func NewHintFileEntryReader(rd io.Reader, fileId int64) *HintFileEntryReader {
	return &HintFileEntryReader{
		rd:     rd,
		fileId: fileId,
	}
}

// Reads an entry from the .hint file. This consists of reading the timestamp,
// the key and value sizes, the value position, and the key itself.
func (r *HintFileEntryReader) ReadEntry() ([]byte, *Entry, error) {
	// Read the timestamp.
	var timestamp int64
	if err := binary.Read(r.rd, binary.BigEndian, &timestamp); err != nil {
		return nil, nil, err
	}

	// Read the key size.
	var keySize int64
	if err := binary.Read(r.rd, binary.BigEndian, &keySize); err != nil {
		return nil, nil, err
	}

	// Read the value size.
	var valueSize int64
	if err := binary.Read(r.rd, binary.BigEndian, &valueSize); err != nil {
		return nil, nil, err
	}

	// Read the value position.
	var valuePos int64
	if err := binary.Read(r.rd, binary.BigEndian, &valuePos); err != nil {
		return nil, nil, err
	}

	// Read the key.
	key := make([]byte, keySize)
	if _, err := io.ReadFull(r.rd, key); err != nil {
		return nil, nil, err
	}

	return key, &Entry{
		Timestamp: timestamp,
		FileID:    uint64(r.fileId),
		ValueSize: valueSize,
		ValuePos:  valuePos,
	}, nil
}

type EntryReader struct {
	rd     *bufio.Reader
	offset int64
	fileID uint64
}

func NewEntryReader(rd io.Reader, fileID uint64) *EntryReader {
	return &EntryReader{
		rd:     bufio.NewReader(rd),
		offset: 0,
		fileID: fileID,
	}
}

// Reads an entry from the Keychain database. To read an entry, we need only
// read the timestamp, the key size, the value size, and the key itself, and
// we can simply discard the value bytes.
func (r *EntryReader) ReadEntry() (key []byte, entry *Entry, err error) {
	var timestamp int64
	if err = binary.Read(r.rd, binary.BigEndian, &timestamp); err != nil {
		return
	}

	r.offset += 8

	var keySize int64
	if err = binary.Read(r.rd, binary.BigEndian, &keySize); err != nil {
		return
	}

	r.offset += 8

	var valueSize int64
	if err = binary.Read(r.rd, binary.BigEndian, &valueSize); err != nil {
		return
	}

	r.offset += 8

	key = make([]byte, keySize)
	n, err := io.ReadFull(r.rd, key)
	if err != nil {
		return
	}

	r.offset += int64(n)
	valuePos := r.offset

	n2, err := io.CopyN(ioutil.Discard, r.rd, valueSize)
	if err != nil {
		return
	}

	r.offset += n2

	entry = &Entry{
		FileID:    r.fileID,
		ValueSize: valueSize,
		ValuePos:  valuePos,
	}

	return
}
