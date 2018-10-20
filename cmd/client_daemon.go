package cmd

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/daeMOn63/gorrent/tracker/actions"

	"github.com/daeMOn63/gorrent/gorrent"

	"github.com/daeMOn63/gorrent/tracker"
)

// PeerDaemon is a cli command, allowing to start a gorrent peer
type PeerDaemon struct {
	flagSet *flag.FlagSet

	id           string
	publicIP     string
	publicPort   uint
	sockPath     string
	readTimeout  uint
	writeTimeout uint
	dbPath       string
	storagePath  string
}

var _ Command = &PeerDaemon{}

// NewPeerDaemon instantiates the command
func NewPeerDaemon() Command {

	cmd := &PeerDaemon{
		flagSet: flag.NewFlagSet("peerd", flag.ExitOnError),
	}

	cmd.flagSet.StringVar(&cmd.id, "id", "", "peer identifier")
	cmd.flagSet.StringVar(&cmd.publicIP, "public-ip", "", "ip to listen for incomming peer connections")
	cmd.flagSet.UintVar(&cmd.publicPort, "public-port", 4443, "port to listen for incomming peer connections")
	cmd.flagSet.StringVar(&cmd.sockPath, "sock", "/run/gorrentd.sock", "local socket path for interactions with the gorrent client")
	cmd.flagSet.UintVar(&cmd.readTimeout, "read-timeout", 100, "maximum number of milisecond for network reads")
	cmd.flagSet.UintVar(&cmd.writeTimeout, "write-timeout", 100, "maximum number of milisecond for network writes")
	cmd.flagSet.StringVar(&cmd.dbPath, "db", "/var/lib/gorrent/gorrent.db", "path to the peer database")
	cmd.flagSet.StringVar(&cmd.storagePath, "storage", "", "path where the gorrent data will get downloaded")

	return cmd
}

// FlagSet returns command flags
func (c *PeerDaemon) FlagSet() *flag.FlagSet {
	return c.flagSet
}

// Run executes the command
func (c *PeerDaemon) Run(w io.Writer, r io.Reader) error {

	if len(c.id) == 0 {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		c.id = hostname
	}

	var peerID actions.PeerID
	peerID.SetString(c.id)

	ip := net.ParseIP(c.publicIP)
	if ip == nil {
		return fmt.Errorf("invalid public-ip %s", c.publicIP)
	}

	if c.publicPort <= 1024 || c.publicPort > 65535 {
		return fmt.Errorf("invalid public-port %d", c.publicPort)
	}

	// TODO:
	// Start peer listening on c.publicIP:c.publicPort
	// Load gorrent db
	//     check integrity in local workdir
	//     announce

	peer := actions.Peer{
		ID: peerID,
		PeerAddr: actions.PeerAddr{
			IPAddr: ip2Long(ip),
			Port:   uint16(c.publicPort),
		},
	}

	serverCfg := tracker.ServerConfig{
		Addr:     "127.0.0.1:4444",
		Protocol: "udp",
	}

	hasher := gorrent.NewHasher()

	trackerClient := tracker.NewClient(peer, serverCfg, hasher)

	g := &gorrent.Gorrent{}
	evt := actions.AnnounceEventStarted
	status := actions.AnnounceStatus{}

	return trackerClient.Announce(g, evt, status)
}

func ip2Long(ip net.IP) uint32 {
	var long uint32
	binary.Read(bytes.NewBuffer(ip.To4()), binary.BigEndian, &long)
	return long
}
