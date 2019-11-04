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

		return r.ReadBulkString(length)
	}

	return nil, fmt.Errorf("resp: failed to parse %q", line)
}

func (r *Reader) ReadBulkString(length int64) ([]byte, error) {
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

func NewReader(rd io.Reader) *Reader {
	return &Reader{rd: bufio.NewReader(rd)}
}
