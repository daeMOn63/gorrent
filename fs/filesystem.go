package fs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// FileSystem allow interactions with the file system
type FileSystem interface {
	FindFiles(rootPath string, maxWorkers int) (filepaths []string, err error)
	Open(path string) (File, error)
	Create(path string) (File, error)
	Stat(path string) (os.FileInfo, error)
	MkdirAll(path string, mode os.FileMode) error
	Remove(path string) error
}

// diskFS is a FileSystem reading from disk
type diskFS struct {
	tokens chan struct{}
}

var _ FileSystem = &diskFS{}
var _ FileSystem = &DummyFS{}

// NewFileSystem create a new diskFS, spawning at max numWorkers goroutine
func NewFileSystem() FileSystem {
	return &diskFS{}
}

// FindFiles return all file paths recursively under given path
// or path if it is not a directory. maxWorkers can be used to control the
// maximum number of goroutine, to avoid too many open files errors
func (dfs *diskFS) FindFiles(path string, maxWorkers int) ([]string, error) {
	if maxWorkers <= 0 {
		return nil, errors.New("maxWorkers must be greater than 0")
	}

	ch := make(chan getFileOutput)

	tokens := make(chan struct{}, maxWorkers)

	go dfs.getFiles(path, tokens, ch)
	var files []string

	childs := 1
	for childs > 0 {
		select {
		case f := <-ch:
			if f.err != nil {
				return nil, f.err
			}
			// Add all childs and remove current file
			childs += f.childs - 1
			if f.file != "" {
				files = append(files, strings.Replace(f.file, path, "", 1))
			}
		}
	}

	return files, nil
}

// Open returns a File from given path by wrapping around os.Open
func (dfs *diskFS) Open(path string) (File, error) {
	return os.Open(path)
}

// Create create a new file by wrapping around os.Create
func (dfs *diskFS) Create(path string) (File, error) {
	return os.Create(path)
}

// Stat wrap os.Stat
func (dfs *diskFS) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// MkdirAll wrap os.MkdirAll
func (dfs *diskFS) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

// Remove wrap os.Remove
func (dfs *diskFS) Remove(path string) error {
	return os.Remove(path)
}

type getFileOutput struct {
	file   string
	err    error
	childs int
}

func (dfs *diskFS) getFiles(path string, tokens chan struct{}, ch chan getFileOutput) {
	tokens <- struct{}{}
	defer func() { <-tokens }()

	root, err := os.Open(path)
	if err != nil {
		ch <- getFileOutput{err: err}
		return
	}
	defer root.Close()

	info, err := root.Stat()
	if err != nil {
		ch <- getFileOutput{err: err}
		return
	}

	if info.IsDir() == false {
		ch <- getFileOutput{file: path}
		return
	}

	finfos, err := root.Readdir(-1)
	if err != nil {
		ch <- getFileOutput{file: path}
		return
	}

	ch <- getFileOutput{childs: len(finfos)}

	for _, finfo := range finfos {
		p := filepath.Join(path, finfo.Name())
		go dfs.getFiles(p, tokens, ch)
	}
}

// DummyFS implements FileSystem but allow to configure its behavior
type DummyFS struct {
	FindFilesFunc func(string, int) ([]string, error)
	OpenFunc      func(string) (File, error)
	CreateFunc    func(string) (File, error)
	StatFunc      func(path string) (os.FileInfo, error)
	MkdirAllFunc  func(path string, mode os.FileMode) error
	RemoveFunc    func(path string) error
}

// FindFiles calls FindFilesFunc
func (filesystem *DummyFS) FindFiles(rootPath string, maxWorkers int) (filepaths []string, err error) {
	return filesystem.FindFilesFunc(rootPath, maxWorkers)
}

// Open calls OpenFunc
func (filesystem *DummyFS) Open(path string) (File, error) {
	return filesystem.OpenFunc(path)
}

// Create calls CreateFunc
func (filesystem *DummyFS) Create(path string) (File, error) {
	return filesystem.CreateFunc(path)
}

// Stat calls StatFunc
func (filesystem *DummyFS) Stat(path string) (os.FileInfo, error) {
	return filesystem.StatFunc(path)
}

// MkdirAll calls MkdirAllFunc
func (filesystem *DummyFS) MkdirAll(path string, mode os.FileMode) error {
	return filesystem.MkdirAllFunc(path, mode)
}

func (filesystem *DummyFS) Remove(path string) error {
	return filesystem.RemoveFunc(path)
}
