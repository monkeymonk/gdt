package cli

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/monkeymonk/gdt/internal/proxy"
	"github.com/spf13/cobra"
)

func newLspCmd(app *App) *cobra.Command {
	var port int
	var projectPath string

	cmd := &cobra.Command{
		Use:   "lsp",
		Short: "Start LSP proxy (stdin/stdout to Godot LSP)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLanguageProxy(app, port, "--lsp-port", projectPath)
		},
	}

	cmd.Flags().IntVar(&port, "port", 6005, "Godot LSP TCP port")
	cmd.Flags().StringVarP(&projectPath, "path", "C", "", "Path to Godot project directory")
	return cmd
}

func runLanguageProxy(app *App, port int, portFlag string, projectPath string) error {
	if projectPath != "" {
		if err := os.Chdir(projectPath); err != nil {
			return fmt.Errorf("cannot change to project directory: %w", err)
		}
	}

	_, _, binPath, err := resolveProjectVersion(app)
	if err != nil {
		return err
	}

	godotCmd := exec.Command(binPath, "--headless", portFlag, strconv.Itoa(port))
	godotCmd.Stderr = os.Stderr
	if err := godotCmd.Start(); err != nil {
		return fmt.Errorf("failed to start Godot: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		godotCmd.Process.Kill()
	}()

	defer func() {
		godotCmd.Process.Kill()
		godotCmd.Wait()
	}()

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	return proxy.Bridge(addr, os.Stdin, os.Stdout)
}
