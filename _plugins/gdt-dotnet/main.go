package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "doctor":
		os.Exit(cmdDoctor())
	case "restore":
		os.Exit(cmdRestore())
	case "build":
		os.Exit(cmdBuild())
	case "run":
		os.Exit(cmdRun())
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `gdt-dotnet - C# and .NET tooling for Godot projects

Usage: gdt-dotnet <command>

Commands:
  doctor   Check .NET SDK and toolchain status
  restore  Run dotnet restore in the project root
  build    Run dotnet build in the project root
  run      Run the project via dotnet run or Godot engine
  help     Show this help message

Environment:
  GDT_PROJECT_ROOT  Path to the Godot project root
  GDT_GODOT_VERSION Current Godot version
  GDT_ENGINE_PATH   Path to the Godot engine binary
  GDT_HOME          Path to gdt home directory`)
}

// cmdDoctor checks .NET SDK installation, version compatibility, and MSBuild.
func cmdDoctor() int {
	ok := true

	// Check dotnet SDK
	ver, err := runCapture("dotnet", "--version")
	if err != nil {
		fmt.Fprintln(os.Stderr, "[FAIL] .NET SDK not found. Install from https://dotnet.microsoft.com/download")
		ok = false
	} else {
		fmt.Fprintf(os.Stderr, "[ OK ] .NET SDK %s\n", strings.TrimSpace(ver))
	}

	// Check target framework compatibility
	if ver != "" {
		trimmed := strings.TrimSpace(ver)
		if strings.HasPrefix(trimmed, "6.") || strings.HasPrefix(trimmed, "8.") {
			fmt.Fprintf(os.Stderr, "[ OK ] SDK version compatible with Godot (net%s.0)\n", trimmed[:1])
		} else if strings.HasPrefix(trimmed, "9.") {
			fmt.Fprintln(os.Stderr, "[WARN] .NET 9 detected; Godot currently targets net6.0 or net8.0")
		} else {
			fmt.Fprintln(os.Stderr, "[WARN] Unexpected SDK version; Godot requires net6.0 or net8.0")
		}
	}

	// Check MSBuild via dotnet msbuild
	_, err = runCapture("dotnet", "msbuild", "--version")
	if err != nil {
		fmt.Fprintln(os.Stderr, "[FAIL] MSBuild not available via dotnet")
		ok = false
	} else {
		fmt.Fprintln(os.Stderr, "[ OK ] MSBuild available")
	}

	// Check project root for .csproj
	root := projectRoot()
	if root != "" {
		matches, _ := filepath.Glob(filepath.Join(root, "*.csproj"))
		if len(matches) > 0 {
			fmt.Fprintf(os.Stderr, "[ OK ] Found %d .csproj file(s) in project root\n", len(matches))
		} else {
			fmt.Fprintln(os.Stderr, "[WARN] No .csproj files found in project root")
		}
	} else {
		fmt.Fprintln(os.Stderr, "[WARN] GDT_PROJECT_ROOT not set")
	}

	if !ok {
		return 1
	}
	return 0
}

// cmdRestore runs dotnet restore in the project root.
func cmdRestore() int {
	root := requireProjectRoot()
	if root == "" {
		return 1
	}
	return runInDir(root, "dotnet", "restore")
}

// cmdBuild runs dotnet build in the project root.
func cmdBuild() int {
	root := requireProjectRoot()
	if root == "" {
		return 1
	}
	return runInDir(root, "dotnet", "build")
}

// cmdRun attempts dotnet run, falling back to the Godot engine.
func cmdRun() int {
	root := requireProjectRoot()
	if root == "" {
		return 1
	}

	// Try dotnet run first
	matches, _ := filepath.Glob(filepath.Join(root, "*.csproj"))
	if len(matches) > 0 {
		fmt.Fprintln(os.Stderr, "Running via dotnet run...")
		code := runInDir(root, "dotnet", "run")
		if code == 0 {
			return 0
		}
		fmt.Fprintln(os.Stderr, "dotnet run failed, attempting Godot engine fallback...")
	}

	// Fallback to Godot engine
	enginePath := os.Getenv("GDT_ENGINE_PATH")
	if enginePath == "" {
		fmt.Fprintln(os.Stderr, "error: GDT_ENGINE_PATH not set and dotnet run unavailable")
		return 1
	}

	fmt.Fprintf(os.Stderr, "Running via Godot engine: %s\n", enginePath)
	return runInDir(root, enginePath, "--path", root)
}

// projectRoot returns GDT_PROJECT_ROOT or empty string.
func projectRoot() string {
	return os.Getenv("GDT_PROJECT_ROOT")
}

// requireProjectRoot returns the project root or prints an error.
func requireProjectRoot() string {
	root := projectRoot()
	if root == "" {
		fmt.Fprintln(os.Stderr, "error: GDT_PROJECT_ROOT environment variable not set")
		return ""
	}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error: project root does not exist: %s\n", root)
		return ""
	}
	return root
}

// runCapture runs a command and returns its combined output.
func runCapture(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// runInDir executes a command in the given directory, forwarding stdio.
func runInDir(dir, name string, args ...string) int {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}
