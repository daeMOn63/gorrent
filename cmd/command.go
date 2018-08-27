package cmd

import (
	"fmt"
	"io"
)

// ErrRequiredFlag define a type for flag error. Should be used to determine when to print command usage
type ErrRequiredFlag struct {
	Name string
}

func (rfe ErrRequiredFlag) Error() string {
	return fmt.Sprintf("-%s flag is required.", rfe.Name)
}

// Command define a cli command
type Command interface {
	// Flags is settings flags needed by the command. Caller still need to call flag.Parse() when ready.
	Flags()
	// Run execute the command, writing output to the writer, and reading inputs from the reader
	Run(w io.Writer, r io.Reader) error
}
