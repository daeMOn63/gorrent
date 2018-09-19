package buffer

import (
	"encoding/hex"
	"reflect"
	"testing"
)

func TestByteBuffer(t *testing.T) {
	t.Run("empty initial buffer must have zero length", func(t *testing.T) {
		b := NewByteBuffer(nil)
		if b.Len() != 0 {
			t.Fatalf("Expected len to be 0, got %d", b.Len())
		}
	})

	t.Run("non empty initial buffer must be preserved", func(t *testing.T) {
		buf := []byte("abcd")
		b := NewByteBuffer(buf)

		if b.Len() != len(buf) {
			t.Fatalf("Expected len to be %d, got %d", len(buf), b.Len())
		}

		if reflect.DeepEqual(b.Bytes(), buf) == false {
			t.Fatalf("Expected bytes to be %v, got %v", buf, b.Bytes())
		}
	})

	t.Run("Write must append to internal buffer and update length", func(t *testing.T) {
		b := NewByteBuffer(nil)
		buf := []byte("abcd")
		n, err := b.Write(buf)
		if n != len(buf) {
			t.Fatalf("Expected n to be %d, got %d", len(buf), n)
		}

		if err != nil {
			t.Fatalf("Expected err to be nil, got %s", err)
		}

		if reflect.DeepEqual(b.Bytes(), buf) == false {
			t.Fatalf("Expected bytes to be %v, got %v", buf, b.Bytes())
		}

		n, err = b.Write(buf)
		if n != len(buf) {
			t.Fatalf("Expected n to be %d, got %d", len(buf), n)
		}

		if err != nil {
			t.Fatalf("Expected err to be nil, got %s", err)
		}

		expectedBuf := append(buf, buf...)
		if reflect.DeepEqual(b.Bytes(), expectedBuf) == false {
			t.Fatalf("Expected bytes to be %v, got %v", expectedBuf, b.Bytes())
		}
	})

	t.Run("Sha1 must return proper sha1", func(t *testing.T) {
		b := NewByteBuffer(nil)

		expectedSha1 := "da39a3ee5e6b4b0d3255bfef95601890afd80709"
		sha1b := b.Sha1()
		sha1 := hex.EncodeToString(sha1b[:])

		if sha1 != expectedSha1 {
			t.Fatalf("Expected sha1 to be %s, got %s", expectedSha1, sha1)
		}

		b.Write([]byte("abcd"))

		expectedSha1 = "81fe8bfe87576c3ecb22426f8e57847382917acf"
		sha1b = b.Sha1()
		sha1 = hex.EncodeToString(sha1b[:])
		if sha1 != expectedSha1 {
			t.Fatalf("Expected sha1 to be %s, got %s", expectedSha1, sha1)
		}
	})

	t.Run("Reset must reset length and content", func(t *testing.T) {
		b := NewByteBuffer([]byte("abcd"))

		b.Reset()
		if b.Len() != 0 {
			t.Fatalf("Expected len to be 0, got %d", b.Len())
		}

		if reflect.DeepEqual(b.Bytes(), []byte{}) == false {
			t.Fatalf("Expected bytes to be empty, got %v", b.Bytes())
		}
	})
}
