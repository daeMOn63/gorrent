package gorrent

import (
	"crypto/sha1"
)

// PieceBuffer is a struct allowing to create Pieces from bytes
type PieceBuffer struct {
	buf         []byte
	pieceLength int64
}

// NewPieceBuffer create a new pieceBuffer
func NewPieceBuffer(length int64) *PieceBuffer {
	return &PieceBuffer{
		buf:         make([]byte, 0, length),
		pieceLength: length,
	}
}

// Length return the piece size of the pieceBuffer
func (pb *PieceBuffer) Length() int64 {
	return pb.pieceLength
}

// CreatePieces create as much pieces as possible (from pieceLength) and leave remaining bytes in internal buffer
func (pb *PieceBuffer) CreatePieces(data []byte) []Sha1Hash {
	var newPieces []Sha1Hash

	var cur int64
	dataLen := int64(len(data))

	for cur < dataLen {
		end := cur + pb.pieceLength
		if end > dataLen {
			end = dataLen
		}

		pb.buf = append(pb.buf, data[cur:end]...)
		cur += pb.pieceLength

		if int64(len(pb.buf)) > pb.pieceLength {
			hash := sha1.Sum(pb.buf[:pb.pieceLength])
			newPieces = append(newPieces, Sha1Hash(hash))
			pb.buf = pb.buf[pb.pieceLength:]
		}
	}

	return newPieces
}

// Empty return true when the internal buffer is empty.
func (pb *PieceBuffer) Empty() bool {
	return len(pb.buf) == 0
}

// Flush return a sha1 hash of the current buffer and reset it.
// It returns an error when Flush is called when internal buffer is empty.
func (pb *PieceBuffer) Flush() Sha1Hash {
	hash := sha1.Sum(pb.buf[:len(pb.buf)])
	pb.buf = make([]byte, 0, pb.pieceLength)

	return Sha1Hash(hash)
}
