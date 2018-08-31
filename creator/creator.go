package creator

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/gob"
	"io"
	"path/filepath"
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
}

// NewCreator create a new Gorrent Creator
func NewCreator(pb buffer.PieceBuffer, filesystem fs.FileSystem) *Creator {
	return &Creator{
		pieceBuffer: pb,
		filesystem:  filesystem,
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

	for _, path := range filepaths {
		file, err := c.filesystem.Open(filepath.Join(rootDir, path))
		if err != nil {
			return nil, err
		}

		hash := sha1.New()
		buf := bytes.NewBuffer(nil)
		tee := io.TeeReader(file, buf)
		_, err = io.Copy(hash, tee)
		if err != nil {
			return nil, err
		}

		var sha1Hash gorrent.Sha1Hash
		copy(sha1Hash[:], hash.Sum(nil))

		finfo, err := file.Stat()
		if err != nil {
			return nil, err
		}

		g.Files = append(g.Files, gorrent.File{
			Name:   strings.Replace(file.Name(), rootDir, "", 1),
			Length: finfo.Size(),
			Hash:   sha1Hash,
		})

		newPieces, err := c.pieceBuffer.CreatePieces(buf)
		if err != nil {
			return nil, err
		}

		g.Pieces = append(g.Pieces, newPieces...)

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

	gzwriter := gzip.NewWriter(f)
	defer gzwriter.Close()

	gob.Register(gorrent.Gorrent{})
	encoder := gob.NewEncoder(gzwriter)
	return encoder.Encode(g)
}

// Open load gorrent from src file and returns it
func (c *Creator) Open(src string) (*gorrent.Gorrent, error) {
	f, err := c.filesystem.Open(src)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gzreader, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gzreader.Close()

	gob.Register(gorrent.Gorrent{})
	decoder := gob.NewDecoder(gzreader)

	g := &gorrent.Gorrent{}
	err = decoder.Decode(g)
	if err != nil {
		return nil, err
	}

	return g, nil
}
