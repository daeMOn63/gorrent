package buffer

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"reflect"
	"strings"
	"testing"

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

		expectedHash := gorrent.Sha1Hash(sha1.Sum([]byte("45")))
		if lastPiece != expectedHash {
			t.Fatalf("Expected last piece hash to be %v, got %v", expectedHash, lastPiece)
		}

		if pb.buf.Len() != 0 {
			t.Fatalf("Expected buf lenght to be %d, got %d", 0, pb.buf.Len())
		}

		lastPiece = pb.Flush()
		expectedHash = gorrent.Sha1Hash(sha1.Sum([]byte{}))
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
