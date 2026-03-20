package main

import (
	"fmt"
	"os"

	"github.com/monkeymonk/gdt/internal/cli"
	"github.com/monkeymonk/gdt/internal/engine"
)

var Version = "dev"

func main() {
	if engine.IsShimInvocation(os.Args[0]) {
		runShim()
		return
	}

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

func runShim() {
	app, err := cli.NewApp(Version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	svc := engine.NewService(app.Home, app.Platform, app.Config)
	cwd, _ := os.Getwd()
	resolved, err := svc.Resolve(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	if err := engine.ExecShim(resolved.BinaryPath, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
