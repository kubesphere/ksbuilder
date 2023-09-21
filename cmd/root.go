package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ksbuilder",
		Short:         "ksbuilder is a command line interface for KubeSphere extension system",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(cmd.UsageString())
		},
	}

	cmd.AddCommand(versionCmd(version))
	cmd.AddCommand(createExtensionCmd())
	cmd.AddCommand(publishExtensionCmd())
	cmd.AddCommand(packageExtensionCmd())
	cmd.AddCommand(unpublishExtensionCmd())

	return cmd
}

// Execute invokes the command.
func Execute(version string) error {
	if err := NewRootCmd(version).Execute(); err != nil {
		return fmt.Errorf("error executing command: %+v", err)
	}
	return nil
}
