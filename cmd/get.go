package cmd

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/kubesphere/ksbuilder/pkg/cloud"
	"github.com/kubesphere/ksbuilder/pkg/utils"
)

type getOptions struct{}

func getCmd() *cobra.Command {
	o := getOptions{}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get the summary and snapshots list of the extension on KubeSphere Cloud",
		Args:  cobra.ExactArgs(1),
		RunE:  o.get,
	}
	return cmd
}

func (o *getOptions) get(_ *cobra.Command, args []string) error {
	client, err := cloud.NewClient()
	if err != nil {
		return fmt.Errorf("login failed: %v", err)
	}

	extensionName := args[0]
	extension, err := client.GetExtension(extensionName)
	if err != nil {
		return err
	}
	snapshots, err := client.ListExtensionSnapshots(extensionName)
	if err != nil {
		return err
	}

	tabWriter := utils.NewTabWriter()
	tabWriter.Write([]byte(fmt.Sprintf("Name:\t%s\n", extensionName)))       // nolint
	tabWriter.Write([]byte(fmt.Sprintf("ID:\t%s\n", extension.ExtensionID))) // nolint
	tabWriter.Write([]byte(fmt.Sprintf("Status:\t%s\n", extension.Status)))  // nolint
	if extension.Status == "ready" {
		tabWriter.Write([]byte(fmt.Sprintf("Latest version:\t%s\n", extension.LatestVersion.Version))) // nolint
	}
	tabWriter.Write([]byte("\n")) // nolint
	tabWriter.Flush()             // nolint

	rows := make([]table.Row, 0)
	for _, snapshot := range snapshots {
		rows = append(rows, table.Row{
			snapshot.SnapshotID,
			snapshot.Metadata.Version,
			snapshot.Status,
			snapshot.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	t := utils.NewTableWriter()
	t.AppendHeader(table.Row{"Snapshot ID", "Version", "Status", "Update time"})
	t.AppendRows(rows)
	t.Render()
	return nil
}
