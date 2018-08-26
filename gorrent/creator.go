package gorrent

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/gob"
	"io"
	"path/filepath"
	"time"

	"gorrent/fs"
)

// Creator allow to create Gorrent
type Creator struct {
	pieceBuffer PieceBuffer
	filesystem  fs.FileSystem
}

// NewCreator create a new Gorrent Creator
func NewCreator(pb PieceBuffer, filesystem fs.FileSystem) *Creator {
	return &Creator{
		pieceBuffer: pb,
		filesystem:  filesystem,
	}
}

// Create return a new gorrent from files under rootDir and given pieceLength
func (c *Creator) Create(rootDir string) (*Gorrent, error) {

	filepaths, err := c.filesystem.FindFiles(rootDir)
	if err != nil {
		return nil, err
	}

	g := &Gorrent{
		PieceLength:  c.pieceBuffer.PieceLength(),
		CreationDate: time.Now(),
	}

	for _, path := range filepaths {
		fsFile, err := c.filesystem.Open(filepath.Join(rootDir, path))
		if err != nil {
			return nil, err
		}

		hash := sha1.New()
		var buf bytes.Buffer
		tee := io.TeeReader(fsFile, &buf)
		_, err = io.Copy(hash, tee)
		if err != nil {
			return nil, err
		}

		var sha1Hash Sha1Hash
		copy(sha1Hash[:], hash.Sum(nil))

		g.Files = append(g.Files, File{
			Name:   fsFile.Path(),
			Length: fsFile.Size(),
			Hash:   sha1Hash,
		})

		newPieces, err := c.pieceBuffer.CreatePieces(&buf)
		if err != nil {
			return nil, err
		}

		g.Pieces = append(g.Pieces, newPieces...)

		fsFile.Close()
	}

	if !c.pieceBuffer.Empty() {
		lastPiece := c.pieceBuffer.Flush()
		g.Pieces = append(g.Pieces, lastPiece)
	}

	return g, nil
}

// Save write given gorrent to dst file
func (c *Creator) Save(dst string, g *Gorrent) error {
	f, err := c.filesystem.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	gzwriter := gzip.NewWriter(f)
	defer gzwriter.Close()

	gob.Register(Gorrent{})
	encoder := gob.NewEncoder(gzwriter)
	return encoder.Encode(g)
}

// Open load gorrent from src file and returns it
func (c *Creator) Open(src string) (*Gorrent, error) {
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

	gob.Register(Gorrent{})
	decoder := gob.NewDecoder(gzreader)

	g := &Gorrent{}
	err = decoder.Decode(g)
	if err != nil {
		return nil, err
	}

	return g, nil
}
