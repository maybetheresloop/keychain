package handle

import (
	"os"

	"golang.org/x/exp/mmap"
)

// Represents a handle to a database file. The handle can be backed by
// either an *os.File (used for the active file) or a *mmap.ReaderAt (used for
// read-only files).
type Handle struct {
	// Read handle to the database file (active handle only).
	rd *os.File

	// Write handle to the database file (active handle only).
	wr *os.File

	// Read handle to the database file (read-only handle only).
	rda *mmap.ReaderAt
}

func openReadOnly(filename string) (*Handle, error) {
	return nil, nil
}

func openReadWrite(filename string) (*Handle, error) {
	rda, err := mmap.Open(filename)
	if err != nil {
		return nil, err
	}

	return &Handle{rda: rda}, nil
}

func Open(filename string, readOnly bool) (*Handle, error) {
	if readOnly {
		return openReadOnly(filename)
	}
	return openReadWrite(filename)
}

func (h *Handle) ReadAt(p []byte, off int64) (int, error) {
	if h.rda != nil {
		return h.rda.ReadAt(p, off)
	}
	return h.rd.ReadAt(p, off)
}

// Closes the handle.
func (h *Handle) Close() error {
	if h.rda != nil {
		return h.Close()
	}

	if err := h.rd.Close(); err != nil {
		return err
	}

	return h.wr.Close()
}
