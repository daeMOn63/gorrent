package buffer

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/daeMOn63/gorrent/fs"

	"github.com/daeMOn63/gorrent/gorrent"
)

var errBrokenReader = errors.New("readerror :)")

type brokenReaderT struct{}

func (b *brokenReaderT) Read(p []byte) (int, error) {
	return 0, errBrokenReader
}

func TestMemoryPieceBuffer(t *testing.T) {

	t.Run("NewMemoryPieceBuffer should create a new buffer", func(t *testing.T) {
		var lengths = []int{1234, 0, 666}
		for _, length := range lengths {
			pb := NewMemoryPieceBuffer(length)
			if pb.pieceLength != length {
				t.Fatalf("Expected length '%d', got '%d'", length, pb.pieceLength)
			}
		}
	})

	t.Run("Length() should return proper length", func(t *testing.T) {
		var lengths = []int{1234, 0, 666}
		for _, length := range lengths {
			pb := NewMemoryPieceBuffer(length)
			if pb.PieceLength() != length {
				t.Fatalf("Expected length '%d', got '%d'", length, pb.pieceLength)
			}
		}
	})

	t.Run("CreatePieces should properly create pieces", func(t *testing.T) {
		pb := NewMemoryPieceBuffer(10)
		data := []byte("abcdefghijlkmnopqrstuvwxyz")

		expectedPieces := []gorrent.Sha1Hash{
			sha1.Sum([]byte("abcdefghij")),
			sha1.Sum([]byte("lkmnopqrst")),
		}

		expectedRemainingBuffer := []byte("uvwxyz")

		pieces, err := pb.CreatePieces(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		if reflect.DeepEqual(pieces, expectedPieces) == false {
			t.Fatalf("Expected piece hashes %v, got %v", expectedPieces, pieces)
		}

		if bytes.Equal(pb.buf.Bytes(), expectedRemainingBuffer) == false {
			t.Fatalf("Expected remaining buffer '%s', got '%s'", expectedRemainingBuffer, pb.buf.Bytes())
		}

		expectedPieces = []gorrent.Sha1Hash{
			sha1.Sum([]byte("uvwxyz1234")),
		}
		expectedRemainingBuffer = []byte("56789")

		pieces, err = pb.CreatePieces(strings.NewReader("123456789"))
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		if reflect.DeepEqual(pieces, expectedPieces) == false {
			t.Fatalf("Expected piece hashes %v, got %v", expectedPieces, pieces)
		}

		if bytes.Equal(pb.buf.Bytes(), expectedRemainingBuffer) == false {
			t.Fatalf("Expected remaining buffer '%s', got '%s'", expectedRemainingBuffer, pb.buf.Bytes())
		}
	})

	t.Run("Flush should return last piece and clear itself", func(t *testing.T) {
		pb := NewMemoryPieceBuffer(3)

		pb.CreatePieces(bytes.NewReader([]byte("12345")))
		lastPiece := pb.Flush()

		expectedHash := gorrent.Sha1Hash(sha1.Sum([]byte("45\x00")))
		if lastPiece != expectedHash {
			t.Fatalf("Expected last piece hash to be %v, got %v", expectedHash, lastPiece)
		}

		if pb.buf.Len() != 0 {
			t.Fatalf("Expected buf lenght to be %d, got %d", 0, pb.buf.Len())
		}

		lastPiece = pb.Flush()
		expectedHash = gorrent.Sha1Hash(sha1.Sum([]byte("\x00\x00\x00")))
		if lastPiece != expectedHash {
			t.Fatalf("Expected hash %v, got %v", expectedHash, lastPiece)
		}
	})

	t.Run("Empty should indicate when the internal buffer is empty", func(t *testing.T) {
		pb := NewMemoryPieceBuffer(2)

		if pb.Empty() != true {
			t.Fatalf("A new PieceBuffer must be empty")
		}
		pb.buf.Write([]byte("a"))
		if pb.Empty() != false {
			t.Fatalf("Expected Empty() to return false, got true")
		}

		pb.Flush()

		if pb.Empty() != true {
			t.Fatalf("Expected Empty() to return true, got false")
		}
	})

	t.Run("CreatePieces should return read errors", func(t *testing.T) {
		pb := NewMemoryPieceBuffer(2)

		brokenReader := &brokenReaderT{}
		_, err := pb.CreatePieces(brokenReader)
		if err != errBrokenReader {
			t.Fatalf("Expected error %s, got %s", errBrokenReader, err)
		}
	})
}

func TestPieceReader(t *testing.T) {

	rootDirectory := "./test/pieceReader"
	if err := os.MkdirAll(rootDirectory, 0755); err != nil {
		t.Fatalf("Cannot create root directory %s: %s", rootDirectory, err)
	}

	file1Path := filepath.Join(rootDirectory, "a")
	file1Content := []byte(strings.Repeat("A", 10))

	file2Path := filepath.Join(rootDirectory, "b")
	file2Content := []byte(strings.Repeat("B", 10))

	file3Path := filepath.Join(rootDirectory, "c")
	file3Content := []byte(strings.Repeat("C", 10))

	if err := ioutil.WriteFile(file1Path, file1Content, 0755); err != nil {
		t.Fatalf("Cannot write file %s: %s", file1Path, err)
	}
	if err := ioutil.WriteFile(file2Path, file2Content, 0755); err != nil {
		t.Fatalf("Cannot write file %s: %s", file1Path, err)
	}
	if err := ioutil.WriteFile(file3Path, file3Content, 0755); err != nil {
		t.Fatalf("Cannot write file %s: %s", file1Path, err)
	}

	gorrentFiles := []gorrent.File{
		gorrent.File{IsDir: false, Length: 10, Name: "a"},
		gorrent.File{IsDir: true, Length: 0, Name: "whatever"},
		gorrent.File{IsDir: false, Length: 10, Name: "b"},
		gorrent.File{IsDir: true, Length: 0, Name: "whatever2"},
		gorrent.File{IsDir: false, Length: 10, Name: "c"},
	}

	defer func() {
		if err := os.RemoveAll(rootDirectory); err != nil {
			t.Fatalf("Cannot remove root directory %s: %s", rootDirectory, err)
		}
	}()

	pieceReader := NewPieceReader(fs.NewFileSystem())

	t.Run("ReadPiece returns proper bytes", func(t *testing.T) {
		testData := []struct {
			pieceIndex   int64
			expectedData []byte
		}{
			{0, file1Content},
			{1, file2Content},
			{2, file3Content},
		}

		for _, tdata := range testData {
			data, err := pieceReader.ReadPiece(rootDirectory, gorrentFiles, tdata.pieceIndex, 10)
			if err != nil {
				t.Fatalf("Expected err to be nil, got %s", err)
			}

			if bytes.Equal(data, tdata.expectedData) == false {
				t.Fatalf("Expected data to be %v, got %v", tdata.expectedData, data)
			}
		}
	})

	t.Run("ReadPiece properly deal with piece extending on multiple files", func(t *testing.T) {
		data, err := pieceReader.ReadPiece(rootDirectory, gorrentFiles, 0, 30)
		if err != nil {
			t.Fatalf("Expected err to be nil, got %s", err)
		}

		expectedData := append(file1Content, file2Content...)
		expectedData = append(expectedData, file3Content...)
		if bytes.Equal(data, expectedData) == false {
			t.Fatalf("Expected data to be %v, got %v", expectedData, data)
		}
	})

	t.Run("ReadPiece properly deal with smaller pieces", func(t *testing.T) {
		expectedData := append(file1Content, file2Content...)
		expectedData = append(expectedData, file3Content...)

		size := 3
		for i := 0; i < len(expectedData)/size; i++ {
			data, err := pieceReader.ReadPiece(rootDirectory, gorrentFiles, int64(i), size)
			if err != nil {
				t.Fatalf("Expected err to be nil, got %s", err)
			}

			if bytes.Equal(data, expectedData[i*size:(i*size)+size]) == false {
				t.Fatalf("Expected data to be %v, got %v", expectedData[i*size:(i*size)+size], data)
			}
		}
	})

	t.Run("ReadPiece returns error on partial or empty piece", func(t *testing.T) {
		data, err := pieceReader.ReadPiece(rootDirectory, gorrentFiles, 4, 10)
		if data != nil {
			t.Fatalf("Expected data to be nil, got %v", data)
		}

		if err != ErrReadPieceNoData {
			t.Fatalf("Expected err to be %s, got %s", ErrReadPieceNoData, err)
		}

		data, err = pieceReader.ReadPiece(rootDirectory, gorrentFiles, 4, 7)
		if bytes.Equal(data, []byte("CC")) == false {
			t.Fatalf("Expected data to be %v, got %v", []byte("CC"), data)
		}
	})
}
