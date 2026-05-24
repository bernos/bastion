package connect

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2ic "github.com/aws/aws-sdk-go-v2/service/ec2instanceconnect"
	"github.com/bernos/bastion/internal/bastion"
	"golang.org/x/crypto/ssh"
)

type ec2icSender interface {
	SendSSHPublicKey(ctx context.Context, params *ec2ic.SendSSHPublicKeyInput, optFns ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error)
}

type connectService struct {
	bastionSvc bastion.BastionService
	ec2ic      ec2icSender
}

func NewConnectService(bastionSvc bastion.BastionService, ec2ic ec2icSender) ConnectService {
	return &connectService{bastionSvc: bastionSvc, ec2ic: ec2ic}
}

func (s *connectService) CheckDependencies() error {
	if _, err := exec.LookPath("aws"); err != nil {
		return fmt.Errorf("aws CLI not found on PATH — install it from https://aws.amazon.com/cli/")
	}
	if _, err := exec.LookPath("session-manager-plugin"); err != nil {
		return fmt.Errorf("session-manager-plugin not found on PATH — install it from https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html")
	}
	return nil
}

func (s *connectService) Prepare(ctx context.Context, input *PrepareInput) (*Connection, error) {
	described, err := s.bastionSvc.DescribeBastion(ctx, &bastion.DescribeBastionInput{
		BastionName: input.BastionName,
	})
	if err != nil {
		return nil, err
	}

	if err := s.bastionSvc.WaitForSSMReady(ctx, &bastion.WaitForSSMReadyInput{
		InstanceID: described.InstanceID,
	}); err != nil {
		return nil, err
	}

	privateKeyPEM, publicKey, err := generateEphemeralKeyPair()
	if err != nil {
		return nil, err
	}

	osUser := input.OSUser
	if osUser == "" {
		osUser = "ec2-user"
	}

	if _, err := s.ec2ic.SendSSHPublicKey(ctx, &ec2ic.SendSSHPublicKeyInput{
		InstanceId:       aws.String(described.InstanceID),
		AvailabilityZone: aws.String(described.AvailabilityZone),
		InstanceOSUser:   aws.String(osUser),
		SSHPublicKey:     aws.String(publicKey),
	}); err != nil {
		return nil, fmt.Errorf("uploading SSH public key: %w", err)
	}

	keyPath := input.PrivateKeyFile.Name()

	if err := input.PrivateKeyFile.Chmod(0o600); err != nil {
		_ = input.PrivateKeyFile.Close()
		return nil, fmt.Errorf("setting key file permissions: %w", err)
	}
	if _, err := input.PrivateKeyFile.Write(privateKeyPEM); err != nil {
		_ = input.PrivateKeyFile.Close()
		return nil, fmt.Errorf("writing key file: %w", err)
	}
	if err := input.PrivateKeyFile.Close(); err != nil {
		return nil, fmt.Errorf("closing key file: %w", err)
	}

	sshArgs := buildSSHArgs(described.InstanceID, input.Region, keyPath, osUser, input.ExtraSSHArgs)
	return &Connection{SSHArgs: sshArgs}, nil
}

func generateEphemeralKeyPair() (privateKeyPEM []byte, publicKeyAuthorized string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, "", fmt.Errorf("generating ed25519 key pair: %w", err)
	}

	privBytes, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return nil, "", fmt.Errorf("marshalling private key: %w", err)
	}
	privateKeyPEM = pem.EncodeToMemory(privBytes)

	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, "", fmt.Errorf("converting public key: %w", err)
	}
	publicKeyAuthorized = string(ssh.MarshalAuthorizedKey(sshPub))

	return privateKeyPEM, publicKeyAuthorized, nil
}

func buildSSHArgs(instanceID, region, keyFile, osUser string, extraArgs []string) []string {
	proxyCmd := fmt.Sprintf(
		"aws ssm start-session --target %%h --document-name AWS-StartSSHSession --parameters portNumber=22 --region %s",
		region,
	)

	// The SSH target is an EC2 instance ID (e.g. i-0abc123), not a real hostname,
	// so it will never match a known_hosts entry. Disable host key checking to
	// avoid polluting known_hosts with ephemeral instance IDs.
	args := []string{
		"-i", keyFile,
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ProxyCommand=" + proxyCmd,
	}
	args = append(args, extraArgs...)
	args = append(args, fmt.Sprintf("%s@%s", osUser, instanceID))
	return args
}
