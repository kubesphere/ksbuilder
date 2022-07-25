package cmd

import (
	"errors"
	"fmt"
	"github.com/chenz24/ksbuilder/pkg/plugin"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"os"
	"path"
)

type promptContent struct {
	errorMsg string
	label    string
}

func newPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create a new plugin",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(0),
		RunE:         run,
	}

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	pluginNamePrompt := promptContent{
		errorMsg: "plugin name can't be empty",
		label:    "Please input plugin name: ",
	}
	name := promptGetInput(pluginNamePrompt)

	pluginDescPrompt := promptContent{
		errorMsg: "plugin description can't be empty",
		label:    "Please input plugin description: ",
	}
	desc := promptGetInput(pluginDescPrompt)

	categoryPromptContent := promptContent{
		"Please provide a category.",
		fmt.Sprintf("What category does %s belong to?", name),
	}
	category := promptGetSelect(categoryPromptContent)

	pluginAuthorPrompt := promptContent{
		errorMsg: "plugin author can't be empty",
		label:    "Please input plugin author: ",
	}
	author := promptGetInput(pluginAuthorPrompt)

	pluginEmailPrompt := promptContent{
		errorMsg: "Email can't be empty",
		label:    "Please input Email: ",
	}
	email := promptGetInput(pluginEmailPrompt)

	pluginConfig := plugin.Config{
		Name:     name,
		Desc:     desc,
		Category: category,
		Author:   author,
		Email:    email,
	}

	pwd, _ := os.Getwd()
	p := path.Join(pwd, name)
	if err := plugin.Create(p, pluginConfig); err != nil {
		return err
	}

	fmt.Printf("Directory: %s\n\n", p)
	fmt.Println("The plugin charts has been created.")

	return nil
}

func promptGetInput(pc promptContent) string {
	validate := func(input string) error {
		if len(input) <= 0 {
			return errors.New(pc.errorMsg)
		}
		return nil
	}

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     pc.label,
		Templates: templates,
		Validate:  validate,
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Input: %s\n", result)

	return result
}

func promptGetSelect(pc promptContent) string {
	items := []string{"Performance", "Monitoring", "Logging", "Messaging"}
	index := -1
	var result string
	var err error

	for index < 0 {
		prompt := promptui.SelectWithAdd{
			Label:    pc.label,
			Items:    items,
			AddLabel: "Other",
		}

		index, result, err = prompt.Run()

		if index == -1 {
			items = append(items, result)
		}
	}

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Input: %s\n", result)

	return result
}
