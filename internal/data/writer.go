package data

import (
	"bufio"
	"encoding/binary"
	"io"
	"time"
)

type Clock interface {
	Now() time.Time
}

type Writer struct {
	wr     io.Writer
	clock  Clock
	offset int64
}

func NewWriter(wr io.Writer, clock Clock) *Writer {
	return &Writer{
		clock: clock,
		wr:    wr,
	}
}

func NewWriterFromBuffered(wr *bufio.Writer) *Writer {
	return &Writer{wr: wr}
}

func (w *Writer) WriteItem(item *Item) (int64, error) {
	valueOffset := w.offset + valueOffset(item.KeySize)

	if err := binary.Write(w.wr, binary.BigEndian, w.clock.Now().UnixNano()); err != nil {
		return -1, err
	}

	if err := binary.Write(w.wr, binary.BigEndian, item.KeySize); err != nil {
		return -1, err
	}

	if err := binary.Write(w.wr, binary.BigEndian, item.ValueSize); err != nil {
		return -1, err
	}

	if item.KeySize > 0 {
		if _, err := w.wr.Write(item.Key); err != nil {
			return -1, err
		}
	}

	if item.ValueSize > 0 {
		if _, err := w.wr.Write(item.Value); err != nil {
			return -1, err
		}
	}

	w.offset += itemSize(item)
	return valueOffset, nil
}

func itemSize(item *Item) int64 {
	return 3*8 + int64(len(item.Key)+len(item.Value))
}

func valueOffset(keyLen int64) int64 {
	return 3*8 + keyLen
}

//func (w *Writer) Flush() error {
//	return w.wr.Flush()
//}
