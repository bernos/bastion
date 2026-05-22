package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2ic "github.com/aws/aws-sdk-go-v2/service/ec2instanceconnect"
	"github.com/bernos/bastion/internal/bastion"
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

			return runConnect(cmd, cfg.Name, cfg.Region, svc, ec2ic.NewFromConfig(awsCfg), nil)
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

// runConnect is the shared connection flow for both ssh and proxy commands.
// extraSSHArgs are inserted before the target (e.g. ["-D", "1080", "-N"] for proxy).
func runConnect(cmd *cobra.Command, name, region string, svc bastion.BastionService, ec2icClient ec2icSender, extraSSHArgs []string) error {
	ctx := cmd.Context()

	described, err := svc.DescribeBastion(ctx, &bastion.DescribeBastionInput{
		BastionName: name,
		Region:      region,
	})
	if err != nil {
		return err
	}

	if err := svc.WaitForSSMReady(ctx, &bastion.WaitForSSMReadyInput{
		InstanceID: described.InstanceID,
		Region:     region,
	}); err != nil {
		return err
	}

	privateKeyPEM, publicKey, err := generateEphemeralKeyPair()
	if err != nil {
		return err
	}

	if _, err := ec2icClient.SendSSHPublicKey(ctx, &ec2ic.SendSSHPublicKeyInput{
		InstanceId:       aws.String(described.InstanceID),
		AvailabilityZone: aws.String(described.AvailabilityZone),
		InstanceOSUser:   aws.String("ec2-user"),
		SSHPublicKey:     aws.String(publicKey),
	}); err != nil {
		return fmt.Errorf("uploading SSH public key: %w", err)
	}

	keyFile, err := os.CreateTemp("", "bastion-key-*.pem")
	if err != nil {
		return fmt.Errorf("creating temp key file: %w", err)
	}
	keyPath := keyFile.Name()
	defer func() { _ = os.Remove(keyPath) }()

	if err := keyFile.Chmod(0o600); err != nil {
		_ = keyFile.Close()
		return fmt.Errorf("setting key file permissions: %w", err)
	}
	if _, err := keyFile.Write(privateKeyPEM); err != nil {
		_ = keyFile.Close()
		return fmt.Errorf("writing key file: %w", err)
	}
	if err := keyFile.Close(); err != nil {
		return fmt.Errorf("closing key file: %w", err)
	}

	sshArgs := buildSSHArgs(described.InstanceID, region, keyPath, extraSSHArgs)
	sshCmd := exec.CommandContext(ctx, "ssh", sshArgs...)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	if err := sshCmd.Start(); err != nil {
		return fmt.Errorf("starting ssh: %w", err)
	}

	// Forward SIGINT and SIGTERM to the ssh child so both interactive sessions
	// and proxy mode shut down cleanly on Ctrl-C or SIGTERM.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range sigCh {
			if sshCmd.Process != nil {
				_ = sshCmd.Process.Signal(sig)
			}
		}
	}()

	err = sshCmd.Wait()
	signal.Stop(sigCh)
	close(sigCh)
	return err
}

type ec2icSender interface {
	SendSSHPublicKey(ctx context.Context, params *ec2ic.SendSSHPublicKeyInput, optFns ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error)
}
