package main

import (
	"fmt"
	"os"

	"github.com/monkeymonk/gdt/internal/cli"
)

var Version = "dev"

func main() {
	app, err := cli.NewApp(Version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	root := cli.NewRootCmd(app)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
