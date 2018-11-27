package peer

import "github.com/daeMOn63/gorrent/gorrent"

const (
	// MaxUDPPacketSize defines the maximum size of udp packets
	MaxUDPPacketSize = 1024
)

// ChunkRequest defines the data transfered on a chunkRequest
type ChunkRequest struct {
	InfoHash gorrent.Sha1Hash
	ChunkID  int64
}
