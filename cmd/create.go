package cmd

import (
	"flag"
	"fmt"
	"io"
	"time"

	"gorrent/fs"
	"gorrent/gorrent"
)

// Create is a cli command, allowing to create gorrent
type Create struct {
	pieceLength int
	src         string
	dst         string
	fsWorkers   int
}

var _ Command = &Create{}

// NewCreate instantiate the command
func NewCreate() Command {
	return &Create{}
}

// Flags set cli flags for the create command. You still need to call flag.Parse() when ready.
func (c *Create) Flags() {
	flag.StringVar(&c.src, "src", "", "Required. File / folder to create the gorrent from.")
	flag.StringVar(&c.dst, "dst", fmt.Sprintf("./%d.gorrent", time.Now().Unix()), "Output filename")
	flag.IntVar(&c.pieceLength, "pieceLength", gorrent.DefaultPieceLength, "Gorrent pieces length.")
	flag.IntVar(&c.fsWorkers, "fsWorkers", 10, "Number of parallel workers when accessing file system")
}

// Run execute the command
func (c *Create) Run(w io.Writer, r io.Reader) error {

	if c.src == "" {
		return ErrRequiredFlag{Name: "src"}
	}

	if c.dst == "" {
		return ErrRequiredFlag{Name: "dst"}
	}

	pb := gorrent.NewMemoryPieceBuffer(c.pieceLength)
	filesystem := fs.NewFileSystem()
	creator := gorrent.NewCreator(pb, filesystem)

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
