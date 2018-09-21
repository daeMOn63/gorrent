package gorrent

import (
	"bytes"
	"crypto/sha1"
	"reflect"
	"testing"
	"time"
)

func TestHasher(t *testing.T) {
	t.Run("InfoHash creates correct hashes", func(t *testing.T) {
		sha1Hash := sha1.New()

		hasher := NewHasher()

		now := time.Now()
		g := &Gorrent{
			CreationDate: now,
		}

		var i uint8
		for i = 0; i < 5; i++ {
			var hash Sha1Hash
			b := bytes.Repeat([]byte{i}, 20)
			copy(hash[:], b)

			g.Files = append(g.Files, File{
				Hash: hash,
			})

			sha1Hash.Write(b)
		}
		sha1Hash.Write([]byte(now.String()))

		hash := hasher.InfoHash(g)

		var expectedHash Sha1Hash
		copy(expectedHash[:], sha1Hash.Sum(nil))

		if reflect.DeepEqual(hash, expectedHash) == false {
			t.Fatalf("Expected hash to be %v, got %v", expectedHash, hash)
		}
	})
}
