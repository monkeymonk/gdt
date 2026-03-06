package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLsRemoteCmd(app *App) *cobra.Command {
	var refresh bool

	cmd := &cobra.Command{
		Use:   "ls-remote",
		Short: "List available remote versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			releases, err := loadMetadata(app, refresh)
			if err != nil {
				return err
			}
			for _, r := range releases {
				label := ""
				if r.Stable {
					label = " stable"
				}
				fmt.Printf("%s%s\n", r.Version, label)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&refresh, "refresh", false, "Force refresh metadata cache")
	return cmd
}
