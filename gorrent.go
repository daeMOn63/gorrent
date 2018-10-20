package main

import (
	"fmt"
	"os"

	"github.com/daeMOn63/gorrent/cmd"
)

var create, peerd, trackerd cmd.Command

func main() {

	create = cmd.NewCreate()
	peerd = cmd.NewPeerDaemon()
	trackerd = cmd.NewTrackerDaemon()

	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case create.FlagSet().Name():
		create.FlagSet().Parse(os.Args[2:])
	case peerd.FlagSet().Name():
		peerd.FlagSet().Parse(os.Args[2:])
	case trackerd.FlagSet().Name():
		trackerd.FlagSet().Parse(os.Args[2:])
	default:
		fmt.Printf("error: unknown command `%s`\n", os.Args[1])
		usage()
	}

	if create.FlagSet().Parsed() {
		err := create.Run(os.Stdout, os.Stdin)
		checkCmdError(create, err)
	} else if peerd.FlagSet().Parsed() {
		err := peerd.Run(os.Stdout, os.Stdin)
		checkCmdError(peerd, err)
	} else if trackerd.FlagSet().Parsed() {
		err := trackerd.Run(os.Stdout, os.Stdin)
		checkCmdError(trackerd, err)
	}
}

func checkCmdError(c cmd.Command, err error) {
	if err != nil {
		fmt.Printf("error:\n  %s.\n", err)

		if _, ok := err.(cmd.ErrRequiredFlag); ok {
			fmt.Println()
			c.FlagSet().Usage()
			os.Exit(1)
		}
	}
}

func usage() {
	fmt.Printf("\nUsage of %s:\n", os.Args[0])
	fmt.Printf("\t%s <subcommand> [flag...]\n", os.Args[0])
	fmt.Println()
	fmt.Printf("Available subcommands:\n\n")
	fmt.Printf("create -src <path> [-dst <path>] [-fsWorkers <num>] [-pieceLength <num>]\n")
	fmt.Printf("  Create a new gorrent file from a source file or directory.\n\n")
	fmt.Printf("peerd -public-ip <string> -storage <path> [-id <string]\n")
	fmt.Printf("      [-public-port <int>] [-sock <path>]\n")
	fmt.Printf("      [-read-timeout <num>] [-write-timeout <num>]\n")
	fmt.Printf("      [-db <path>]\n")
	fmt.Printf("  Start a gorrent peer daemon.\n\n")
	fmt.Printf("trackerd [-bind <ip>:<port>] [-read-timeout <num>] [-write-timeout <num>]\n")
	fmt.Printf("  Start a gorrent tracker daemon.\n\n")
	fmt.Println()
	os.Exit(1)
}
