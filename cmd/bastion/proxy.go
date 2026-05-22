package main

import (
	"fmt"
	"os"

	ec2ic "github.com/aws/aws-sdk-go-v2/service/ec2instanceconnect"
	"github.com/bernos/bastion/internal/config"
	"github.com/bernos/bastion/internal/dependencies"
	"github.com/spf13/cobra"
)

func NewProxyCommand(cfg *config.Config) (*cobra.Command, error) {
	var port int

	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Start a local SOCKS5 proxy through a bastion host",
		Long:  "Upload an ephemeral key and start a local SOCKS5 proxy via SSH -D through SSM Session Manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := checkSSMDependencies(); err != nil {
				return err
			}

			deps := dependencies.New(cfg)
			svc, err := deps.BastionService(ctx)
			if err != nil {
				return err
			}
			awsCfg, err := deps.AwsConfig(ctx)
			if err != nil {
				return err
			}

			extraArgs := []string{"-D", fmt.Sprintf("%d", port), "-N"}
			onReady := func() {
				fmt.Fprintf(os.Stderr, "SOCKS5 proxy on localhost:%d via bastion %q — Ctrl-C to stop\n", port, cfg.Name)
			}
			return runConnect(cmd, cfg.Name, cfg.Region, svc, ec2ic.NewFromConfig(awsCfg), extraArgs, onReady)
		},
	}

	cmd.Flags().String("name", "", "name of the bastion host to proxy through")
	cmd.Flags().String("region", "", "AWS region where the bastion is deployed")
	cmd.Flags().IntVar(&port, "port", 1080, "local port to listen on for SOCKS5 connections")

	for _, flag := range []string{"name", "region"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			return nil, err
		}
	}

	return cmd, nil
}
