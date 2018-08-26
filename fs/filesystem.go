package fs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// FileSystem allow interactions with the file system
type FileSystem interface {
	FindFiles(string) ([]string, error)
	Open(string) (File, error)
	Create(string) (File, error)
}

// DiskFS is a FileSystem reading from disk
type DiskFS struct {
	tokens chan struct{}
}

var _ FileSystem = &DiskFS{}

// NewDiskFS create a new DiskFS, spawning at max numWorkers goroutine
func NewDiskFS(numWorkers int) (*DiskFS, error) {
	if numWorkers <= 0 {
		return nil, errors.New("numWorkers must be greater than 0")
	}

	return &DiskFS{
		tokens: make(chan struct{}, numWorkers),
	}, nil
}

// FindFiles return all file paths recursively under given path
// or path if it is not a directory
func (dfs *DiskFS) FindFiles(path string) ([]string, error) {
	ch := make(chan getFileOutput)

	go dfs.getFiles(path, ch)
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

// Open returns a File from given path
func (dfs *DiskFS) Open(path string) (File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	finfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return &fsFile{
		path: path,
		osf:  file,
		osfi: finfo,
	}, nil
}

// Create create a new file
func (dfs *DiskFS) Create(path string) (File, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return &fsFile{
		path: path,
		osf:  f,
		osfi: info,
	}, nil
}

type getFileOutput struct {
	file   string
	err    error
	childs int
}

func (dfs *DiskFS) getFiles(path string, ch chan getFileOutput) {
	dfs.tokens <- struct{}{}
	defer func() { <-dfs.tokens }()

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
		go dfs.getFiles(p, ch)
	}
}
