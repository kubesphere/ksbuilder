package utils

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/clientcmd"
	"kubesphere.io/client-go/kubesphere/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func BuildClientFromFlags(kubeConfigPath string) (client.Client, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return client.New(restConfig, client.Options{
		Scheme: scheme.Scheme,
	})
}

func Apply(ctx context.Context, c client.Client, obj client.Object) error {
	key := client.ObjectKeyFromObject(obj)
	newObj := obj.DeepCopyObject().(client.Object)
	if err := c.Get(ctx, key, newObj); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		createOptions := []client.CreateOption{
			client.FieldOwner("ksbuilder"),
		}
		return c.Create(ctx, obj, createOptions...)
	}

	patchOptions := []client.PatchOption{
		client.FieldOwner("ksbuilder"),
		client.ForceOwnership,
	}
	return c.Patch(ctx, obj, client.Apply, patchOptions...)
}
