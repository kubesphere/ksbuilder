package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ksnext",
		Short: "ksnext is a command line interface for KubeSphere plugin system",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(cmd.UsageString())

			return nil
		},
	}

	cmd.AddCommand(newVersionCmd(version)) // version subcommand
	cmd.AddCommand(newProjectCmd())        // init_project subcommand
	cmd.AddCommand(newPluginCmd())         // create_plugin subcommand
	cmd.AddCommand(installPluginCmd())     // publish_plugin subcommand
	cmd.AddCommand(uninstallPluginCmd())   // uninstall_plugin subcommand
	cmd.AddCommand(upgradePluginCmd())     // upgrade_plugin subcommand

	return cmd
}

// Execute invokes the command.
func Execute(version string) error {
	if err := newRootCmd(version).Execute(); err != nil {
		return fmt.Errorf("error executing root command: %w", err)
	}

	return nil
}
