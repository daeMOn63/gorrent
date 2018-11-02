package tracker

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"

	"github.com/daeMOn63/gorrent/gorrent"
	"github.com/daeMOn63/gorrent/tracker/actions"
)

var (
	// ErrInvalidResponse is returned when the client failed to decode the response
	ErrInvalidResponse = errors.New("invalid response")
)

// Client interface list the tracker client methods to interact with the server
type Client interface {
	Announce(g *gorrent.Gorrent, evt actions.AnnounceEvent, status actions.AnnounceStatus) ([]gorrent.PeerAddr, error)
}

type client struct {
	peer     gorrent.Peer
	protocol string
}

var _ Client = &client{}

// NewClient returns a new tracker client
func NewClient(peer gorrent.Peer, protocol string) Client {
	return &client{
		peer:     peer,
		protocol: protocol,
	}
}

// Announce reports the client gorrent status to the server
// This will allow the tracker to list (or unlist) the client from the peer list
func (c *client) Announce(g *gorrent.Gorrent, evt actions.AnnounceEvent, status actions.AnnounceStatus) ([]gorrent.PeerAddr, error) {
	conn, err := net.Dial(c.protocol, g.Announce)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	data := &actions.Announce{
		InfoHash: g.InfoHash(),
		Peer:     c.peer,
		Status:   status,
		Event:    evt,
	}

	buf := bytes.NewBuffer(nil)
	buf.WriteByte(uint8(actions.AnnounceID))
	if err := binary.Write(buf, binary.BigEndian, data); err != nil {
		return nil, err
	}

	n, err := conn.Write(buf.Bytes())
	if err != nil {
		return nil, err
	}

	resp := make([]byte, 1024)

	n, err = conn.Read(resp)
	if err != nil {
		return nil, err
	}

	addrs, err := decodeResponse(resp[:n])
	if err != nil {
		return nil, err
	}

	return addrs, nil
}

func decodeResponse(b []byte) ([]gorrent.PeerAddr, error) {
	if len(b)%6 != 0 {
		return nil, ErrInvalidResponse
	}

	var addrs []gorrent.PeerAddr

	for i := 0; i < len(b); i += 6 {
		buf := bytes.NewBuffer(b[i : i+6])
		var addr gorrent.PeerAddr
		if err := binary.Read(buf, binary.BigEndian, &addr); err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}

	return addrs, nil
}

// buf := bytes.NewBuffer(nil)
// err := binary.Write(buf, binary.BigEndian, p)
// if err != nil {
// 	return nil, err
// }

// return buf.Bytes(), nil
