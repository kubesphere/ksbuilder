package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/iawia002/lia/kubernetes/client"
	"github.com/iawia002/lia/kubernetes/client/generic"
	"github.com/spf13/cobra"
	"kubesphere.io/client-go/kubesphere/scheme"

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
		Use:   "publish",
		Short: "Publish an extension into the market",
		Args:  cobra.ExactArgs(1),
		RunE:  o.publish,
	}
	cmd.Flags().StringVar(&o.kubeconfig, "kubeconfig", "", "kubeconfig file path of the target cluster")
	return cmd
}

func (o *publishOptions) publish(_ *cobra.Command, args []string) error {
	pwd, _ := os.Getwd()
	p := path.Join(pwd, args[0])
	fmt.Printf("publish extension %s\n", args[0])

	if o.kubeconfig == "" {
		homeDir, _ := os.UserHomeDir()
		o.kubeconfig = fmt.Sprintf("%s/.kube/config", homeDir)
	}
	config, err := client.BuildConfigFromFlags("", o.kubeconfig, client.SetQPS(25, 50))
	if err != nil {
		return err
	}
	genericClient, err := generic.NewClient(config, generic.WithScheme(scheme.Scheme), generic.WithCacheReader(false))
	if err != nil {
		return err
	}

	ext, err := extension.Load(p)
	if err != nil {
		return err
	}
	for _, obj := range ext.ToKubernetesResources() {
		fmt.Printf("creating %s %s\n", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
		if err = client.Apply(context.Background(), genericClient, obj, client.WithFieldManager("ksbuilder")); err != nil {
			return err
		}
	}
	return nil
}
