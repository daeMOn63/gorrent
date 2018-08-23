package gorrent

import (
	"bytes"
	"crypto/sha1"
	"reflect"
	"testing"
)

func TestPieceBuffer(t *testing.T) {

	t.Run("NewPieceBuffer should create a new buffer", func(t *testing.T) {
		var lengths = []int64{1234, 0, 666}
		for _, length := range lengths {
			pb := NewPieceBuffer(length)
			if pb.pieceLength != length {
				t.Fatalf("Expected length '%d', got '%d'", length, pb.pieceLength)
			}
		}
	})

	t.Run("Length() should return proper length", func(t *testing.T) {
		var lengths = []int64{1234, 0, 666}
		for _, length := range lengths {
			pb := NewPieceBuffer(length)
			if pb.Length() != length {
				t.Fatalf("Expected length '%d', got '%d'", length, pb.pieceLength)
			}
		}
	})

	t.Run("CreatePieces should properly create pieces", func(t *testing.T) {
		pb := NewPieceBuffer(int64(10))
		data := []byte("abcdefghijlkmnopqrstuvwxyz")

		expectedPieces := []Sha1Hash{
			sha1.Sum([]byte("abcdefghij")),
			sha1.Sum([]byte("lkmnopqrst")),
		}

		expectedRemainingBuffer := []byte("uvwxyz")

		pieces := pb.CreatePieces(data)

		if reflect.DeepEqual(pieces, expectedPieces) == false {
			t.Fatalf("Expected piece hashes %v, got %v", expectedPieces, pieces)
		}

		if bytes.Equal(pb.buf, expectedRemainingBuffer) == false {
			t.Fatalf("Expected remaining buffer '%s', got '%s'", expectedRemainingBuffer, pb.buf)
		}

		expectedPieces = []Sha1Hash{
			sha1.Sum([]byte("uvwxyz1234")),
		}
		expectedRemainingBuffer = []byte("56789")

		pieces = pb.CreatePieces([]byte("123456789"))

		if reflect.DeepEqual(pieces, expectedPieces) == false {
			t.Fatalf("Expected piece hashes %v, got %v", expectedPieces, pieces)
		}

		if bytes.Equal(pb.buf, expectedRemainingBuffer) == false {
			t.Fatalf("Expected remaining buffer '%s', got '%s'", expectedRemainingBuffer, pb.buf)
		}
	})

	t.Run("Flush should return last piece and clear itself", func(t *testing.T) {
		pb := NewPieceBuffer(3)

		pb.CreatePieces([]byte("12345"))
		lastPiece := pb.Flush()

		expectedHash := Sha1Hash(sha1.Sum([]byte("45")))
		if lastPiece != expectedHash {
			t.Fatalf("Expected last piece hash to be %v, got %v", expectedHash, lastPiece)
		}

		if len(pb.buf) != 0 {
			t.Fatalf("Expected buf lenght to be %d, got %d", 0, len(pb.buf))
		}

		lastPiece = pb.Flush()
		expectedHash = Sha1Hash(sha1.Sum([]byte{}))
		if lastPiece != expectedHash {
			t.Fatalf("Expected hash %v, got %v", expectedHash, lastPiece)
		}
	})

	t.Run("Empty should indicate when the internal buffer is empty", func(t *testing.T) {
		pb := NewPieceBuffer(2)

		if pb.Empty() != true {
			t.Fatalf("A new PieceBuffer must be empty")
		}
		pb.buf = []byte("a")
		if pb.Empty() != false {
			t.Fatalf("Expected Empty() to return false, got true")
		}

		pb.Flush()

		if pb.Empty() != true {
			t.Fatalf("Expected Empty() to return true, got false")
		}
	})
}
