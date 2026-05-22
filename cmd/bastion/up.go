package main

import (
	"encoding/json"
	"os"

	"github.com/bernos/bastion/internal/bastion"
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

			out, err := svc.DeployBastion(ctx, &bastion.DeployBastionInput{
				BastionName: cfg.Name,
				Owner:       cfg.Owner,
				SubnetID:    cfg.SubnetID,
				VPCID:       cfg.VPCID,
				Region:      cfg.Region,
			})
			if err != nil {
				return err
			}

			return json.NewEncoder(os.Stdout).Encode(struct {
				StackName        string `json:"stackName"`
				InstanceID       string `json:"instanceId"`
				AvailabilityZone string `json:"availabilityZone"`
			}{
				StackName:        out.StackName,
				InstanceID:       out.InstanceID,
				AvailabilityZone: out.AvailabilityZone,
			})
		},
	}

	cmd.Flags().String("name", "", "name for the bastion host and associated resources")
	cmd.Flags().String("owner", "", "owner tag applied to bastion resources")
	cmd.Flags().String("subnet-id", "", "private subnet ID in which to launch the bastion instance")
	cmd.Flags().String("vpc-id", "", "VPC ID for the bastion security group")
	cmd.Flags().String("region", "", "AWS region to deploy the bastion into")

	for _, flag := range []string{"name", "owner", "subnet-id", "vpc-id", "region"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			return nil, err
		}
	}

	return cmd, nil
}
