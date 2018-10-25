package tracker

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/daeMOn63/gorrent/tracker/actions"

	"github.com/daeMOn63/gorrent/gorrent"
)

// Client interface list the tracker client methods to interact with the server
type Client interface {
	Announce(g *gorrent.Gorrent, evt actions.AnnounceEvent, status actions.AnnounceStatus) error
}

type client struct {
	peer      actions.Peer
	serverCfg ServerConfig
}

var _ Client = &client{}

// NewClient returns a new tracker client
func NewClient(peer actions.Peer, serverCfg ServerConfig) Client {
	return &client{
		peer:      peer,
		serverCfg: serverCfg,
	}
}

// Announce reports the client gorrent status to the server
// This will allow the tracker to list (or unlist) the client from the peer list
func (c *client) Announce(g *gorrent.Gorrent, evt actions.AnnounceEvent, status actions.AnnounceStatus) error {
	conn, err := net.Dial(c.serverCfg.Protocol, c.serverCfg.Addr)
	if err != nil {
		return err
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
		return err
	}

	n, err := conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	fmt.Printf("Wrote %d bytes to server\n", n)

	resp := make([]byte, 1024)

	n, err = conn.Read(resp)

	fmt.Printf("Server replied: %s\n", resp[:n])

	return nil
}
