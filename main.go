package main

import (
	"fmt"
	"os"

	"github.com/daeMOn63/gorrent/cmd"
)

var create, client, tracker cmd.Command

func main() {

	create = cmd.NewCreate()
	client = cmd.NewClient()
	tracker = cmd.NewTracker()

	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case create.FlagSet().Name():
		create.FlagSet().Parse(os.Args[2:])
	case client.FlagSet().Name():
		client.FlagSet().Parse(os.Args[2:])
	case tracker.FlagSet().Name():
		tracker.FlagSet().Parse(os.Args[2:])
	default:
		fmt.Printf("error: unknow command `%s`\n", os.Args[1])
		usage()
	}

	if create.FlagSet().Parsed() {
		err := create.Run(os.Stdout, os.Stdin)
		checkCmdError(create, err)
	} else if client.FlagSet().Parsed() {
		err := client.Run(os.Stdout, os.Stdin)
		checkCmdError(client, err)
	} else if tracker.FlagSet().Parsed() {
		err := tracker.Run(os.Stdout, os.Stdin)
		checkCmdError(tracker, err)
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
	fmt.Printf("client -todo\n")
	fmt.Printf("  Start a gorrent client.\n\n")
	fmt.Printf("tracker -todo\n")
	fmt.Printf("  Start a gorrent tracker.\n\n")
	fmt.Println()
	os.Exit(1)
}
