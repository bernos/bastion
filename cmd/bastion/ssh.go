package main

import (
	"os"

	"github.com/bernos/bastion/internal/commands"
	"github.com/bernos/bastion/internal/config"
	"github.com/bernos/bastion/internal/dependencies"
	"github.com/spf13/cobra"
)

func NewSSHCommand(cfg *config.Config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "Open an interactive SSH session to a bastion host",
		Long:  "Upload an ephemeral key and open an interactive SSH session via SSM Session Manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			deps := dependencies.New(cfg)
			connectSvc, err := deps.ConnectService(ctx)
			if err != nil {
				return err
			}

			c := commands.NewConnectCommand(connectSvc, os.Stdin, os.Stdout, os.Stderr)

			return c.Run(ctx, &commands.ConnectInput{
				BastionName: cfg.Name,
				Region:      cfg.Region,
			})
		},
	}

	cmd.Flags().String("name", "", "name of the bastion host to connect to")
	cmd.Flags().String("region", "", "AWS region where the bastion is deployed")

	for _, flag := range []string{"name", "region"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			return nil, err
		}
	}

	return cmd, nil
}
