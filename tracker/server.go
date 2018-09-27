package tracker

import (
	"log"
	"net"

	"github.com/daeMOn63/gorrent/tracker/actions"
)

const (
	// MaxUDPPacketSize defines the maximum size of udp packets
	MaxUDPPacketSize = 1024
)

// Server is the base struct for the gorrent tracker server
type Server interface {
	Listen() error
}

// ServerConfig describe configuration needed for the tracker server
type ServerConfig struct {
	Addr     string
	Protocol string
}

type server struct {
	cfg          ServerConfig
	actionReader actions.Reader
	actionRouter actions.Router
}

var _ Server = &server{}

// NewServer creates a new server
func NewServer(cfg ServerConfig, reader actions.Reader, router actions.Router) Server {
	return &server{
		cfg:          cfg,
		actionReader: reader,
		actionRouter: router,
	}
}

// Listen makes the server to listen on configured address and protocol
func (t *server) Listen() error {
	link, err := net.ListenPacket(t.cfg.Protocol, t.cfg.Addr)
	if err != nil {
		return err
	}
	defer link.Close()

	for {
		buf := make([]byte, MaxUDPPacketSize)
		n, client, err := link.ReadFrom(buf)
		if err != nil {
			log.Print("Error while reading from link: ", err)

			continue
		}

		buf = buf[:n]

		action, err := t.actionReader.Read(buf)
		if err != nil {
			log.Printf("failed to read action: %s", err)

			continue
		}
		log.Printf("[%s] sent action %#x", client, action.ID())
		resp, err := t.actionRouter.Handle(action)
		if err != nil {
			log.Printf("[%s] action %#x handler error: %s", client, action.ID(), err)

			continue
		}

		_, err = link.WriteTo(resp, client)
		if err != nil {
			log.Printf("[%s] writing action %#x response error: %s", client, action.ID(), err)

			continue
		}
	}
}
