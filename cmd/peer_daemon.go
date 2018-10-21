package cmd

import (
	"flag"
	"io"
	"log"

	"github.com/daeMOn63/gorrent/fs"
	"github.com/daeMOn63/gorrent/peer"
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

	server := peer.NewServer(cfg, filesystem)

	log.Printf("peer daemon listening on unix://%s", cfg.SockPath)
	return server.Listen()
}
