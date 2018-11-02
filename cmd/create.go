package cmd

import (
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/daeMOn63/gorrent/buffer"
	"github.com/daeMOn63/gorrent/creator"
	"github.com/daeMOn63/gorrent/fs"
	"github.com/daeMOn63/gorrent/gorrent"
)

// Create is a cli command, allowing to create gorrent
type Create struct {
	pieceLength int
	src         string
	dst         string
	fsWorkers   int
	announce    string

	flagSet *flag.FlagSet
}

var _ Command = &Create{}

// NewCreate instantiates the command
func NewCreate() Command {

	cmd := &Create{
		flagSet: flag.NewFlagSet("create", flag.ExitOnError),
	}

	cmd.flagSet.StringVar(&cmd.src, "src", "", "Required. File / folder to create the gorrent from.")
	cmd.flagSet.StringVar(&cmd.announce, "announce", "", "Required. Tracker ip / port.")
	cmd.flagSet.StringVar(&cmd.dst, "dst", fmt.Sprintf("./%d.gorrent", time.Now().Unix()), "Output filename")
	cmd.flagSet.IntVar(&cmd.pieceLength, "pieceLength", gorrent.DefaultPieceLength, "Gorrent pieces length.")
	cmd.flagSet.IntVar(&cmd.fsWorkers, "fsWorkers", 10, "Number of parallel workers when accessing file system")
	return cmd
}

// FlagSet returns command flags
func (c *Create) FlagSet() *flag.FlagSet {
	return c.flagSet
}

// Run executes the command
func (c *Create) Run(w io.Writer, r io.Reader) error {

	if c.src == "" {
		return ErrRequiredFlag{Name: "src"}
	}

	if c.dst == "" {
		return ErrRequiredFlag{Name: "dst"}
	}

	if c.announce == "" {
		return ErrRequiredFlag{Name: "announce"}
	}

	pb := buffer.NewMemoryPieceBuffer(c.pieceLength)
	filesystem := fs.NewFileSystem()
	rw := gorrent.NewReadWriter()

	creator := creator.NewCreator(pb, filesystem, rw)

	fmt.Fprintf(w, "creating new gorrent:\n")
	fmt.Fprintf(w, "\t - from: %s\n", c.src)
	fmt.Fprintf(w, "\t - pieceLength: %d bytes\n", c.pieceLength)
	fmt.Fprintf(w, "\t - fsWorkers: %d\n", c.fsWorkers)

	start := time.Now()
	g, err := creator.Create(c.src, c.fsWorkers)
	if err != nil {
		return err
	}
	elapsed := time.Since(start)
	g.Announce = c.announce

	fmt.Fprintf(w, "gorrent created in %s\n", elapsed)
	fmt.Fprintf(w, "\t - announce url: %s\n", g.Announce)
	fmt.Fprintf(w, "\t - files %d\n", len(g.Files))
	fmt.Fprintf(w, "\t - pieces %d\n", len(g.Pieces))
	fmt.Fprintf(w, "\t - total file size: %d bytes\n", g.TotalFileSize())

	if err := creator.Save(c.dst, g); err != nil {
		return err
	}
	fmt.Fprintf(w, "saved new gorrent to %s\n", c.dst)

	return nil
}
