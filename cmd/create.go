package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/kubesphere/ksbuilder/pkg/extension"
)

type inputPromptContent struct {
	text     string
	optional bool
	errorMsg string
}

type selectPromptContent struct {
	text              string
	items             []string
	startInSearchMode bool
}

type createOptions struct {
	from string
}

func createExtensionCmd() *cobra.Command {
	o := &createOptions{}

	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create a new KubeSphere extension",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(0),
		RunE:         o.run,
	}
	cmd.Flags().StringVar(&o.from, "from", "", "application helm chart file path of application class")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	extensionNamePrompt := inputPromptContent{
		text:     "Please input extension name",
		errorMsg: "Extension name can't be empty",
	}
	name := promptGetInput(extensionNamePrompt)

	categoryPromptContent := selectPromptContent{
		text:  fmt.Sprintf("What category does %s belong to?", name),
		items: extension.Categories,
	}
	category := promptGetSelect(categoryPromptContent)

	authorPrompt := inputPromptContent{
		text:     "Please input extension author",
		errorMsg: "Extension author can't be empty",
	}
	author := promptGetInput(authorPrompt)

	emailPrompt := inputPromptContent{
		text:     "Please input Email",
		optional: true,
	}
	email := promptGetInput(emailPrompt)

	urlPrompt := inputPromptContent{
		text:     "Please input author's URL",
		optional: true,
	}
	url := promptGetInput(urlPrompt)

	extensionConfig := extension.Config{
		Name:     name,
		Category: category,
		Author:   author,
		Email:    email,
		URL:      url,
	}

	pwd, _ := os.Getwd()
	p := path.Join(pwd, name)
	if err := extension.Create(p, extensionConfig); err != nil {
		return err
	}

	if o.from != "" {
		chartPath := path.Join(pwd, o.from)
		appChart, err := os.ReadFile(chartPath)
		if err != nil {
			return err
		}
		if err = extension.CreateAppChart(p, name, appChart); err != nil {
			return err
		}
	}

	fmt.Printf("Directory: %s\n\n", p)
	fmt.Println("The extension charts has been created.")

	return nil
}

var (
	bold  = promptui.Styler(promptui.FGBold)
	faint = promptui.Styler(promptui.FGFaint)
)

func promptGetInput(pc inputPromptContent) string {
	prompt := promptui.Prompt{
		Label: pc.text,
	}

	if pc.optional {
		prompt.Templates = &promptui.PromptTemplates{
			Valid:   fmt.Sprintf("%s {{ . | bold }} %s ", bold(promptui.IconGood), bold("(optional):")),
			Success: fmt.Sprintf("{{ . | faint }} %s ", faint("(optional):")),
		}
	} else {
		prompt.Validate = func(input string) error {
			if len(strings.TrimSpace(input)) <= 0 {
				return errors.New(pc.errorMsg)
			}
			return nil
		}
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(result)
}

func promptGetSelect(pc selectPromptContent) string {
	prompt := promptui.Select{
		Label: pc.text,
		Items: pc.items,
		Searcher: func(input string, index int) bool {
			return strings.Contains(strings.ToLower(pc.items[index]), strings.ToLower(input))
		},
		StartInSearchMode: pc.startInSearchMode,
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	return result
}
