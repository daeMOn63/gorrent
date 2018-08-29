package gorrent

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/gob"
	"errors"
	"gorrent/fs"
	"io"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type DummyFS struct {
	FindFilesFunc func(string, int) ([]string, error)
	OpenFunc      func(string) (fs.File, error)
	CreateFunc    func(string) (fs.File, error)
}

func (filesystem *DummyFS) FindFiles(rootPath string, maxWorkers int) (filepaths []string, err error) {
	return filesystem.FindFilesFunc(rootPath, maxWorkers)
}
func (filesystem *DummyFS) Open(path string) (fs.File, error) {
	return filesystem.OpenFunc(path)
}

func (filesystem *DummyFS) Create(path string) (fs.File, error) {
	return filesystem.CreateFunc(path)
}

type DummyFile struct {
	PathVal    string
	Content    []byte
	CurReadPtr int
	ReadFunc   func([]byte) (int, error)
	WriteFunc  func([]byte) (int, error)
	CloseErr   error
}

func (f *DummyFile) Path() string {
	return f.PathVal
}

func (f *DummyFile) Size() int64 {
	return int64(len(f.Content))
}

func (f *DummyFile) Read(p []byte) (int, error) {
	if f.CurReadPtr >= len(f.Content) {
		return 0, io.EOF
	}

	maxRead := len(p)
	if maxRead > len(f.Content[f.CurReadPtr:]) {
		maxRead = len(f.Content[f.CurReadPtr:])
	}

	len := maxRead - f.CurReadPtr
	copy(p, f.Content[f.CurReadPtr:maxRead])
	f.CurReadPtr += len
	return len, nil
}

func (f *DummyFile) Write(p []byte) (int, error) {
	f.Content = append(f.Content, p...)
	return len(p), nil
}

func (f *DummyFile) Close() error {
	return f.CloseErr
}

func TestCreator(t *testing.T) {
	expectedPieceLength := 4
	pieceBuffer := NewMemoryPieceBuffer(expectedPieceLength)

	expectedFiles := []string{"fileA", "child/fileB", "kittens.jpg"}
	expectedSourcePath := "some-dir/source"
	expectedMaxWorkers := 5
	expectedFileContent := []byte("aaaaaaaaaa")
	expectedTargetFile := "target/file.gorrent"

	filesystem := &DummyFS{}

	properOpenFunc := func(path string) (fs.File, error) {
		d := &DummyFile{
			PathVal: path,
			Content: expectedFileContent,
		}

		return d, nil
	}

	properFindFilesFunc := func(path string, maxWorkers int) ([]string, error) {
		if path != expectedSourcePath {
			t.Fatalf("Expected path to be %s, got %s", expectedSourcePath, path)
		}
		if maxWorkers != expectedMaxWorkers {
			t.Fatalf("Expected maxWorkers to be %d, got %d", expectedMaxWorkers, maxWorkers)
		}
		return expectedFiles, nil
	}

	t.Run("Create returns a valid gorrent", func(t *testing.T) {
		filesystem.FindFilesFunc = properFindFilesFunc
		filesystem.OpenFunc = properOpenFunc

		creator := NewCreator(pieceBuffer, filesystem)

		g, err := creator.Create(expectedSourcePath, expectedMaxWorkers)
		if err != nil {
			t.Fatal(err)
		}

		if len(g.Files) != len(expectedFiles) {
			t.Fatalf("Expected %d files, got %d", len(expectedFiles), len(g.Files))
		}

		for i, f := range g.Files {
			expectedFile := filepath.Join(expectedSourcePath, expectedFiles[i])
			if f.Name != expectedFile {
				t.Fatalf("Expected filename to be %s, got %s", expectedFile, f.Name)
			}

			expectedHash := sha1.Sum(expectedFileContent)
			if f.Hash != expectedHash {
				t.Fatalf("Expected file hash to be %v, got %v", expectedHash, f.Hash)
			}

			if int(f.Length) != len(expectedFileContent) {
				t.Fatalf("Expected length to be %v, got %v", len(expectedFileContent), f.Length)
			}
		}

		zeroTime := time.Time{}
		if g.CreationDate == zeroTime {
			t.Fatal("Expected creationDate to not be a time zero value")
		}

		if g.PieceLength != expectedPieceLength {
			t.Fatalf("Expected pieceLength to be %v, got %v", expectedPieceLength, g.PieceLength)
		}

		if len(g.Pieces) != 8 {
			t.Fatalf("Expected 8 pieces, got %d", len(g.Pieces))
		}
	})

	t.Run("Create should handle FindFiles errors properly", func(t *testing.T) {
		expectedErr := errors.New("findfiles errstring")
		filesystem.FindFilesFunc = func(path string, maxWorkers int) ([]string, error) {
			return nil, expectedErr
		}

		creator := NewCreator(pieceBuffer, filesystem)
		g, err := creator.Create(expectedSourcePath, expectedMaxWorkers)
		if g != nil {
			t.Fatalf("Expected gorrent to be nil")
		}

		if err != expectedErr {
			t.Fatalf("Expected error to be %s, got %s", expectedErr, err)
		}
	})

	t.Run("Create should handle Open errors properly", func(t *testing.T) {
		expectedErr := errors.New("open errstring")
		filesystem.FindFilesFunc = properFindFilesFunc
		filesystem.OpenFunc = func(path string) (fs.File, error) {
			return nil, expectedErr
		}

		creator := NewCreator(pieceBuffer, filesystem)
		g, err := creator.Create(expectedSourcePath, expectedMaxWorkers)
		if g != nil {
			t.Fatalf("Expected gorrent to be nil")
		}

		if err != expectedErr {
			t.Fatalf("Expected error to be %v, got %v", expectedErr, err)
		}
	})

	t.Run("Save should properly save gorrent", func(t *testing.T) {

		g := &Gorrent{}

		expectedBytes := bytes.NewBuffer(nil)

		gzwriter := gzip.NewWriter(expectedBytes)
		gob.Register(Gorrent{})
		encoder := gob.NewEncoder(gzwriter)
		encoder.Encode(g)
		gzwriter.Close()

		d := &DummyFile{}

		filesystem.CreateFunc = func(path string) (fs.File, error) {
			if path != expectedTargetFile {
				t.Fatalf("Expected path to be %s, got %s", expectedTargetFile, path)
			}

			d.PathVal = path

			return d, nil
		}

		creator := NewCreator(pieceBuffer, filesystem)

		err := creator.Save(expectedTargetFile, g)
		if err != nil {
			t.Fatalf("Expected no errors, got %s", err)
		}

		if reflect.DeepEqual(d.Content, expectedBytes.Bytes()) == false {
			t.Fatalf("Expected file content to be %v, got %v", expectedBytes.Bytes(), d.Content)
		}
	})

	t.Run("Save should handle filesystem errors", func(t *testing.T) {
		expectedErr := errors.New("create errstring")
		filesystem.CreateFunc = func(path string) (fs.File, error) {
			return nil, expectedErr
		}

		g := &Gorrent{}

		creator := NewCreator(pieceBuffer, filesystem)
		err := creator.Save(expectedTargetFile, g)

		if err != expectedErr {
			t.Fatalf("Expected error %s, got %s", expectedErr, err)
		}
	})

	t.Run("Open should properly open gorrent", func(t *testing.T) {

		expectedGorrent := &Gorrent{}

		expectedBytes := bytes.NewBuffer(nil)

		gzwriter := gzip.NewWriter(expectedBytes)
		gob.Register(Gorrent{})
		encoder := gob.NewEncoder(gzwriter)
		encoder.Encode(expectedGorrent)
		gzwriter.Close()

		d := &DummyFile{}

		filesystem.OpenFunc = func(path string) (fs.File, error) {
			if path != expectedTargetFile {
				t.Fatalf("Expected path to be %s, got %s", expectedTargetFile, path)
			}

			d.PathVal = path
			d.Content = expectedBytes.Bytes()

			return d, nil
		}

		creator := NewCreator(pieceBuffer, filesystem)

		g, err := creator.Open(expectedTargetFile)
		if err != nil {
			t.Fatalf("Expected no errors, got %s", err)
		}

		if reflect.DeepEqual(g, expectedGorrent) == false {
			t.Fatalf("Expected gorrent to be %v, got %v", expectedGorrent, g)
		}
	})

	t.Run("Open should properly handle filesystem errors", func(t *testing.T) {
		expectedErr := errors.New("open errstring")
		filesystem.OpenFunc = func(path string) (fs.File, error) {
			return nil, expectedErr
		}

		creator := NewCreator(pieceBuffer, filesystem)
		_, err := creator.Open(expectedTargetFile)

		if err != expectedErr {
			t.Fatalf("Expected error %s, got %s", expectedErr, err)
		}
	})

}
