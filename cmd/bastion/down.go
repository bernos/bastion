package main

import (
	"os"

	"github.com/bernos/bastion/internal/commands"
	"github.com/bernos/bastion/internal/config"
	"github.com/bernos/bastion/internal/dependencies"
	"github.com/spf13/cobra"
)

func NewDownCommand(cfg *config.Config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Tear down a bastion host",
		Long:  "Delete the CloudFormation stack for a named bastion host and wait for deletion to complete",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			svc, err := dependencies.New(cfg).BastionService(ctx)
			if err != nil {
				return err
			}

			c := commands.NewDownCommand(svc, os.Stdin, os.Stdout, os.Stderr)

			return c.Run(ctx, &commands.DownInput{
				BastionName: cfg.Name,
			})
		},
	}

	cmd.Flags().String("name", "", "name of the bastion host to tear down")
	cmd.Flags().String("region", "", "AWS region where the bastion is deployed")

	for _, flag := range []string{"name", "region"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			return nil, err
		}
	}

	return cmd, nil
}
