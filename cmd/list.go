package cmd

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/kubesphere/ksbuilder/pkg/cloud"
	"github.com/kubesphere/ksbuilder/pkg/utils"
)

type listOptions struct{}

func listCmd() *cobra.Command {
	o := listOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all extensions of the current user on KubeSphere Cloud",
		Args:  cobra.NoArgs,
		RunE:  o.list,
	}
	return cmd
}

func (o *listOptions) list(_ *cobra.Command, _ []string) error {
	client, err := cloud.NewClient()
	if err != nil {
		return fmt.Errorf("login failed: %v", err)
	}

	extensions, err := client.ListExtensions()
	if err != nil {
		return err
	}
	rows := make([]table.Row, 0)
	for _, extension := range extensions.Extensions {
		rows = append(rows, table.Row{
			extension.ExtensionID,
			extension.Name,
			extension.Status,
			extension.LatestVersion.Version,
		})
	}

	t := utils.NewTableWriter()
	t.AppendHeader(table.Row{"ID", "Name", "Status", "Latest version"})
	t.AppendRows(rows)
	t.Render()
	return nil
}
