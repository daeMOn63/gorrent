package cmd

import (
	"flag"
	"io"
	"log"
	"time"

	"github.com/daeMOn63/gorrent/tracker"
	"github.com/daeMOn63/gorrent/tracker/actions"
	"github.com/daeMOn63/gorrent/tracker/handlers"
	"github.com/daeMOn63/gorrent/tracker/store"
)

// TrackerDaemon is a cli command, allowing to start a gorrent Tracker
type TrackerDaemon struct {
	flagSet *flag.FlagSet

	bind         string
	maxPeerAge   int
	readTimeout  int64
	writeTimeout int64
}

var _ Command = &TrackerDaemon{}

// NewTrackerDaemon instantiates the command
func NewTrackerDaemon() Command {

	cmd := &TrackerDaemon{
		flagSet: flag.NewFlagSet("trackerd", flag.ExitOnError),
	}

	cmd.flagSet.StringVar(&cmd.bind, "bind", ":4444", "interface:port where the tracker will listen on.")
	cmd.flagSet.IntVar(&cmd.maxPeerAge, "maxPeerAge", 5000, "threshold in millisecond where peer are considered dead if they not send an announce")
	cmd.flagSet.Int64Var(&cmd.readTimeout, "read-timeout", 100, "maximum network read time")
	cmd.flagSet.Int64Var(&cmd.writeTimeout, "write-timeout", 100, "maximum network write time")

	return cmd
}

// FlagSet returns command flags
func (c *TrackerDaemon) FlagSet() *flag.FlagSet {

	return c.flagSet
}

// Run executes the command
func (c *TrackerDaemon) Run(w io.Writer, r io.Reader) error {
	if c.bind == "" {
		return ErrRequiredFlag{Name: "bind"}
	}

	actionReader := actions.NewReader()
	actionRouter := actions.NewRouter()

	announceStore := store.NewAnnounceMemory()
	announceHandler := handlers.NewAnnounce(announceStore, time.Duration(c.maxPeerAge)*time.Millisecond)

	actionRouter.Register(actions.AnnounceID, announceHandler)

	cfg := tracker.ServerConfig{
		Addr:     c.bind,
		Protocol: "udp",
	}

	t := tracker.NewServer(cfg, actionReader, actionRouter)
	log.Printf("tracker listening on udp %s", c.bind)
	return t.Listen()
}
