package options

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
)

type LintOptions struct {
	Client    *action.Lint
	ValueOpts *values.Options
	Settings  *cli.EnvSettings
}

func NewLintOptions() *LintOptions {
	o := &LintOptions{}
	o.Client = action.NewLint()
	o.ValueOpts = new(values.Options)
	o.Settings = cli.New()
	return o
}

func (o *LintOptions) AddFlags(cmd *cobra.Command, f *pflag.FlagSet) {
	// client flags
	cmd.Flags().BoolVar(&o.Client.Strict, "strict", false, "fail on lint warnings")
	cmd.Flags().BoolVar(&o.Client.WithSubcharts, "with-subcharts", false, "lint dependent charts")
	cmd.Flags().BoolVar(&o.Client.Quiet, "quiet", false, "print only warnings and errors")

	// value flags
	cmd.Flags().StringSliceVarP(&o.ValueOpts.ValueFiles, "values", "f", []string{}, "specify values in a YAML file or a URL (can specify multiple)")
	cmd.Flags().StringArrayVar(&o.ValueOpts.Values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.Flags().StringArrayVar(&o.ValueOpts.StringValues, "set-string", []string{}, "set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.Flags().StringArrayVar(&o.ValueOpts.FileValues, "set-file", []string{}, "set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)")
	cmd.Flags().StringArrayVar(&o.ValueOpts.JSONValues, "set-json", []string{}, "set JSON values on the command line (can specify multiple or separate values with commas: key1=jsonval1,key2=jsonval2)")
}
