package gorrent

import (
	"crypto/sha1"
	"encoding/hex"
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

// Bytes returns the Sha1Hash as a byte slice
func (s Sha1Hash) Bytes() []byte {
	return s[:]
}

// HexString returns the hexadecimal string representation of Sha1Hash
func (s Sha1Hash) HexString() string {
	return hex.EncodeToString(s.Bytes())
}

// TotalFileSize return the summed size of all files in this gorrent
func (g *Gorrent) TotalFileSize() uint64 {
	var t uint64

	for _, f := range g.Files {
		t += uint64(f.Length)
	}

	return t
}

// InfoHash returns the gorrent InfoHash, used to uniquely identify it
func (g *Gorrent) InfoHash() Sha1Hash {
	hash := sha1.New()
	for _, f := range g.Files {
		hash.Write(f.Hash[:])
	}

	hash.Write([]byte(g.CreationDate.String()))

	out := Sha1Hash{}
	copy(out[:], hash.Sum(nil))

	return out
}
