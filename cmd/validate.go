package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/kubesphere/ksbuilder/pkg/extension"
)

type validateOptions struct{}

func validateExtensionCmd() *cobra.Command {
	o := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "",
		Args:  cobra.ExactArgs(1),
		RunE:  o.validate,
	}
	return cmd
}

func (o *validateOptions) validate(_ *cobra.Command, args []string) error {
	pwd, _ := os.Getwd()
	p := path.Join(pwd, args[0])
	fmt.Printf("validating extension %s\n", args[0])

	if _, err := extension.Load(p); err != nil {
		return err
	}
	fmt.Println("\nno issues found")
	return nil
}
