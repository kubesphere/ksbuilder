package cmd

import (
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/kubesphere/ksbuilder/pkg/extension"
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
	client, err := cloud.NewClient()
	if err != nil {
		return fmt.Errorf("login failed: %v", err)
	}

	snapshot := args[0]
	if !o.likelySnapshotID(args[0]) {
		pwd, _ := os.Getwd()
		p := path.Join(pwd, args[0])
		ext, err := extension.Load(p)
		if err != nil {
			return err
		}
		snap, err := client.LocateExtensionSnapshot(ext.Metadata.Name, ext.Metadata.Version)
		if err != nil {
			return fmt.Errorf("failed to locate snapshot on kubesphere.cloud: %v\n", err)
		}
		fmt.Printf("Snapshot found on kubesphere.cloud\n")
		snapshot = snap.SnapshotID
	}
	fmt.Printf("unpush snapshot %s\n", snapshot)

	if err = client.CancelSubmitExtension(snapshot); err != nil {
		return err
	}
	fmt.Printf("Snapshot %s has been unsubmitted and reverted to draft state\n", snapshot)
	return nil
}

func (o *unpushOptions) likelySnapshotID(s string) bool {
	re := regexp.MustCompile(`^\d+$`)
	return re.MatchString(s)
}
