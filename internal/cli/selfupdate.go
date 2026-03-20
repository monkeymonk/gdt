package cli

import (
	"context"
	"fmt"

	"github.com/monkeymonk/gdt/internal/selfupdate"
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
			result, err := selfupdate.Update(context.Background(), app.Version)
			if err != nil {
				return err
			}
			if result.Updated {
				fmt.Printf("  updated gdt to %s\n", result.NewVersion)
			} else {
				fmt.Println("  already up to date")
			}
			return nil
		},
	}

	cmd.AddCommand(updateCmd)
	return cmd
}
