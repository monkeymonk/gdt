package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/monkeymonk/gdt/internal/download"
	"github.com/spf13/cobra"
)

func newSelfUpdateCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "self",
		Short: "Self management",
	}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update gdt to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSelfUpdate(app)
		},
	}

	cmd.AddCommand(updateCmd)
	return cmd
}

func runSelfUpdate(app *App) error {
	resp, err := http.Get("https://api.github.com/repos/monkeymonk/gdt/releases/latest")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &release); err != nil {
		return err
	}

	latestVersion := release.TagName
	if latestVersion == "v"+app.Version || latestVersion == app.Version {
		fmt.Fprintln(os.Stderr, "Already up to date")
		return nil
	}

	artifact := fmt.Sprintf("gdt-%s-%s-%s", latestVersion, runtime.GOOS, runtime.GOARCH)
	var downloadURL string
	for _, a := range release.Assets {
		if a.Name == artifact+".tar.gz" || a.Name == artifact+".zip" {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	tmpPath := exe + ".new"
	fmt.Fprintf(os.Stderr, "Updating to %s...\n", latestVersion)
	if err := download.File(downloadURL, tmpPath); err != nil {
		return err
	}

	os.Chmod(tmpPath, 0755)
	if err := os.Rename(tmpPath, exe); err != nil {
		return fmt.Errorf("failed to replace binary: %w\n\n  Try: sudo mv %s %s", tmpPath, exe)
	}

	fmt.Fprintf(os.Stderr, "Updated to %s\n", latestVersion)
	return nil
}
