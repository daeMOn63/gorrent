package cmd

import (
	"flag"
	"fmt"
	"io"
)

// Client is a cli command, allowing to start a gorrent client
type Client struct {
	flagSet *flag.FlagSet
}

var _ Command = &Client{}

// NewClient instantiates the command
func NewClient() Command {

	cmd := &Client{
		flagSet: flag.NewFlagSet("client", flag.ExitOnError),
	}

	return cmd
}

// FlagSet returns command flags
func (c *Client) FlagSet() *flag.FlagSet {
	return c.flagSet
}

// Run executes the command
func (c *Client) Run(w io.Writer, r io.Reader) error {
	fmt.Fprintln(w, "todo")
	return nil
}
