package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/bernos/bastion/internal/config"
	"github.com/bernos/bastion/internal/dependencies"
	"github.com/spf13/cobra"
)

func NewUpCommand(cfg *config.Config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "up command short",
		Long:  "up command long",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("This is the up command")
			fmt.Printf("name: %s\n", cfg.Name)

			ctx := cmd.Context()
			deps := dependencies.New(cfg)

			stsClient, err := deps.STSClient(ctx)
			if err != nil {
				return err
			}

			output, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
			if err != nil {
				return err
			}

			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return err
			}

			fmt.Printf("%s", data)

			return nil
		},
	}

	cmd.Flags().String("name", "", "name for the bastion host and associated resources")
	cmd.Flags().String("subnet-id", "", "private subnet ID in which to launch the bastion instance")
	cmd.Flags().String("vpc-id", "", "VPC ID for the bastion security group")

	if err := cmd.MarkFlagRequired("name"); err != nil {
		return nil, err
	}
	if err := cmd.MarkFlagRequired("subnet-id"); err != nil {
		return nil, err
	}
	if err := cmd.MarkFlagRequired("vpc-id"); err != nil {
		return nil, err
	}

	return cmd, nil
}
