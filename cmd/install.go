package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"path"
)

type installOptions struct {
	namespace string
}

func defaultPublishOptions() *installOptions {
	return &installOptions{
		namespace: "extension-default",
	}
}

func installExtensionCmd() *cobra.Command {
	o := defaultPublishOptions()

	cmd := &cobra.Command{
		Use:          "install",
		Short:        "install a extension",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE:         o.publish,
	}

	cmd.Flags().StringVarP(&o.namespace, "namespace", "n", o.namespace, "namespace")

	return cmd
}

func (o *installOptions) publish(cmd *cobra.Command, args []string) error {
	pwd, _ := os.Getwd()
	p := path.Join(pwd, args[0])
	fmt.Printf("install extension %s\n", args[0])

	namespace := o.namespace

	out, err := exec.Command("helm", "install", "--create-namespace", "-n", namespace, args[0], p).Output()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(out))

	return nil
}

func uninstallExtensionCmd() *cobra.Command {
	o := defaultPublishOptions()

	cmd := &cobra.Command{
		Use:          "uninstall",
		Short:        "uninstall a extension",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE:         o.uninstall,
	}

	cmd.Flags().StringVarP(&o.namespace, "namespace", "n", o.namespace, "namespace")

	return cmd
}

func (o *installOptions) uninstall(cmd *cobra.Command, args []string) error {
	fmt.Printf("uninstall extension %s\n", args[0])

	namespace := o.namespace

	out, err := exec.Command("helm", "uninstall", "-n", namespace, args[0]).Output()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(out))

	return nil
}

func upgradeExtensionCmd() *cobra.Command {
	o := defaultPublishOptions()

	cmd := &cobra.Command{
		Use:          "upgrade",
		Short:        "upgrade a extension",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE:         o.upgrade,
	}

	cmd.Flags().StringVarP(&o.namespace, "namespace", "n", o.namespace, "namespace")

	return cmd
}

func (o *installOptions) upgrade(cmd *cobra.Command, args []string) error {
	fmt.Printf("upgrade extension %s\n", args[0])

	pwd, _ := os.Getwd()
	p := path.Join(pwd, args[0])
	namespace := o.namespace

	out, err := exec.Command("helm", "upgrade", "-n", namespace, args[0], p).Output()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(out))

	return nil
}
