package cmd

import (
	"fmt"
	"github.com/kubesphere/ksbuilder/pkg/extension"
	"github.com/spf13/cobra"
)

type createSimpleOptions struct {
	from string
}

func createSimpleExtensionCmd() *cobra.Command {
	o := &createSimpleOptions{}

	cmd := &cobra.Command{
		Use:   "createsimple",
		Short: "Create a new KubeSphere extension with a chart file",
		Args:  cobra.ExactArgs(0),
		RunE:  o.run,
	}
	cmd.Flags().StringVar(&o.from, "from", "", "helm chart file")

	return cmd
}

func (o *createSimpleOptions) run(_ *cobra.Command, _ []string) error {
	if o.from == "" {
		return fmt.Errorf("must specify a helm chart file with --from")
	}
	if err := extension.CreateSimple(o.from); err != nil {
		return err
	}
	return nil
}
