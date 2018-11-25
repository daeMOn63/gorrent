package creator

import (
	"bytes"
	"crypto/sha1"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/daeMOn63/gorrent/buffer"
	"github.com/daeMOn63/gorrent/fs"
	"github.com/daeMOn63/gorrent/gorrent"
)

// Creator allow to create Gorrent
type Creator struct {
	pieceBuffer buffer.PieceBuffer
	filesystem  fs.FileSystem
	readWriter  gorrent.ReadWriter
}

// NewCreator create a new Gorrent Creator
func NewCreator(pb buffer.PieceBuffer, filesystem fs.FileSystem, rw gorrent.ReadWriter) *Creator {
	return &Creator{
		pieceBuffer: pb,
		filesystem:  filesystem,
		readWriter:  rw,
	}
}

// Create return a new gorrent from files under rootDir and given pieceLength
func (c *Creator) Create(rootDir string, maxWorkers int) (*gorrent.Gorrent, error) {

	filepaths, err := c.filesystem.FindFiles(rootDir, maxWorkers)
	if err != nil {
		return nil, err
	}
	g := &gorrent.Gorrent{
		PieceLength:  c.pieceBuffer.PieceLength(),
		CreationDate: time.Now(),
	}

	sort.Strings(filepaths)
	for _, path := range filepaths {
		file, err := c.filesystem.Open(filepath.Join(rootDir, path))
		if err != nil {
			return nil, err
		}

		finfo, err := file.Stat()
		if err != nil {
			return nil, err
		}

		var sha1Hash gorrent.Sha1Hash
		hash := sha1.New()
		buf := bytes.NewBuffer(nil)
		if !finfo.IsDir() {
			tee := io.TeeReader(file, buf)
			_, err = io.Copy(hash, tee)
			if err != nil {
				return nil, err
			}
			copy(sha1Hash[:], hash.Sum(nil))
		}

		g.Files = append(g.Files, gorrent.File{
			Name:   strings.Replace(file.Name(), rootDir, "", 1),
			Length: finfo.Size(),
			IsDir:  finfo.IsDir(),
			Hash:   sha1Hash,
		})

		if buf.Len() > 0 {
			newPieces, err := c.pieceBuffer.CreatePieces(buf)
			if err != nil {
				return nil, err
			}

			g.Pieces = append(g.Pieces, newPieces...)
		}
		file.Close()
	}

	if !c.pieceBuffer.Empty() {
		lastPiece := c.pieceBuffer.Flush()
		g.Pieces = append(g.Pieces, lastPiece)
	}

	return g, nil
}

// Save write given gorrent to dst file
func (c *Creator) Save(dst string, g *gorrent.Gorrent) error {
	f, err := c.filesystem.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	return c.readWriter.Write(f, g)
}

// Open load gorrent from src file and returns it
func (c *Creator) Open(src string) (*gorrent.Gorrent, error) {
	f, err := c.filesystem.Open(src)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return c.readWriter.Read(f)
}
