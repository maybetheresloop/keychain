package resp

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadString(t *testing.T) {
	input := "+hello, world\r\n"
	expected := "hello, world"
	r := NewReader(strings.NewReader(input))

	result, err := r.ReadMessage()
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

func TestReadError(t *testing.T) {
	input := "-Error message\r\n"
	expected := NewRespError("Error message")
	r := NewReader(strings.NewReader(input))

	result, err := r.ReadMessage()
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

		result, err := r.ReadMessage()
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

		result, err := r.ReadMessage()
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
