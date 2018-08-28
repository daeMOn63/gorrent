package gorrent

import (
	"crypto/sha1"
	"gorrent/fs"
	"io"
	"path/filepath"
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
	return f.ReadFunc(p)
}

func (f *DummyFile) Write(p []byte) (int, error) {
	return f.WriteFunc(p)
}

func (f *DummyFile) Close() error {
	return f.CloseErr
}

func TestCreator(t *testing.T) {
	expectedPieceLength := 4
	pieceBuffer := NewMemoryPieceBuffer(expectedPieceLength)

	expectedFiles := []string{"fileA", "child/fileB", "kittens.jpg"}
	expectedPath := "some-dir"
	expectedMaxWorkers := 5
	expectedFileContent := []byte("aaaaaaaaaa")

	filesystem := &DummyFS{}

	filesystem.OpenFunc = func(path string) (fs.File, error) {
		d := &DummyFile{
			PathVal: path,
			Content: expectedFileContent,
		}

		d.ReadFunc = func(buf []byte) (int, error) {
			if d.CurReadPtr >= len(d.Content) {
				return 0, io.EOF
			}

			maxRead := len(buf)
			if maxRead > len(d.Content[d.CurReadPtr:]) {
				maxRead = len(d.Content[d.CurReadPtr:])
			}

			len := maxRead - d.CurReadPtr
			copy(buf, d.Content[d.CurReadPtr:maxRead])
			d.CurReadPtr += len
			return len, nil
		}

		d.WriteFunc = func(buf []byte) (int, error) {
			d.Content = append(d.Content, buf...)
			return len(buf), nil
		}

		return d, nil
	}

	filesystem.FindFilesFunc = func(path string, maxWorkers int) ([]string, error) {
		if path != expectedPath {
			t.Fatalf("Expected path to be %s, got %s", expectedPath, path)
		}
		if maxWorkers != expectedMaxWorkers {
			t.Fatalf("Expected maxWorkers to be %d, got %d", expectedMaxWorkers, maxWorkers)
		}
		return expectedFiles, nil
	}

	t.Run("Create returns a valid gorrent", func(t *testing.T) {
		creator := NewCreator(pieceBuffer, filesystem)

		g, err := creator.Create(expectedPath, expectedMaxWorkers)
		if err != nil {
			t.Fatal(err)
		}

		if len(g.Files) != len(expectedFiles) {
			t.Fatalf("Expected %d files, got %d", len(expectedFiles), len(g.Files))
		}

		for i, f := range g.Files {
			expectedFile := filepath.Join(expectedPath, expectedFiles[i])
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

}
