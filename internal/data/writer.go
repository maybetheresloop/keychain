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

type defaultClock struct{}

func (d defaultClock) Now() time.Time {
	return time.Now()
}

type Writer struct {
	wr    *bufio.Writer
	clock Clock
}

func NewWriter(wr io.Writer, clock Clock) *Writer {
	if clock == nil {
		clock = defaultClock{}
	}

	return &Writer{
		wr: bufio.NewWriter(wr),
	}
}

func NewWriterFromBuffered(wr *bufio.Writer) *Writer {
	return &Writer{wr: wr}
}

func (w *Writer) WriteItem(item *Item) error {

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

func (w *Writer) Flush() error {
	return w.wr.Flush()
}
