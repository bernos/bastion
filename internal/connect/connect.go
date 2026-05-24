package connect

import (
	"context"
	"io"
	"io/fs"
)

// PrivateKeyWriter is satisfied by *os.File. Prepare writes the ephemeral private
// key into the caller-provided file rather than creating one itself.
type PrivateKeyWriter interface {
	io.WriteCloser
	Chmod(mode fs.FileMode) error
	Name() string
}

// PrepareInput holds the parameters needed to prepare a bastion connection.
type PrepareInput struct {
	BastionName    string
	Region         string
	OSUser         string           // defaults to "ec2-user" if empty
	ExtraSSHArgs   []string         // e.g. ["-D", "1080", "-N"] for proxy mode
	PrivateKeyFile PrivateKeyWriter // caller-created file that receives the private key
}

// Connection holds exec-ready SSH arguments returned by Prepare.
type Connection struct {
	SSHArgs []string
}

// ConnectService prepares bastion connections for use by the ssh and proxy commands.
type ConnectService interface {
	// CheckDependencies verifies that aws and session-manager-plugin are on PATH.
	CheckDependencies() error

	// Prepare resolves the bastion, waits for SSM readiness, generates an ephemeral
	// key pair, uploads the public key via EC2 Instance Connect, writes the private key
	// to input.PrivateKeyFile, and returns exec-ready SSH arguments.
	Prepare(ctx context.Context, input *PrepareInput) (*Connection, error)
}
