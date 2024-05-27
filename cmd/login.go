package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/kubesphere/ksbuilder/pkg/cloud"
	"github.com/kubesphere/ksbuilder/pkg/config"
)

type loginOptions struct {
	token  string
	server string
}

func loginCmd() *cobra.Command {
	o := loginOptions{}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to KubeSphere Cloud",
		Args:  cobra.NoArgs,
		RunE:  o.login,
	}
	cmd.Flags().StringVarP(&o.token, "token", "t", "", "API access token")
	cmd.Flags().StringVar(&o.server, "server", "https://apis.kubesphere.cloud", "API server address")
	return cmd
}

func (o *loginOptions) login(_ *cobra.Command, _ []string) error {
	if o.token == "" {
		prompt := promptui.Prompt{
			Label: "Enter API token",
			Mask:  '*',
			Validate: func(input string) error {
				if len(strings.TrimSpace(input)) <= 0 {
					return errors.New("token can't be empty")
				}
				return nil
			},
		}
		result, err := prompt.Run()
		if err != nil {
			return err
		}
		o.token = result
	}

	if _, err := cloud.NewClient(cloud.WithToken(o.token), cloud.WithServer(o.server)); err != nil {
		return fmt.Errorf("login failed: %v", err)
	}
	if err := config.Write([]byte(o.token), o.server); err != nil {
		return err
	}
	fmt.Println("Login Succeeded")
	return nil
}
