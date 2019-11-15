package resp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadString(t *testing.T) {
	input := "+hello, world\r\n"
	expected := "hello, world"
	r := NewReader(strings.NewReader(input))

	result, err := r.ReadMessage(nil)
	if err != nil {
		t.Fatalf("error reading message - %v", err)
	}

	got, ok := result.(string)
	if !ok {
		t.Fatalf("incorrect type - expected=string")
	}

	if got != expected {
		t.Fatalf("incorrect string - expected=%q, got=%q", expected, result)
	}
}

func TestReadStringSlice(t *testing.T) {
	input := "*3\r\n+foo\r\n+bar\r\n+baz\r\n"
	expected := []string{"foo", "bar", "baz"}

	r := NewReader(strings.NewReader(input))
	res, err := r.ReadMessage(StringSliceParser)

	assert.Nil(t, err)

	assert.Equal(t, expected, res)
}

func TestReadError(t *testing.T) {
	input := "-Error message\r\n"
	expected := NewRespError("Error message")
	r := NewReader(strings.NewReader(input))

	result, err := r.ReadMessage(nil)
	if err != nil {
		t.Fatalf("error reading message - %v", err)
	}

	got, ok := result.(RespError)
	if !ok {
		t.Fatalf("incorrect type - expected=RespError")
	}

	if got != expected {
		t.Fatalf("incorrect string - expected=%q, got=%q", expected.message, got.message)
	}
}

func TestReadInteger(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{input: ":1234\r\n", expected: 1234},
		{input: ":-1234\r\n", expected: -1234},
	}

	for i, tt := range tests {
		r := NewReader(strings.NewReader(tt.input))

		result, err := r.ReadMessage(nil)
		if err != nil {
			t.Fatalf("tests[%d] - error reading message, %v", i, err)
		}

		got, ok := result.(int64)
		if !ok {
			t.Fatalf("tests[%d] - incorrect type, expected=%q", i, "int64")
		}

		if got != tt.expected {
			t.Fatalf("tests[%d]: incorrect result, expected=%d, got=%d", i, tt.expected, got)
		}
	}
}

func TestReadBulkString(t *testing.T) {
	tests := []struct {
		input    string
		expected []byte
	}{
		{"$6\r\nfoobar\r\n",
			[]byte("foobar")},
		{"$8\r\nfoobar\r\n\r\n",
			[]byte("foobar\r\n")},
	}

	for i, tt := range tests {
		r := NewReader(strings.NewReader(tt.input))

		result, err := r.ReadMessage(nil)
		if err != nil {
			t.Fatalf("tests[%d] - error reading message, %v", i, err)
		}

		got, ok := result.([]byte)
		if !ok {
			t.Fatalf("tests[%d] - incorrect type, expected=%q", i, "string")
		}

		if bytes.Compare(got, tt.expected) != 0 {
			t.Fatalf("tests[%d]: incorrect result, expected=%q, got=%q", i, tt.expected, got)
		}
	}
}

func compareRespTypes(expected interface{}, result interface{}, i int, t *testing.T) {
	switch ev := expected.(type) {
	case string:
		rv, ok := result.(string)
		if !ok {
			t.Fatalf("tests[%d] - mismatched types, expected=%q", i, "string")
		}

		if ev != rv {
			t.Fatalf("tests[%d] - incorrect result, expected=%q, got=%q", i, ev, rv)
		}
	case []byte:
		rv, ok := result.([]byte)
		if !ok {
			t.Fatalf("tests[%d] - mismatched types, expected=%q", i, "[]byte")
		}

		if bytes.Compare(ev, rv) != 0 {
			t.Fatalf("tests[%d] - incorrect result, expected=%q, got=%q", i, ev, rv)
		}
	}
}

func TestReadArray(t *testing.T) {
	tests := []struct {
		input    string
		expected []interface{}
	}{
		{"*2\r\n+Hello\r\n$5\r\nworld\r\n", []interface{}{"Hello", []byte("world")}},
		{"*0\r\n", []interface{}{}},
		{"*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", []interface{}{[]byte("foo"), []byte("bar")}},
		{"*3\r\n:1\r\n:2\r\n:3\r\n", []interface{}{1, 2, 3}},
		{"*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$6\r\nfoobar\r\n", []interface{}{1, 2, 3, 4, []byte("foobar")}},
	}

	for i, tt := range tests {
		r := NewReader(strings.NewReader(tt.input))

		result, err := r.ReadMessage(GenericSliceParser)
		if err != nil {
			t.Fatalf("tests[%d] - error reading message, %v", i, err)
		}

		s, ok := result.([]interface{})
		if !ok {
			t.Fatalf("tests[%d] - incorrect type, expected=%q", i, "interface{}")
		}

		if len(s) != len(tt.expected) {
			t.Fatalf("tests[%d] - incorrect length, expected=%d, got=%d", i, len(tt.expected), len(s))
		}

		for j := 0; j < len(s); j++ {
			compareRespTypes(s[j], tt.expected[j], i, t)
		}
	}
}
