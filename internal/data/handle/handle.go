package handle

import (
	"errors"
	"io"
	"log"
	"os"

	"golang.org/x/exp/mmap"
)

var (
	InvalidOperation = errors.New("invalid operation")
)

// Represents a handle to a database file. The handle can be backed by
// either an *os.File (used for the active file) or a *mmap.ReaderAt (used for
// read-only files).
type Handle struct {
	rd       *os.File       // read handle to the file (active handle only).
	wr       *os.File       // write handle to the file (active handle only).
	rda      *mmap.ReaderAt // read handle to the file (read-only handle only).
	offset   int64          // current offset of the file
	len      int            // number of keys in the file
	size     int64          // size of the underlying file, in bytes (active handle only).
	deadKeys int            // number of dead keys in the file
}

func openActive(filename string) (*Handle, error) {
	wr, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	rd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return &Handle{
		rd: rd,
		wr: wr,
	}, nil
}

func openReadOnly(filename string) (*Handle, error) {
	rda, err := mmap.Open(filename)
	if err != nil {
		return nil, err
	}

	return &Handle{rda: rda, size: -1}, nil
}

func Open(filename string, readOnly bool) (*Handle, error) {
	if readOnly {
		return openReadOnly(filename)
	}
	return openActive(filename)
}

func (h *Handle) Read(p []byte) (n int, err error) {
	if h.rda != nil {
		if h.rda.Len() == 0 {
			return 0, io.EOF
		}

		n, err = h.rda.ReadAt(p, h.offset)
		h.offset += int64(n)
		return
	}
	return h.rd.Read(p)
}

func (h *Handle) ReadAt(p []byte, off int64) (int, error) {
	if h.rda != nil {
		return h.rda.ReadAt(p, off)
	}

	log.Println("readat file")
	return h.rd.ReadAt(p, off)
}

// Closes the handle.
func (h *Handle) Close() error {
	if h.rda != nil {
		return h.rda.Close()
	}

	if err := h.rd.Close(); err != nil {
		return err
	}

	return h.wr.Close()
}

func (h *Handle) Write(b []byte) (n int, err error) {
	if h.wr == nil {
		return 0, InvalidOperation
	}

	return h.wr.Write(b)
}

func (h *Handle) Sync() error {
	if h.wr == nil {
		return InvalidOperation
	}

	return h.wr.Sync()
}

// Increments the handle's deadKey counter.
func (h *Handle) DeadKey() {
	h.deadKeys += 1
}
