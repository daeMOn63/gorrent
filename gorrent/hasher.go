package gorrent

import (
	"crypto/sha1"
)

// Hasher interface list methods for implementing a gorrent Hasher
type Hasher interface {
	InfoHash(g *Gorrent) Sha1Hash
}

type hasher struct{}

// NewHasher creates a new Hasher
func NewHasher() Hasher {
	return &hasher{}
}

var _ Hasher = &hasher{}

// InfoHash returns the Gorrent InfoHash hash used
// to uniquely identify it.
func (h *hasher) InfoHash(g *Gorrent) Sha1Hash {
	hash := sha1.New()
	for _, f := range g.Files {
		hash.Write(f.Hash[:])
	}

	hash.Write([]byte(g.CreationDate.String()))

	out := Sha1Hash{}
	copy(out[:], hash.Sum(nil))

	return out
}
