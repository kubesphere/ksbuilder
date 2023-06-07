package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/kubesphere/ksbuilder/pkg/extension"
)

type packageOptions struct {
}

func defaultPackageOptions() *packageOptions {
	return &packageOptions{}
}

func packageExtensionCmd() *cobra.Command {
	o := defaultPackageOptions()

	cmd := &cobra.Command{
		Use:          "package",
		Short:        "package an extension",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE:         o.packageCmd,
	}
	return cmd
}

func (o *packageOptions) packageCmd(cmd *cobra.Command, args []string) error {
	pwd, _ := os.Getwd()
	p := path.Join(pwd, args[0])
	fmt.Printf("package extension %s\n", args[0])

	tempDir, err := os.MkdirTemp("", "chart")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir) // nolint

	if err = copy.Copy(p, tempDir); err != nil {
		return err
	}

	metadata, err := extension.LoadMetadata(p)
	if err != nil {
		return err
	}
	chartYaml, err := metadata.ToChartYaml()
	if err != nil {
		return err
	}
	chartMetadata, err := yaml.Marshal(chartYaml)
	if err != nil {
		return err
	}

	if err = os.WriteFile(tempDir+"/Chart.yaml", chartMetadata, 0644); err != nil {
		return err
	}

	ch, err := loader.LoadDir(tempDir)
	if err != nil {
		return err
	}
	chartFilename, err := chartutil.Save(ch, pwd)
	if err != nil {
		return err
	}
	fmt.Printf("package saved to %s\n", chartFilename)
	return nil
}
