package buffer

import (
	"bufio"
	"crypto/sha1"
	"errors"
	"io"
	"path/filepath"

	"github.com/daeMOn63/gorrent/fs"

	"github.com/daeMOn63/gorrent/gorrent"
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

		pb.buf.Write(buf[:n])
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
	lastBytes := make([]byte, 0, pb.pieceLength)
	lastBytes = append(lastBytes, pb.buf.Bytes()...)

	if len(lastBytes) < pb.pieceLength {
		lastBytes = lastBytes[:pb.pieceLength]
	}

	pb.buf.Reset()

	return gorrent.Sha1Hash(sha1.Sum(lastBytes))
}

var (
	// ErrReadPieceNoData is returned when nothing could be read from given pieceID
	ErrReadPieceNoData = errors.New("no data")
	// ErrReadPieceInvalidData is returned when the length of data doesn't match the expected pieceLength
	ErrReadPieceInvalidData = errors.New("invalid data")
)

// PieceReader allow to read a Piece
type PieceReader interface {
	ReadPiece(workingDir string, files []gorrent.File, pieceID int64, pieceLength int) ([]byte, error)
}

type pieceReader struct {
	filesystem fs.FileSystem
}

var _ PieceReader = &pieceReader{}

// NewPieceReader creates a new piece Reader
func NewPieceReader(filesystem fs.FileSystem) PieceReader {
	return &pieceReader{
		filesystem: filesystem,
	}
}

func (p *pieceReader) ReadPiece(workingDir string, files []gorrent.File, pieceID int64, pieceLength int) ([]byte, error) {
	firstByte := int64(pieceLength) * pieceID
	chunkData := make([]byte, 0, pieceLength)
	currentPos := int64(0)

	for _, f := range files {
		if f.IsDir {
			continue
		}

		// Done reading
		missing := pieceLength - len(chunkData)
		if missing == 0 {
			break
		}

		prevPos := currentPos
		currentPos += f.Length

		path := filepath.Join(workingDir, f.Name)

		if currentPos >= firstByte {
			// On first file, we may need to skip some initial bytes if the piece start in the middle of it.
			startOffset := int64(0)
			if prevPos < firstByte {
				startOffset = firstByte - prevPos
			}

			err := func() error {
				fd, err := p.filesystem.Open(path)
				if err != nil {
					return err
				}
				defer fd.Close()

				if startOffset > 0 {
					if _, err := fd.Seek(startOffset, 0); err != nil {
						return err
					}
				}

				r := bufio.NewReader(fd)

				maxRead := int64(missing)
				if (f.Length - startOffset) < maxRead {
					maxRead = f.Length - startOffset
				}

				buf := make([]byte, maxRead)
				n, err := r.Read(buf)
				if err != nil {
					return err
				}

				chunkData = append(chunkData, buf[:n]...)

				return nil
			}()

			if err != nil {
				return nil, err
			}
		}
	}

	if len(chunkData) == 0 {
		return nil, ErrReadPieceNoData
	}

	if len(chunkData) != pieceLength {
		return nil, ErrReadPieceInvalidData
	}

	return chunkData, nil
}
