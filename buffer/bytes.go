package buffer

import (
	"bytes"
	"crypto/sha1"
)

// ByteBuffer extend bytes.Buffer to allow control over the buffer length.
type ByteBuffer struct {
	buffer *bytes.Buffer
	len    int
}

// NewByteBuffer create a new bytes buffer
func NewByteBuffer(b []byte) *ByteBuffer {
	return &ByteBuffer{
		buffer: bytes.NewBuffer(b),
		len:    len(b),
	}
}

// Write writes bytes in the buffer
func (bbuf *ByteBuffer) Write(b []byte) (int, error) {
	bbuf.len += len(b)
	return bbuf.buffer.Write(b)
}

// Len return the current length of the buffer
func (bbuf *ByteBuffer) Len() int {
	return bbuf.len
}

// Bytes return bytes contained by the buffer
func (bbuf *ByteBuffer) Bytes() []byte {
	return bbuf.buffer.Bytes()[:bbuf.len]
}

// Sha1 return sha1 hash of buffer bytes
func (bbuf *ByteBuffer) Sha1() [sha1.Size]byte {
	return sha1.Sum(bbuf.Bytes())
}

// Reset empty the buffer
func (bbuf *ByteBuffer) Reset() {
	bbuf.len = 0
	bbuf.buffer.Reset()
}
