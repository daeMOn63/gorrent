package gorrent

import (
	"math/rand"
	"time"
)

// RandomSha1Hash returns a random Sha1Hash
func RandomSha1Hash() Sha1Hash {
	rand.Seed(time.Now().UnixNano())

	hash := make([]byte, 20)
	rand.Read(hash)

	var sha1Hash Sha1Hash

	copy(sha1Hash[:], hash)

	return sha1Hash
}
