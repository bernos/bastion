package main

import (
	"fmt"
	"os"

	"github.com/bernos/bastion/internal/config"
	"github.com/bernos/bastion/internal/connect"
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

			deps := dependencies.New(cfg)
			connectSvc, err := deps.ConnectService(ctx)
			if err != nil {
				return err
			}

			return runConnect(cmd, connectSvc, &connect.PrepareInput{
				BastionName:  cfg.Name,
				Region:       cfg.Region,
				ExtraSSHArgs: []string{"-D", fmt.Sprintf("%d", port), "-N"},
			}, func() {
				fmt.Fprintf(os.Stderr, "SOCKS5 proxy on localhost:%d via bastion %q — Ctrl-C to stop\n", port, cfg.Name)
			})
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
