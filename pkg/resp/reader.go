package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	SimpleString = '+'
	Error        = '-'
	Integer      = ':'
	BulkString   = '$'
	Array        = '*'
)

type Reader struct {
	rd *bufio.Reader
}

func (r *Reader) ReadMessage() (interface{}, error) {
	line, _, err := r.rd.ReadLine()
	if err != nil {
		return nil, err
	}

	switch line[0] {
	case SimpleString:
		return string(line[1:]), nil
	case Error:
		return NewRespError(string(line[1:])), nil
	case Integer:
		return strconv.ParseInt(string(line[1:]), 10, 64)
	case BulkString:
		length, err := strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return nil, err
		}

		return r.readBulkString(length)
	case Array:
		length, err := strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return nil, err
		}

		return r.readArray(length)
	}

	return nil, fmt.Errorf("resp: failed to parse %q", line)
}

func (r *Reader) readBulkString(length int64) ([]byte, error) {
	b := make([]byte, length)

	_, err := io.ReadFull(r.rd, b)
	if err != nil {
		return nil, err
	}

	line, _, err := r.rd.ReadLine()
	if err != nil {
		return nil, err
	}

	// The line should be empty.
	if len(line) != 0 {
		return nil, fmt.Errorf("incorrect number of bytes in bulk string")
	}

	return b, nil
}

func (r *Reader) readArray(length int64) ([]interface{}, error) {
	s := make([]interface{}, length)
	for i := int64(0); i < length; i++ {
		item, err := r.ReadMessage()
		if err != nil {
			return nil, err
		}

		s[i] = item
	}

	return s, nil
}

func NewReader(rd io.Reader) *Reader {
	return &Reader{rd: bufio.NewReader(rd)}
}
