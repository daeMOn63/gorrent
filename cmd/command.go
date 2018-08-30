package cmd

import (
	"flag"
	"fmt"
	"io"
)

// ErrRequiredFlag define a type for flag error. Should be used to determine when to print command usage
type ErrRequiredFlag struct {
	Name string
}

// Error returns the error message
func (rfe ErrRequiredFlag) Error() string {
	return fmt.Sprintf("-%s flag is required.", rfe.Name)
}

// Command define a cli command
type Command interface {
	// FlagSet returns the command flags as a flag.FlagSet
	FlagSet() *flag.FlagSet
	// Run execute the command, writing output to the writer, and reading inputs from the reader
	Run(w io.Writer, r io.Reader) error
}
