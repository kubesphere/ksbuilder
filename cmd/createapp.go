package cmd

import (
	"fmt"
	"github.com/kubesphere/ksbuilder/pkg/extension"
	"github.com/spf13/cobra"
)

type createAppOptions struct {
	from string
}

func createAppExtensionCmd() *cobra.Command {
	o := &createAppOptions{}

	cmd := &cobra.Command{
		Use:   "createapp",
		Short: "Create a new KubeSphere extension with a chart file to appstore",
		Args:  cobra.ExactArgs(0),
		RunE:  o.run,
	}
	cmd.Flags().StringVar(&o.from, "from", "", "helm chart file")

	return cmd
}

func (o *createAppOptions) run(_ *cobra.Command, _ []string) error {
	if o.from == "" {
		return fmt.Errorf("must specify a helm chart file with --from")
	}
	if err := extension.CreateApp(o.from); err != nil {
		return err
	}
	return nil
}
