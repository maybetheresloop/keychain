package resp

import (
	"bufio"
	"io"
	"strconv"
)

type Writer struct {
	wr *bufio.Writer
}

func NewWriter(wr io.Writer) *Writer {
	return &Writer{
		wr: bufio.NewWriter(wr),
	}
}

func (w *Writer) Flush() error {
	return w.wr.Flush()
}

func (w *Writer) WriteMessage(message interface{}) error {
	switch v := message.(type) {
	case string:
		return w.WriteSimpleString(v)
	case RespError:
		return w.WriteError(v)
	case []byte:
		return w.WriteBulkString(v)
	case []interface{}:
		return w.WriteArray(v)
	default:
		return ErrInvalidType(message)
	}
}

func (w *Writer) WriteArray(s []interface{}) error {
	if s == nil {
		_, err := w.wr.Write([]byte("*-1\r\n"))
		return err
	}

	l := len(s)
	if err := w.wr.WriteByte(Array); err != nil {
		return err
	}

	if _, err := w.wr.WriteString(strconv.Itoa(l)); err != nil {
		return err
	}

	for i := 0; i < l; i++ {
		if err := w.WriteMessage(s[i]); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) WriteSimpleString(s string) error {
	if err := w.wr.WriteByte(SimpleString); err != nil {
		return err
	}

	_, err := w.wr.WriteString(s)
	if err != nil {
		return err
	}

	return w.writeCRLF()
}

func (w *Writer) WriteError(respError RespError) error {
	if err := w.wr.WriteByte(Error); err != nil {
		return err
	}

	_, err := w.wr.WriteString(respError.message)
	if err != nil {
		return err
	}

	return w.writeCRLF()
}

func (w *Writer) WriteBulkString(b []byte) error {
	if err := w.wr.WriteByte(BulkString); err != nil {
		return err
	}

	if b == nil {
		if err := w.wr.WriteByte('-'); err != nil {
			return err
		}

		if err := w.wr.WriteByte('1'); err != nil {
			return err
		}

		return w.writeCRLF()
	}

	if _, err := w.wr.WriteString(strconv.Itoa(len(b))); err != nil {
		return err
	}

	if err := w.writeCRLF(); err != nil {
		return err
	}

	if _, err := w.wr.Write(b); err != nil {
		return err
	}

	_, err := w.wr.Write([]byte("\r\n"))
	return err
}

func (w *Writer) writeCRLF() error {
	if err := w.wr.WriteByte('\r'); err != nil {
		return err
	}

	return w.wr.WriteByte('\n')
}
