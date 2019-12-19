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
	wr    io.Writer
	clock Clock
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

func (w *Writer) WriteItem(item *Item) error {

	if err := binary.Write(w.wr, binary.BigEndian, w.clock.Now().UnixNano()); err != nil {
		return err
	}

	if err := binary.Write(w.wr, binary.BigEndian, item.KeySize); err != nil {
		return err
	}

	if err := binary.Write(w.wr, binary.BigEndian, item.ValueSize); err != nil {
		return err
	}

	if item.KeySize > 0 {
		if _, err := w.wr.Write(item.Key); err != nil {
			return err
		}
	}

	if item.ValueSize > 0 {
		if _, err := w.wr.Write(item.Value); err != nil {
			return err
		}
	}

	return nil
}

//func (w *Writer) Flush() error {
//	return w.wr.Flush()
//}
