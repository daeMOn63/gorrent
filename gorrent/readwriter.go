package gorrent

import (
	"compress/gzip"
	"encoding/gob"
	"io"
)

// ReadWriter defines a gorrent reader and writer
type ReadWriter interface {
	Read(r io.Reader) (*Gorrent, error)
	Write(w io.Writer, g *Gorrent) error
}

type readWriter struct {
}

var _ ReadWriter = &readWriter{}

// NewReadWriter creates a new gorrent reader and writer
func NewReadWriter() ReadWriter {
	return &readWriter{}
}

// Read reads bytes from given reader and returns a gorrent
func (rw *readWriter) Read(r io.Reader) (*Gorrent, error) {
	gzreader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gzreader.Close()

	gob.Register(Gorrent{})
	decoder := gob.NewDecoder(gzreader)

	g := &Gorrent{}
	err = decoder.Decode(g)
	if err != nil {
		return nil, err
	}

	return g, nil
}

// Write encode and write gorrent to given writer
func (rw *readWriter) Write(w io.Writer, g *Gorrent) error {
	gzwriter := gzip.NewWriter(w)
	defer gzwriter.Close()

	gob.Register(Gorrent{})
	encoder := gob.NewEncoder(gzwriter)

	return encoder.Encode(g)
}
