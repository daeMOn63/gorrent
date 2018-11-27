package cmd

import (
	"flag"
	"io"
	"log"
	"time"

	"github.com/daeMOn63/gorrent/buffer"
	"github.com/daeMOn63/gorrent/fs"
	"github.com/daeMOn63/gorrent/gorrent"
	"github.com/daeMOn63/gorrent/peer"
	"github.com/daeMOn63/gorrent/peer/server"
	"github.com/daeMOn63/gorrent/tracker"
)

// PeerDaemon is a cli command, allowing to start a gorrent peer
type PeerDaemon struct {
	flagSet *flag.FlagSet

	configPath string
}

var _ Command = &PeerDaemon{}

// NewPeerDaemon instantiates the command
func NewPeerDaemon() Command {

	cmd := &PeerDaemon{
		flagSet: flag.NewFlagSet("peerd", flag.ExitOnError),
	}

	cmd.flagSet.StringVar(&cmd.configPath, "config", "/etc/gorrent/peerd.json", "peerd configuration file path")

	return cmd
}

// FlagSet returns command flags
func (c *PeerDaemon) FlagSet() *flag.FlagSet {
	return c.flagSet
}

// Run executes the command
func (c *PeerDaemon) Run(w io.Writer, r io.Reader) error {
	filesystem := fs.NewFileSystem()
	validator := peer.NewConfigValidator()

	configurator := peer.NewJSONConfigurator(filesystem, validator)

	cfg, err := configurator.Load(c.configPath)
	if err != nil {
		return err
	}

	store, err := peer.NewStore(cfg.DbPath, 0600)
	if err != nil {
		return err
	}

	// Start watcher
	fileBuffer := buffer.NewFile(filesystem, cfg.TmpPath)

	var peerID gorrent.PeerID
	peerID.SetString(cfg.ID)
	// TODO: move timeout to config
	peerClient := peer.NewClient(2 * time.Second)

	peerData := *gorrent.NewPeer(cfg.ID, cfg.PublicIP, cfg.PublicPort)
	tracker := tracker.NewClient(peerData, cfg.TrackerProtocol)

	watcher := peer.NewWatcher(store, filesystem, fileBuffer, tracker, peerClient)
	go func() {
		if err := watcher.Watch(); err != nil {
			log.Println("watcher error: ", err)
		}
	}()

	// Start announcer
	announcer := peer.NewAnnouncer(store, tracker, time.Duration(cfg.AnnounceDelay)*time.Millisecond)
	go announcer.AnnounceForever()

	// Start public server
	publicServer := server.NewPublicServer(peerData, filesystem, store)
	go func() {
		if err := publicServer.Listen(); err != nil {
			log.Println("public server error: ", err)
		}
	}()

	// Start local server
	localServer := server.NewLocalServer(cfg.SockPath, filesystem, store)

	return localServer.Listen()
}
