package cmd

import (
	"flag"
	"fmt"
	"io"
)

// Tracker is a cli command, allowing to start a gorrent Tracker
type Tracker struct {
	flagSet *flag.FlagSet
}

var _ Command = &Tracker{}

// NewTracker instantiates the command
func NewTracker() Command {

	cmd := &Tracker{
		flagSet: flag.NewFlagSet("tracker", flag.ExitOnError),
	}

	return cmd
}

// FlagSet returns command flags
func (c *Tracker) FlagSet() *flag.FlagSet {
	return c.flagSet
}

// Run executes the command
func (c *Tracker) Run(w io.Writer, r io.Reader) error {
	fmt.Fprintln(w, "todo")
	return nil
}
