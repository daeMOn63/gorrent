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
}

// diskFS is a FileSystem reading from disk
type diskFS struct {
	tokens chan struct{}
}

var _ FileSystem = &diskFS{}

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

// Open returns a File from given path
func (dfs *diskFS) Open(path string) (File, error) {
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
func (dfs *diskFS) Create(path string) (File, error) {
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