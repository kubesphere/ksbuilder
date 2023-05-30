package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"

	"github.com/kubesphere/ksbuilder/pkg/extension"
)

type publishOptions struct {
	kubeconfig string
}

func defaultPublishOptions() *publishOptions {
	return &publishOptions{}
}

func publishExtensionCmd() *cobra.Command {
	o := defaultPublishOptions()

	cmd := &cobra.Command{
		Use:          "publish",
		Short:        "Publish an extension into the market",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE:         o.publish,
	}
	cmd.Flags().StringVar(&o.kubeconfig, "kubeconfig", "", "kubeconfig file path of the target cluster")
	return cmd
}

func (o *publishOptions) publish(cmd *cobra.Command, args []string) error {
	pwd, _ := os.Getwd()
	p := path.Join(pwd, args[0])
	fmt.Printf("publish extension %s\n", args[0])

	ext, err := extension.Load(p)
	if err != nil {
		return err
	}

	dir, err := os.MkdirTemp("", "publish")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir) // nolint

	filePath := path.Join(dir, "extension.yaml")
	if err = os.WriteFile(filePath, ext.ToKubernetesResources(), 0644); err != nil {
		return err
	}

	kubectlArgs := []string{"apply", "--server-side=true", "-f", filePath}
	if o.kubeconfig != "" {
		kubectlArgs = append(kubectlArgs, "--kubeconfig", o.kubeconfig)
	}
	command := exec.Command("kubectl", kubectlArgs...)

	out, err := command.CombinedOutput()
	fmt.Printf(string(out))

	return err
}
