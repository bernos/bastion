package connect

import "context"

// PrepareInput holds the parameters needed to prepare a bastion connection.
type PrepareInput struct {
	BastionName  string
	Region       string
	OSUser       string   // defaults to "ec2-user" if empty
	ExtraSSHArgs []string // e.g. ["-D", "1080", "-N"] for proxy mode
	OnReady      func()   // called after SSM ready, before key upload; nil is safe
}

// Connection holds exec-ready SSH arguments and the path to the ephemeral private key.
// The caller is responsible for removing KeyPath after the SSH process exits.
type Connection struct {
	SSHArgs []string
	KeyPath string
}

// ConnectService prepares bastion connections for use by the ssh and proxy commands.
type ConnectService interface {
	// CheckDependencies verifies that aws and session-manager-plugin are on PATH.
	CheckDependencies() error

	// Prepare resolves the bastion, waits for SSM readiness, generates an ephemeral
	// key pair, uploads the public key via EC2 Instance Connect, and returns exec-ready
	// SSH arguments along with the temp private key path.
	Prepare(ctx context.Context, input *PrepareInput) (*Connection, error)
}
