package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kubesphere/ksbuilder/cmd/options"
	"github.com/kubesphere/ksbuilder/pkg/extension"
)

func lintExtensionCmd() *cobra.Command {
	o := options.NewLintOptions()

	cmd := &cobra.Command{
		Use:        "lint PATH [flags]",
		Aliases:    nil,
		SuggestFor: nil,
		Args:       cobra.MinimumNArgs(1),
		Short: "This command takes a path to a chart and runs a series of tests to verify that\n" +
			"the chart is well-formed.",
		Long: "If the linter encounters things that will cause the chart to fail installation,\n" +
			"it will emit [ERROR] messages. If it encounters issues that break with convention\n" +
			"or recommendation, it will emit [WARNING] messages.",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := []string{"."}
			if len(args) > 0 {
				paths = args
			}

			// when helm lint is error. continue run builtins lint
			_ = extension.WithHelm(o, paths)

			if err := extension.WithBuiltins(o, paths); err != nil {
				return err
			}

			return nil
		},
	}

	o.AddFlags(cmd, cmd.Flags())
	return cmd
}
