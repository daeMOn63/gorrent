package buffer

import (
	"path/filepath"

	"github.com/daeMOn63/gorrent/fs"
)

// File interface list methods of a file buffer
type File interface {
	Create(name string, size int64) error
	Write(name string, data []byte, offset int64) error
}

type file struct {
	path       string
	filesystem fs.FileSystem
}

var _ File = &file{}

// NewFile creates a new file buffer
func NewFile(filesystem fs.FileSystem, path string) File {
	return &file{
		path:       path,
		filesystem: filesystem,
	}

}

// Create creates a blank buffer file of given size
func (f *file) Create(name string, size int64) error {
	filename := filepath.Join(f.path, name)
	_, err := f.filesystem.Create(filename)
	if err != nil {
		return err
	}

	return f.filesystem.Truncate(filename, size)
}

func (f *file) Write(name string, data []byte, offset int64) error {

	// TODO
	return nil
}
