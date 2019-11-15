package resp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriter_WriteSimpleString(t *testing.T) {
	var builder strings.Builder
	wr := NewWriter(&builder)

	err := wr.WriteSimpleString("foo")
	assert.Nil(t, err)

	err = wr.Flush()
	assert.Nil(t, err)

	assert.Equal(t, "+foo\r\n", builder.String())
}

func TestWriter_WriteError(t *testing.T) {
	builder := new(strings.Builder)
	wr := NewWriter(builder)

	err := wr.WriteError(NewRespError("error"))
	assert.Nil(t, err)

	err = wr.Flush()
	assert.Nil(t, err)

	assert.Equal(t, "-error\r\n", builder.String())
}

func TestWriter_WriteBulkString(t *testing.T) {
	buf := new(bytes.Buffer)
	wr := NewWriter(buf)

	err := wr.WriteBulkString([]byte("\x01\x02\r\n"))
	assert.Nil(t, err)

	err = wr.Flush()
	assert.Nil(t, err)

	assert.Equal(t, []byte("$4\r\n\x01\x02\r\n\r\n"), buf.Bytes())
}

func TestWriter_WriteBulkStringNil(t *testing.T) {
	buf := new(bytes.Buffer)
	wr := NewWriter(buf)

	err := wr.WriteBulkString(nil)
	assert.Nil(t, err)

	err = wr.Flush()
	assert.Nil(t, err)

	assert.Equal(t, []byte("$-1\r\n"), buf.Bytes())
}

func TestWriter_WriteArray(t *testing.T) {
	buf := new(bytes.Buffer)
	wr := NewWriter(buf)

	err := wr.WriteArray([]interface{}{"simplestring", NewRespError("error"), 1, []byte("bulkstring")})
	assert.Nil(t, err)

	err = wr.Flush()
	assert.Nil(t, err)

	assert.Equal(t, []byte("*4\r\n+simplestring\r\n-error\r\n:1\r\n$10\r\nbulkstring\r\n"), buf.Bytes())
}
