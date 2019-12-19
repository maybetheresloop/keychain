package handle

import (
	"errors"
	"io"
	"os"

	"golang.org/x/exp/mmap"
)

var (
	InvalidOperation = errors.New("invalid operation")
)

type activeHandle struct {
	isInit   bool
	filename string
	rd       *os.File
	wr       *os.File
}

func openActiveHandle(filename string) *activeHandle {
	return &activeHandle{
		filename: filename,
	}
}

func (h *activeHandle) init() error {
	if h.isInit {
		return nil
	}

	var err error
	if h.wr, err = os.OpenFile(h.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644); err != nil {
		return err
	}

	if h.rd, err = os.Open(h.filename); err != nil {
		return err
	}

	h.isInit = true
	return nil
}

func (h *activeHandle) Read(p []byte) (int, error) {
	if !h.isInit {
		if err := h.init(); err != nil {
			return 0, err
		}
	}

	return h.rd.Read(p)
}

func (h *activeHandle) ReadAt(p []byte, off int64) (int, error) {
	if !h.isInit {
		if err := h.init(); err != nil {
			return 0, err
		}
	}

	return h.rd.ReadAt(p, off)
}

func (h *activeHandle) Write(b []byte) (n int, err error) {
	if !h.isInit {
		if err := h.init(); err != nil {
			return 0, nil
		}
	}

	return h.wr.Write(b)
}

func (h *activeHandle) Close() error {
	if !h.isInit {
		return nil
	}

	if err := h.rd.Close(); err != nil {
		return err
	}

	err := h.wr.Close()
	return err
}

func (h *activeHandle) Sync() error {
	if !h.isInit {
		return nil
	}

	return h.wr.Sync()
}

// Represents a handle to a database file. The handle can be backed by
// either an *os.File (used for the active file) or a *mmap.ReaderAt (used for
// read-only files).
type Handle struct {
	active   *activeHandle
	rda      *mmap.ReaderAt // read handle to the file (read-only handle only).
	offset   int64          // current offset of the file
	len      int            // number of keys in the file
	size     int64          // size of the underlying file, in bytes (active handle only).
	deadKeys int            // number of dead keys in the file
}

func openActive(filename string) (*Handle, error) {
	return &Handle{
		active: openActiveHandle(filename),
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
	return h.active.Read(p)
}

func (h *Handle) ReadAt(p []byte, off int64) (int, error) {
	if h.rda != nil {
		return h.rda.ReadAt(p, off)
	}

	return h.active.ReadAt(p, off)
}

// Closes the handle.
func (h *Handle) Close() error {
	if h.rda != nil {
		return h.rda.Close()
	}

	return h.active.Close()
}

func (h *Handle) Write(b []byte) (n int, err error) {
	if h.active == nil {
		return 0, InvalidOperation
	}

	return h.active.Write(b)
}

func (h *Handle) Sync() error {
	if h.active == nil {
		return InvalidOperation
	}

	return h.active.Sync()
}

// Increments the handle's deadKey counter.
func (h *Handle) DeadKey() {
	h.deadKeys += 1
}
