package main

import (
	"fmt"
	"os"

	"github.com/monkeymonk/gdt/internal/cli"
	"github.com/monkeymonk/gdt/internal/shim"
	"github.com/monkeymonk/gdt/internal/versions"
)

var Version = "dev"

func main() {
	if shim.IsShimInvocation(os.Args[0]) {
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

	cwd, _ := os.Getwd()
	installed, _ := versions.List(app.VersionsDir())
	version, err := versions.Resolve(cwd, os.Getenv("GDT_GODOT_VERSION"), app.Config.DefaultVersion, installed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	binPath, err := versions.AbsoluteBinaryPath(app.VersionsDir(), version, app.Platform.OS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	if err := shim.Exec(binPath, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
