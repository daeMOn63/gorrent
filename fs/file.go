package fs

import (
	"os"
)

// File interface describe methods available on a file
type File interface {
	Path() string
	Size() int64
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
}

type fsFile struct {
	path string
	osf  *os.File
	osfi os.FileInfo
}

var _ File = fsFile{}

func (f fsFile) Path() string {
	return f.path
}

func (f fsFile) Size() int64 {
	return f.osfi.Size()
}

func (f fsFile) Read(p []byte) (int, error) {
	return f.osf.Read(p)
}

func (f fsFile) Write(p []byte) (int, error) {
	return f.osf.Write(p)
}

func (f fsFile) Close() error {
	return f.osf.Close()
}
