package data

import (
	"bufio"
	"encoding/binary"
	"io"
	"io/ioutil"
)

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
