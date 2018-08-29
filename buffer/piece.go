package buffer

import (
	"bufio"
	"gorrent/gorrent"
	"io"
)

// PieceBuffer is an interface for pieceBuffer
type PieceBuffer interface {
	PieceLength() int
	CreatePieces(io.Reader) ([]gorrent.Sha1Hash, error)
	Empty() bool
	Flush() gorrent.Sha1Hash
}

// MemoryPieceBuffer is a struct allowing to create Pieces from bytes using only memory as storage
type MemoryPieceBuffer struct {
	buf         *ByteBuffer
	pieceLength int
}

var _ PieceBuffer = &MemoryPieceBuffer{}

// NewMemoryPieceBuffer create a new pieceBuffer
func NewMemoryPieceBuffer(length int) *MemoryPieceBuffer {

	return &MemoryPieceBuffer{
		buf:         NewByteBuffer(make([]byte, 0, length)),
		pieceLength: length,
	}
}

// PieceLength return the piece size of the pieceBuffer
func (pb *MemoryPieceBuffer) PieceLength() int {
	return pb.pieceLength
}

// CreatePieces create as much pieces as possible (from pieceLength) and leave remaining bytes in internal buffer
func (pb *MemoryPieceBuffer) CreatePieces(r io.Reader) ([]gorrent.Sha1Hash, error) {
	var newPieces []gorrent.Sha1Hash

	br := bufio.NewReader(r)

	for {
		remaining := pb.pieceLength - pb.buf.Len()

		buf := make([]byte, remaining)
		n, err := br.Read(buf)
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			return newPieces, nil
		}

		if n < len(buf) {
			buf = buf[:n]
		}

		pb.buf.Write(buf)
		if pb.buf.Len() == pb.pieceLength {
			newPieces = append(newPieces, pb.buf.Sha1())
			pb.buf.Reset()
		}
	}
}

// Empty return true when the internal buffer is empty.
func (pb *MemoryPieceBuffer) Empty() bool {
	return pb.buf.Len() == 0
}

// Flush return a sha1 hash of the current buffer and reset it.
// It returns an error when Flush is called when internal buffer is empty.
func (pb *MemoryPieceBuffer) Flush() gorrent.Sha1Hash {
	hash := pb.buf.Sha1()
	pb.buf.Reset()

	return gorrent.Sha1Hash(hash)
}
