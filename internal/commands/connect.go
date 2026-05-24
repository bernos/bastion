package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/bernos/bastion/internal/connect"
)

type ConnectInput struct {
	BastionName  string
	Region       string
	OSUser       string   // defaults to "ec2-user" if empty
	ExtraSSHArgs []string // e.g. ["-D", "1080", "-N"] for proxy mode
	OnReady      func()
}

// func runConnect(ctx context.Context, connectSvc connect.ConnectService, input *connect.PrepareInput, onReady func()) error {
type connectCommand struct {
	connectService connect.ConnectService
	stdin          io.Reader
	stdout         io.Writer
	stderr         io.Writer
}

func NewConnectCommand(svc connect.ConnectService, stdin io.Reader, stdout io.Writer, stderr io.Writer) *connectCommand {
	return &connectCommand{
		connectService: svc,
		stdin:          stdin,
		stdout:         stdout,
		stderr:         stderr,
	}
}

func (c *connectCommand) Run(ctx context.Context, input *ConnectInput) error {

	if err := c.connectService.CheckDependencies(); err != nil {
		return err
	}

	keyFile, err := os.CreateTemp("", "bastion-key-*.pem")
	if err != nil {
		return fmt.Errorf("creating temp key file: %w", err)
	}
	defer func() { _ = os.Remove(keyFile.Name()) }()

	conn, err := c.connectService.Prepare(ctx, &connect.PrepareInput{
		BastionName:    input.BastionName,
		Region:         input.Region,
		OSUser:         input.OSUser,
		ExtraSSHArgs:   input.ExtraSSHArgs,
		PrivateKeyFile: keyFile,
	})

	if err != nil {
		return err
	}

	if input.OnReady != nil {
		input.OnReady()
	}

	sshCmd := exec.Command("ssh", conn.SSHArgs...)
	sshCmd.Stdin = c.stdin
	sshCmd.Stdout = c.stdout
	sshCmd.Stderr = c.stderr

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
