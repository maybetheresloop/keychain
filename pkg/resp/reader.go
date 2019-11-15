package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Reader struct {
	rd *bufio.Reader
}

type ArrayParser func(r *Reader, num int64) (interface{}, error)

// Convenience function for parsing a slice of strings.
func StringSliceParser(r *Reader, num int64) ([]string, error) {

	s := make([]string, 0, num)

	for i := int64(0); i < num; i++ {
		line, _, err := r.rd.ReadLine()
		if err != nil {
			return nil, err
		}

		if line[0] == SimpleString {
			s = append(s, string(line[1:]))
		} else {
			return nil, errors.New("resp: unexpected type, expected string")
		}
	}

	return s, nil
}

func (r *Reader) ReadMessage(parser ArrayParser) (interface{}, error) {
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

		if length == -1 {
			return nil, nil
		} else if length < -1 {
			return nil, &InvalidArrayLength{length: length}
		}

		return parser(r, length)
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

//func (r *Reader) readArray(length int64) ([]interface{}, error) {
//	s := make([]interface{}, length)
//	for i := int64(0); i < length; i++ {
//		item, err := r.ReadMessage()
//		if err != nil {
//			return nil, err
//		}
//
//		s[i] = item
//	}
//
//	return s, nil
//}

func GenericSliceParser(r *Reader, length int64) (interface{}, error) {
	s := make([]interface{}, length)
	for i := int64(0); i < length; i++ {
		item, err := r.ReadMessage(GenericSliceParser)
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
