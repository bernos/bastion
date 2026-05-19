package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bernos/bastion/internal/bastion"
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

			return runDown(cmd, cfg.Name, svc)
		},
	}

	cmd.Flags().String("name", "", "name of the bastion host to tear down")

	if err := cmd.MarkFlagRequired("name"); err != nil {
		return nil, err
	}

	return cmd, nil
}

func runDown(cmd *cobra.Command, name string, svc bastion.BastionService) error {
	if err := svc.DeleteBastion(cmd.Context(), &bastion.DeleteBastionInput{
		BastionName: name,
	}); err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(struct {
		StackName string `json:"stackName"`
	}{
		StackName: fmt.Sprintf("%s-stack", name),
	})
}
