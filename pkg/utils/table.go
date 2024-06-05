package utils

import (
	"os"
	"text/tabwriter"

	"github.com/jedib0t/go-pretty/v6/table"
)

func NewTableWriter() table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleDefault)
	t.Style().Options = table.Options{}
	t.Style().Box = table.BoxStyle{
		PaddingRight: "   ",
	}
	return t
}

func NewTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 10, 1, 3, ' ', 0)
}
