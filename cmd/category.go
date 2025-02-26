package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func categoryCmd(categories []Category) *cobra.Command {
	return &cobra.Command{
		Use:   "category",
		Short: "List supported extension categories, use the normalized name in extension.yaml.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			for _, c := range categories {
				fmt.Printf("%-30s (Normalized name: %s)\n", c.DisplayNameEN, c.NormalizedName)
			}
		},
	}
}
