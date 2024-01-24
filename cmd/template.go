package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/registry"

	"github.com/kubesphere/ksbuilder/cmd/options"
	"github.com/kubesphere/ksbuilder/pkg/extension"
)

func templateExtensionCmd() *cobra.Command {
	o := options.NewTemplateOptions()
	cmd := &cobra.Command{
		Use:   "template PATH [flags]",
		Short: "locally render template for an extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registryClient, err := newRegistryClient(o)
			if err != nil {
				return fmt.Errorf("missing registry client: %w", err)
			}
			o.Client.SetRegistryClient(registryClient)

			// This is for the case where "" is specifically passed in as a
			// value. When there is no value passed in NoOptDefVal will be used
			// and it is set to client. See addInstallFlags.
			if o.Client.DryRunOption == "" {
				o.Client.DryRunOption = "true"
			}
			o.Client.DryRun = true
			o.Client.ReleaseName = "release-name"
			o.Client.Replace = true // Skip the name check
			o.Client.ClientOnly = !o.Validate
			o.Client.APIVersions = chartutil.VersionSet(o.ExtraAPIs)
			o.Client.IncludeCRDs = o.IncludeCrds

			return extension.PrintTemplate(args, o, os.Stdout)
		},
	}

	f := cmd.Flags()
	o.AddFlags(cmd, f)
	return cmd
}

func newRegistryClient(o *options.TemplateOptions) (*registry.Client, error) {
	if o.Client.CertFile != "" && o.Client.KeyFile != "" || o.Client.CaFile != "" || o.Client.InsecureSkipTLSverify {
		registryClient, err := newRegistryClientWithTLS(o)
		if err != nil {
			return nil, err
		}
		return registryClient, nil
	}
	registryClient, err := newDefaultRegistryClient(o)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}

func newDefaultRegistryClient(o *options.TemplateOptions) (*registry.Client, error) {
	opts := []registry.ClientOption{
		registry.ClientOptDebug(o.Settings.Debug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stderr),
		registry.ClientOptCredentialsFile(o.Settings.RegistryConfig),
	}
	if o.Client.PlainHTTP {
		opts = append(opts, registry.ClientOptPlainHTTP())
	}

	// Create a new registry client
	registryClient, err := registry.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}

func newRegistryClientWithTLS(o *options.TemplateOptions) (*registry.Client, error) {
	// Create a new registry client
	registryClient, err := registry.NewRegistryClientWithTLS(os.Stderr, o.Client.CertFile, o.Client.KeyFile, o.Client.CaFile, o.Client.InsecureSkipTLSverify,
		o.Settings.RegistryConfig, o.Settings.Debug,
	)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}
