package keychain

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func set(keys *Keychain, key []byte, value []byte, t *testing.T) {
	if err := keys.Set(key, value); err != nil {
		t.Fatalf("failed setting value: %v", err)
	}
}

func getAndExpect(keys *Keychain, key []byte, expected []byte, t *testing.T) {
	value, err := keys.Get(key)
	if err != nil {
		t.Fatalf("failed getting value: %v", err)
	}

	if bytes.Compare(expected, value) != 0 {
		t.Fatalf("incorrect value: expected =%s, got =%s", expected, value)
	}
}

func remove(keys *Keychain, key []byte, t *testing.T) {
	_, err := keys.Remove(key)
	if err != nil {
		t.Fatalf("failed removing key: %v", err)
	}
}

func TestAllOperations(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "keychain-test")
	if err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}

	name := f.Name()

	if f.Close() != nil {
		t.Fatalf("could not close temp file: %v", err)
	}

	defer os.Remove(name)

	keys, err := Open(name)
	if err != nil {
		t.Fatalf("could not open database")
	}

	// Test setting keys
	set(keys, []byte("key"), []byte("value"), t)
	set(keys, []byte("key2"), []byte("value2"), t)
	set(keys, []byte("key3"), []byte("value3"), t)
	set(keys, []byte("key4"), []byte(""), t)

	// Test retrieving keys
	getAndExpect(keys, []byte("key"), []byte("value"), t)
	getAndExpect(keys, []byte("key2"), []byte("value2"), t)
	getAndExpect(keys, []byte("key3"), []byte("value3"), t)
	getAndExpect(keys, []byte("key4"), []byte(""), t)

	// Test retrieving non-existent
	getAndExpect(keys, []byte("key5"), nil, t)

	// Test deleting key
	remove(keys, []byte("key"), t)
	getAndExpect(keys, []byte("key"), nil, t)

	// Test overwriting key
	set(keys, []byte("key2"), []byte("value21"), t)
	getAndExpect(keys, []byte("key2"), []byte("value21"), t)

	// Test re-insertion of deleted key
	set(keys, []byte("key"), []byte("valuenew"), t)
	getAndExpect(keys, []byte("key"), []byte("valuenew"), t)

	// Test re-insertion of key with same value
	set(keys, []byte("key3"), []byte("value3"), t)
	getAndExpect(keys, []byte("key3"), []byte("value3"), t)

	// Test close database
	if err := keys.Close(); err != nil {
		t.Fatalf("failed to close database: %v", err)
	}

	// Test reopen database
	keys2, err := Open(name)
	if err != nil {
		t.Fatalf("could not reopen database: %v", err)
	}

	// Test keys are still there
	getAndExpect(keys2, []byte("key"), []byte("valuenew"), t)
	getAndExpect(keys2, []byte("key2"), []byte("value21"), t)
	getAndExpect(keys2, []byte("key3"), []byte("value3"), t)
	getAndExpect(keys2, []byte("key4"), []byte(""), t)

	// Test close database
	if err := keys2.Close(); err != nil {
		t.Fatalf("failed to close database: %v", err)
	}
}
