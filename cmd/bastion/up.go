package main

import (
	"os"

	"github.com/bernos/bastion/cmd/bastion/flags"
	"github.com/bernos/bastion/internal/commands"
	"github.com/bernos/bastion/internal/config"
	"github.com/bernos/bastion/internal/dependencies"
	"github.com/spf13/cobra"
)

func NewUpCommand(cfg *config.Config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Deploy a bastion host",
		Long:  "Deploy a bastion host via CloudFormation",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			deps := dependencies.New(cfg)

			svc, err := deps.BastionService(ctx)
			if err != nil {
				return err
			}

			c := commands.NewUpCommand(svc, os.Stdin, os.Stdout, os.Stderr)

			return c.Run(ctx, &commands.UpInput{
				BastionName:      cfg.Name,
				Owner:            cfg.Owner,
				SubnetID:         cfg.SubnetID,
				VPCID:            cfg.VPCID,
				AMIParameterName: cfg.AMIParameterStoreParamName,
				Tags:             cfg.Tags,
			})
		},
	}

	cmd.Flags().String("name", "", "name for the bastion host and associated resources")
	cmd.Flags().String("owner", "", "owner tag applied to bastion resources")
	cmd.Flags().String("subnet-id", "", "private subnet ID in which to launch the bastion instance")
	cmd.Flags().String("vpc-id", "", "VPC ID for the bastion security group")
	cmd.Flags().String("region", "", "AWS region to deploy the bastion into")
	cmd.Flags().String("ami-parameter-store-param-name", "", "SSM parameter store path for the bastion AMI ID")

	tagsVar := flags.TagMap{}
	cmd.Flags().Var(&tagsVar, "tags", "additional tags to apply to the CloudFormation stack (key=value,...); max 50 tags")

	for _, flag := range []string{"name", "owner", "subnet-id", "vpc-id", "region"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			return nil, err
		}
	}

	return cmd, nil
}
