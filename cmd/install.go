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
		Short:        "install an extension",
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

	installCmd := exec.Command("helm", "install", "--create-namespace", "-n", namespace, args[0], p)
	out, err := installCmd.CombinedOutput()

	if err != nil {
		fmt.Println(string(out))
		log.Fatalf("cmd.Run() failed with %s", err)
	}

	fmt.Println(string(out))

	return nil
}

func uninstallExtensionCmd() *cobra.Command {
	o := defaultPublishOptions()

	cmd := &cobra.Command{
		Use:          "uninstall",
		Short:        "uninstall an extension",
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

	uninstallCmd := exec.Command("helm", "uninstall", "-n", namespace, args[0])
	out, err := uninstallCmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		log.Fatalf("cmd.Run() failed with %s", err)
	}

	fmt.Println(string(out))

	return nil
}

func updateExtensionCmd() *cobra.Command {
	o := defaultPublishOptions()

	cmd := &cobra.Command{
		Use:          "update",
		Short:        "update a extension",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE:         o.update,
	}

	cmd.Flags().StringVarP(&o.namespace, "namespace", "n", o.namespace, "namespace")

	return cmd
}

func (o *installOptions) update(cmd *cobra.Command, args []string) error {
	fmt.Printf("update extension %s\n", args[0])

	pwd, _ := os.Getwd()
	p := path.Join(pwd, args[0])
	namespace := o.namespace

	uninstallCmd := exec.Command("helm", "upgrade", "-n", namespace, args[0], p)
	out, err := uninstallCmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		log.Fatalf("cmd.Run() failed with %s", err)
	}

	fmt.Println(string(out))

	return nil
}
