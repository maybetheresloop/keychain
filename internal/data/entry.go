package data

import (
	"encoding/binary"
	"io"
)

type Entry struct {
	Timestamp int64
	FileID    uint64
	ValueSize int64
	ValuePos  int64
}

func NewEntry(fileID uint64, valueSize int64, valuePos int64, timestamp int64) *Entry {
	return &Entry{
		FileID:    fileID,
		ValueSize: valueSize,
		ValuePos:  valuePos,
	}
}

type HintWriter struct {
	wr io.Writer
}

func NewHintWriter(wr io.Writer) *HintWriter {
	return &HintWriter{wr: wr}
}

func uint64Bytes(n uint64) [8]byte {
	var res [8]byte
	binary.BigEndian.PutUint64(res[:], n)
	return res
}

func (w *HintWriter) WriteFromEntry(key []byte, entry *Entry) error {
	b := uint64Bytes(uint64(entry.Timestamp))
	if _, err := w.wr.Write(b[:]); err != nil {
		return err
	}

	b = uint64Bytes(uint64(len(key)))
	if _, err := w.wr.Write(b[:]); err != nil {
		return err
	}

	b = uint64Bytes(uint64(entry.ValueSize))
	if _, err := w.wr.Write(b[:]); err != nil {
		return err
	}

	b = uint64Bytes(uint64(entry.ValuePos))
	if _, err := w.wr.Write(b[:]); err != nil {
		return err
	}

	if _, err := w.wr.Write(key); err != nil {
		return err
	}

	return nil
}
