package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/maybetheresloop/keychain/internal/data"
)

func main() {
	f, err := os.Open("keychain.db")
	f2, err := os.Open("keychain.db")
	if err != nil {
		log.Fatal(err)
	}

	r := data.NewEntryReader(f, 0)

	var entry *data.Entry
	var key []byte
	for key, entry, err = r.ReadEntry(); err == nil; key, entry, err = r.ReadEntry() {
		fmt.Printf("key: %q, value size: %d, value offset: %d\n", string(key), entry.ValueSize, entry.ValuePos)
		value := make([]byte, entry.ValueSize)
		if _, err := f2.ReadAt(value, entry.ValuePos); err != nil {
			log.Fatalf("error reading value: %v", err)
		}

		fmt.Printf("%q", string(value))
	}

	if err != io.EOF {
		log.Fatalf("error reading item: %v", err)
	}
}
