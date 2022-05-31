package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/chenz24/ks-cli/pkg/project"
	"github.com/spf13/cobra"
)

type projectOptions struct {
	directory string
}

func defaultProjectOptions() *projectOptions {
	return &projectOptions{}
}

func newProjectCmd() *cobra.Command {
	o := defaultProjectOptions()

	cmd := &cobra.Command{
		Use:          "init",
		Short:        "Init a new project",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE:         o.run,
	}

	cmd.Flags().StringVarP(&o.directory, "directory", "d", o.directory, "directory")

	return cmd
}

func (o *projectOptions) run(cmd *cobra.Command, args []string) error {
	pwd, _ := os.Getwd()
	p := path.Join(pwd, args[0])
	//project.Create(p)
	if err := project.Create(p); err != nil {
		return err
	}

	fmt.Printf("Directory: %s\n\n", p)
	fmt.Println("The project has been created.")

	return nil
}
