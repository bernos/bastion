package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/bernos/bastion/internal/config"
	"github.com/bernos/bastion/internal/connect"
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

			return runConnect(cmd, connectSvc, &connect.PrepareInput{
				BastionName: cfg.Name,
				Region:      cfg.Region,
			}, nil)
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

// runConnect is the shared exec flow for both ssh and proxy commands.
// It calls CheckDependencies, Prepare, then execs ssh with the returned args.
func runConnect(cmd *cobra.Command, connectSvc connect.ConnectService, input *connect.PrepareInput, onReady func()) error {
	ctx := cmd.Context()

	if err := connectSvc.CheckDependencies(); err != nil {
		return err
	}

	keyFile, err := os.CreateTemp("", "bastion-key-*.pem")
	if err != nil {
		return fmt.Errorf("creating temp key file: %w", err)
	}
	defer func() { _ = os.Remove(keyFile.Name()) }()

	input.PrivateKeyFile = keyFile

	conn, err := connectSvc.Prepare(ctx, input)
	if err != nil {
		return err
	}

	if onReady != nil {
		onReady()
	}

	sshCmd := exec.Command("ssh", conn.SSHArgs...)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	if err := sshCmd.Start(); err != nil {
		return fmt.Errorf("starting ssh: %w", err)
	}

	// Single shutdown path for both signals and context cancellation.
	// Using exec.Command (not CommandContext) so this goroutine is the sole
	// arbiter: signals are forwarded gracefully; context cancellation sends
	// SIGTERM and escalates to SIGKILL after a grace period.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer signal.Stop(sigCh)
		select {
		case sig := <-sigCh:
			_ = sshCmd.Process.Signal(sig)
		case <-ctx.Done():
			_ = sshCmd.Process.Signal(syscall.SIGTERM)
			time.Sleep(5 * time.Second)
			_ = sshCmd.Process.Kill()
		}
	}()

	return sshCmd.Wait()
}
