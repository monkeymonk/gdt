package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/spf13/cobra"
)

func newTemplatesCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "Manage export templates",
	}

	cmd.AddCommand(newTemplatesInstallCmd(app), newTemplatesListCmd(app), newTemplatesRemoveCmd(app))
	return cmd
}

func newTemplatesListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := engine.NewService(app.Home, app.Platform, app.Config)
			list, err := svc.ListTemplates()
			if err != nil {
				return err
			}
			if len(list) == 0 {
				fmt.Fprintln(os.Stderr, "No templates installed")
				fmt.Fprintln(os.Stderr, "\n  gdt templates install <version>")
				return nil
			}
			fmt.Println("Installed templates")
			for _, t := range list {
				fmt.Printf("  %s\n", t)
			}
			return nil
		},
	}
}

func newTemplatesRemoveCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "remove [version]",
		Aliases: []string{"rm"},
		Short:   "Remove installed export templates",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := ""
			if len(args) > 0 {
				version = args[0]
			}
			if version == "" && isTTY() {
				v, err := promptInstalledTemplate(app, "Templates to remove")
				if err != nil {
					return err
				}
				version = v
			}
			if version == "" {
				return fmt.Errorf("version required\n\n  gdt templates remove <version>")
			}
			if isTTY() {
				ok, err := promptConfirm(fmt.Sprintf("Remove templates for %s?", version))
				if err != nil {
					return err
				}
				if !ok {
					fmt.Fprintln(os.Stderr, "Aborted")
					return nil
				}
			}

			svc := engine.NewService(app.Home, app.Platform, app.Config)
			if err := svc.RemoveTemplates(version); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "Templates for %s removed\n", version)
			return nil
		},
	}
}

func newTemplatesInstallCmd(app *App) *cobra.Command {
	var mono bool
	var refresh bool

	cmd := &cobra.Command{
		Use:   "install [version]",
		Short: "Install export templates",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := ""
			if len(args) > 0 {
				query = args[0]
			}
			if query == "" && isTTY() {
				v, err := promptVersion(app, "Install templates for version")
				if err != nil {
					return err
				}
				query = v
			}
			if query == "" {
				return fmt.Errorf("version required\n\n  gdt templates install <version>")
			}

			svc := engine.NewService(app.Home, app.Platform, app.Config)
			fmt.Fprintf(os.Stderr, "Installing templates for %s...\n", query)
			result, err := svc.InstallTemplates(cmd.Context(), query, engine.InstallOpts{
				Mono:    mono,
				Refresh: refresh,
			})
			if errors.Is(err, engine.ErrAlreadyInstalled) {
				fmt.Fprintf(os.Stderr, "Templates for %s already installed\n", result.VersionName)
				return nil
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Templates for %s installed\n", result.VersionName)
			return nil
		},
	}

	cmd.Flags().BoolVar(&mono, "mono", false, "Install Mono templates")
	cmd.Flags().BoolVar(&refresh, "refresh", false, "Refresh metadata cache")
	return cmd
}
