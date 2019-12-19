package data

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadWriteHints(t *testing.T) {
	var inner bytes.Buffer
	wr := NewHintWriter(&inner)

	tt := []struct {
		key   []byte
		entry *Entry
	}{
		{key: []byte("key1"), entry: &Entry{Timestamp: 1, FileID: 0, ValueSize: 0, ValuePos: 4}},
		{key: []byte("key2"), entry: &Entry{Timestamp: 2, FileID: 0, ValueSize: 4, ValuePos: 8}},
		{key: []byte("key3"), entry: &Entry{Timestamp: 3, FileID: 0, ValueSize: 8, ValuePos: 12}},
		{key: []byte("key4"), entry: &Entry{Timestamp: 4, FileID: 0, ValueSize: 12, ValuePos: 16}},
	}

	for _, test := range tt {
		assert.Nil(t, wr.WriteFromEntry(test.key, test.entry))
	}

	rd := NewHintFileEntryReader(&inner, 0)
	for _, test := range tt {
		key, entry, err := rd.ReadEntry()
		assert.Nil(t, err)
		assert.Equal(t, test.key, key)
		assert.Equal(t, *test.entry, *entry)
	}
}
