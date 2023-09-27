package cmd

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/iawia002/lia/kubernetes/client"
	"github.com/iawia002/lia/kubernetes/client/generic"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	"kubesphere.io/client-go/kubesphere/scheme"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

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
	return createOrApplyObjs(context.Background(), genericClient, ext.ToKubernetesResources()...)
}

func createOrApplyObjs(ctx context.Context, c runtimeclient.Client, objs ...runtimeclient.Object) error {
	for _, obj := range objs {
		key := runtimeclient.ObjectKeyFromObject(obj)
		newObj := obj.DeepCopyObject().(runtimeclient.Object)
		if err := c.Get(ctx, key, newObj); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
			fmt.Printf("creating %s %s\n", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
			if err = c.Create(ctx, obj); err != nil {
				return err
			}
			continue
		}

		fmt.Printf("updating %s %s\n", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
		if err := c.Patch(ctx, obj, runtimeclient.Apply, runtimeclient.ForceOwnership, runtimeclient.FieldOwner("ksbuilder")); err != nil {
			return err
		}
	}
	return nil
}
