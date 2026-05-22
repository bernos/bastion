package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os/exec"

	"golang.org/x/crypto/ssh"
)

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

func checkSSMDependencies() error {
	if _, err := exec.LookPath("aws"); err != nil {
		return fmt.Errorf("aws CLI not found on PATH — install it from https://aws.amazon.com/cli/")
	}
	if _, err := exec.LookPath("session-manager-plugin"); err != nil {
		return fmt.Errorf("session-manager-plugin not found on PATH — install it from https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html")
	}
	return nil
}

func buildSSHArgs(instanceID, region, keyFile string, extraArgs []string) []string {
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
	args = append(args, fmt.Sprintf("ec2-user@%s", instanceID))
	return args
}
