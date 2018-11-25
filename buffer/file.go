package buffer

import (
	"os"
	"path/filepath"

	"github.com/daeMOn63/gorrent/fs"
)

// File interface list methods of a file buffer
type File interface {
	Create(name string, size int64) error
	Open(name string, chunkSize int) (ChunkedFile, error)
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

// Create creates a blank buffer file
func (f *file) Create(name string, size int64) error {
	filename := filepath.Join(f.path, name)
	_, err := f.filesystem.Create(filename)
	if err != nil {
		return err
	}

	return nil
}

func (f *file) Open(name string, chunkSize int) (ChunkedFile, error) {
	filename := filepath.Join(f.path, name)

	file, err := f.filesystem.OpenFile(filename, os.O_APPEND|os.O_RDWR, 0755)
	if err != nil {
		return nil, err
	}

	finfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return &chunkedFile{
		file:      file,
		finfo:     finfo,
		chunkSize: chunkSize,
	}, nil
}

// ChunkedFile allow to manipulate a file by chunk
type ChunkedFile interface {
	Size() int64
	WriteChunk(chunkID int64, data []byte) error
	Read(size int64, offset int64) ([]byte, error)
	Close() error
}

type chunkedFile struct {
	file      fs.File
	finfo     os.FileInfo
	chunkSize int
}

var _ ChunkedFile = &chunkedFile{}

func (b *chunkedFile) Size() int64 {
	return b.finfo.Size()
}

func (b *chunkedFile) WriteChunk(chunkID int64, data []byte) error {
	offset := chunkID * int64(b.chunkSize)
	_, err := b.file.WriteAt(data, offset)
	if err != nil {
		return err
	}

	return nil
}

func (b *chunkedFile) Read(size int64, offset int64) ([]byte, error) {
	buf := make([]byte, size)
	_, err := b.file.ReadAt(buf, offset)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (b *chunkedFile) Close() error {
	return b.file.Close()
}
