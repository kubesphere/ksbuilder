package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func versionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "Display version",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", version)
		},
	}
}
