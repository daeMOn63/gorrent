package main

import (
	"flag"
	"fmt"
	"gorrent/cmd"
	"os"
)

func main() {

	create := cmd.NewCreate()
	create.Flags()

	flag.Parse()

	err := create.Run(os.Stdout, os.Stdin)
	if err != nil {
		fmt.Printf("error:\n\t%s.\n", err)

		if _, ok := err.(cmd.ErrRequiredFlag); ok {
			fmt.Println()
			flag.Usage()
		}
		os.Exit(1)
	}
}
