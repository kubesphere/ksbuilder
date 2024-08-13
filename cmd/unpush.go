package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kubesphere/ksbuilder/pkg/cloud"
)

type unpushOptions struct{}

func unpushCmd() *cobra.Command {
	o := unpushOptions{}

	return &cobra.Command{
		Use:   "unpush",
		Short: "Unpush a snapshot of an extension",
		Args:  cobra.ExactArgs(1),
		RunE:  o.unpush,
	}
}

func (o *unpushOptions) unpush(_ *cobra.Command, args []string) error {
	snapshot := args[0]
	fmt.Printf("unpush snapshot %s\n", snapshot)

	client, err := cloud.NewClient()
	if err != nil {
		return fmt.Errorf("login failed: %v", err)
	}

	if err = client.CancelSubmitExtension(snapshot); err != nil {
		return err
	}
	fmt.Printf("Snapshot %s has been unsubmitted and reverted to draft state\n", snapshot)
	return nil
}
