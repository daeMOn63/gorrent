package buffer

import (
	"bufio"
	"crypto/sha1"
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

// PieceReader allow to read a Piece
type PieceReader interface {
	ReadPiece(workingDir string, g *gorrent.Gorrent, chunkID int64) ([]byte, error)
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

func (p *pieceReader) ReadPiece(workingDir string, g *gorrent.Gorrent, chunkID int64) ([]byte, error) {
	firstByte := int64(g.PieceLength) * chunkID

	var currentPos int64
	var reading bool

	chunkData := make([]byte, 0, g.PieceLength)

	for _, f := range g.Files {
		currentPos += f.Length

		path := filepath.Join(workingDir, f.Name)

		if !reading {
			if currentPos >= firstByte {
				reading = true
				startOffset := firstByte - (currentPos - f.Length)

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

					maxRead := int64(g.PieceLength)
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
		} else {
			missing := g.PieceLength - len(chunkData)
			if missing == 0 {
				break
			}

			err := func() error {
				fd, err := p.filesystem.Open(path)
				if err != nil {
					return err
				}
				defer fd.Close()

				maxRead := int64(missing)
				if f.Length < maxRead {
					maxRead = f.Length
				}

				r := bufio.NewReader(fd)
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

	return chunkData, nil
}
