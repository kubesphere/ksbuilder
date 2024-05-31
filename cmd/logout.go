package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kubesphere/ksbuilder/pkg/config"
)

type logoutOptions struct{}

func logoutCmd() *cobra.Command {
	o := logoutOptions{}

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out from KubeSphere Cloud",
		Args:  cobra.NoArgs,
		RunE:  o.logout,
	}
	return cmd
}

func (o *logoutOptions) logout(_ *cobra.Command, _ []string) error {
	if err := config.Remove(); err != nil {
		return err
	}
	fmt.Println("Logout Succeeded")
	return nil
}
