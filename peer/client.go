package peer

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"

	"github.com/daeMOn63/gorrent/gorrent"
)

// Client interface defines a peer Client
type Client interface {
	GetPiece(peerAddr gorrent.PeerAddr, chunkRequest *ChunkRequest, chunkSize int) ([]byte, error)
}

type client struct {
	readTimeout time.Duration
}

var _ Client = &client{}

// NewClient creates a new peer Client
func NewClient(readTimeout time.Duration) Client {
	return &client{
		readTimeout: readTimeout,
	}
}

// GetPiece fetch a gorrent piece from given peer or return an error on failure
func (c *client) GetPiece(peerAddr gorrent.PeerAddr, chunkRequest *ChunkRequest, chunkSize int) ([]byte, error) {
	conn, err := net.Dial("udp", peerAddr.String())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.BigEndian, chunkRequest); err != nil {
		return nil, err
	}

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return nil, err
	}

	resp := make([]byte, 0, chunkSize)

	for len(resp) < cap(resp) {
		data := make([]byte, MaxUDPPacketSize)

		deadLine := time.Now().Add(c.readTimeout)
		conn.SetReadDeadline(deadLine)
		n, err := conn.Read(data)
		if err != nil {
			return nil, err
		}
		resp = append(resp, data[:n]...)
	}

	return resp, nil
}
