package creator

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/gob"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/daeMOn63/gorrent/buffer"
	"github.com/daeMOn63/gorrent/fs"
	"github.com/daeMOn63/gorrent/gorrent"
)

func TestCreator(t *testing.T) {
	expectedPieceLength := 4
	pieceBuffer := buffer.NewMemoryPieceBuffer(expectedPieceLength)

	expectedFiles := []string{"fileA", "child/fileB", "kittens.jpg"}
	expectedSourcePath := "some-dir/source"
	expectedMaxWorkers := 5
	expectedFileContent := []byte("aaaaaaaaaa")
	expectedTargetFile := "target/file.gorrent"

	filesystem := &fs.DummyFS{}

	properOpenFunc := func(path string) (fs.File, error) {
		d := &fs.DummyFile{
			NameVal: path,
			Content: expectedFileContent,
			StatVal: &fs.DummyFileInfo{
				SizeVal: int64(len(expectedFileContent)),
			},
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

	t.Run("Create should handle Stat errors properly", func(t *testing.T) {
		expectedError := errors.New("staterr-string")

		filesystem.FindFilesFunc = properFindFilesFunc
		filesystem.OpenFunc = func(path string) (fs.File, error) {
			d := &fs.DummyFile{
				NameVal: path,
				Content: expectedFileContent,
				StatErr: expectedError,
			}

			return d, nil
		}

		creator := NewCreator(pieceBuffer, filesystem)

		g, err := creator.Create(expectedSourcePath, expectedMaxWorkers)
		if g != nil {
			t.Fatalf("Expected gorrent to be nil")
		}

		if err != expectedError {
			t.Fatalf("Expected err to be %s, got %s", expectedError, err)
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

		g := &gorrent.Gorrent{}

		expectedBytes := bytes.NewBuffer(nil)

		gzwriter := gzip.NewWriter(expectedBytes)
		gob.Register(gorrent.Gorrent{})
		encoder := gob.NewEncoder(gzwriter)
		encoder.Encode(g)
		gzwriter.Close()

		d := &fs.DummyFile{}

		filesystem.CreateFunc = func(path string) (fs.File, error) {
			if path != expectedTargetFile {
				t.Fatalf("Expected path to be %s, got %s", expectedTargetFile, path)
			}

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

		g := &gorrent.Gorrent{}

		creator := NewCreator(pieceBuffer, filesystem)
		err := creator.Save(expectedTargetFile, g)

		if err != expectedErr {
			t.Fatalf("Expected error %s, got %s", expectedErr, err)
		}
	})

	t.Run("Open should properly open gorrent", func(t *testing.T) {

		expectedGorrent := &gorrent.Gorrent{}

		expectedBytes := bytes.NewBuffer(nil)

		gzwriter := gzip.NewWriter(expectedBytes)
		gob.Register(gorrent.Gorrent{})
		encoder := gob.NewEncoder(gzwriter)
		encoder.Encode(expectedGorrent)
		gzwriter.Close()

		d := &fs.DummyFile{}

		filesystem.OpenFunc = func(path string) (fs.File, error) {
			if path != expectedTargetFile {
				t.Fatalf("Expected path to be %s, got %s", expectedTargetFile, path)
			}

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
