package gorrent

import (
	"crypto/sha1"
	"time"
)

const (
	// DefaultPieceLength is a standard default for the number of bytes of each gorrent pieces
	DefaultPieceLength = 256000
)

// Gorrent is a struct holding informations about the shared file(s).
type Gorrent struct {
	Files        []File
	Announce     string
	CreationDate time.Time
	Pieces       []Sha1Hash
	PieceLength  int
}

// File is a struct containing shared file details
type File struct {
	Name   string
	Length int64
	Hash   Sha1Hash
}

// Sha1Hash is an alias for sha1 hashes
type Sha1Hash [sha1.Size]byte

// TotalFileSize return the summed size of all files in this gorrent
func (g *Gorrent) TotalFileSize() int64 {
	var t int64

	for _, f := range g.Files {
		t += f.Length
	}

	return t
}
