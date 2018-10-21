package gorrent

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"net"
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

// IP2Long convert a net.IP to uint32
func IP2Long(ip net.IP) uint32 {
	var long uint32
	binary.Read(bytes.NewBuffer(ip.To4()), binary.BigEndian, &long)
	return long
}
