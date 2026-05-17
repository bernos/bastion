package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bernos/bastion/internal/bastion"
	"github.com/bernos/bastion/internal/config"
	"github.com/bernos/bastion/internal/dependencies"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func NewUpCommand(cfg *config.Config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Deploy a bastion host",
		Long:  "Deploy a bastion host via CloudFormation",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var publicKeyContent string
			if cfg.PublicKeyPath != "" {
				data, err := os.ReadFile(cfg.PublicKeyPath)
				if err != nil {
					return fmt.Errorf("reading public key file: %w", err)
				}
				if strings.Contains(string(data), "PRIVATE KEY") {
					return fmt.Errorf("--public-key-path appears to be a private key; provide the public key (.pub) file instead")
				}
				if _, _, _, _, err := ssh.ParseAuthorizedKey(data); err != nil {
					return fmt.Errorf("--public-key-path does not contain a valid SSH public key: %w", err)
				}
				publicKeyContent = string(data)
			}

			deps := dependencies.New(cfg)

			svc, err := deps.BastionService(ctx)
			if err != nil {
				return err
			}

			out, err := svc.DeployBastion(ctx, &bastion.DeployBastionInput{
				BastionName:      cfg.Name,
				Owner:            cfg.Owner,
				SubnetID:         cfg.SubnetID,
				VPCID:            cfg.VPCID,
				PublicKeyContent: publicKeyContent,
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
	cmd.Flags().String("public-key-path", "", "path to SSH public key to upload via EC2 Instance Connect (key expires after 60 seconds)")

	for _, flag := range []string{"name", "owner", "subnet-id", "vpc-id"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			return nil, err
		}
	}

	return cmd, nil
}
