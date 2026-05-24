package connect

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ec2ic "github.com/aws/aws-sdk-go-v2/service/ec2instanceconnect"
	"github.com/bernos/bastion/internal/bastion"
)

// mockBastionService satisfies bastion.BastionService for tests.
type mockBastionService struct {
	DeployBastionFn   func(context.Context, *bastion.DeployBastionInput) (*bastion.DeployBastionOutput, error)
	DeleteBastionFn   func(context.Context, *bastion.DeleteBastionInput) error
	DescribeBastionFn func(context.Context, *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error)
	WaitForSSMReadyFn func(context.Context, *bastion.WaitForSSMReadyInput) error
}

func (m *mockBastionService) DeployBastion(ctx context.Context, input *bastion.DeployBastionInput) (*bastion.DeployBastionOutput, error) {
	if m.DeployBastionFn != nil {
		return m.DeployBastionFn(ctx, input)
	}
	return &bastion.DeployBastionOutput{}, nil
}

func (m *mockBastionService) DeleteBastion(ctx context.Context, input *bastion.DeleteBastionInput) error {
	if m.DeleteBastionFn != nil {
		return m.DeleteBastionFn(ctx, input)
	}
	return nil
}

func (m *mockBastionService) DescribeBastion(ctx context.Context, input *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error) {
	if m.DescribeBastionFn != nil {
		return m.DescribeBastionFn(ctx, input)
	}
	return &bastion.DescribeBastionOutput{InstanceID: "i-default", AvailabilityZone: "us-east-1a"}, nil
}

func (m *mockBastionService) WaitForSSMReady(ctx context.Context, input *bastion.WaitForSSMReadyInput) error {
	if m.WaitForSSMReadyFn != nil {
		return m.WaitForSSMReadyFn(ctx, input)
	}
	return nil
}

// mockEC2ICSender satisfies ec2icSender for tests.
type mockEC2ICSender struct {
	SendSSHPublicKeyFn func(context.Context, *ec2ic.SendSSHPublicKeyInput, ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error)
}

func (m *mockEC2ICSender) SendSSHPublicKey(ctx context.Context, params *ec2ic.SendSSHPublicKeyInput, optFns ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error) {
	if m.SendSSHPublicKeyFn != nil {
		return m.SendSSHPublicKeyFn(ctx, params, optFns...)
	}
	return &ec2ic.SendSSHPublicKeyOutput{}, nil
}

// mockPrivateKeyWriter satisfies PrivateKeyWriter for tests.
type mockPrivateKeyWriter struct {
	name    string
	ChmodFn func(fs.FileMode) error
	WriteFn func([]byte) (int, error)
	CloseFn func() error
}

func (m *mockPrivateKeyWriter) Chmod(mode fs.FileMode) error {
	if m.ChmodFn != nil {
		return m.ChmodFn(mode)
	}
	return nil
}

func (m *mockPrivateKeyWriter) Write(p []byte) (int, error) {
	if m.WriteFn != nil {
		return m.WriteFn(p)
	}
	return len(p), nil
}

func (m *mockPrivateKeyWriter) Close() error {
	if m.CloseFn != nil {
		return m.CloseFn()
	}
	return nil
}

func (m *mockPrivateKeyWriter) Name() string { return m.name }

// noopKeyFile returns a PrivateKeyWriter stub for tests that stop before the key-write step.
func noopKeyFile(t *testing.T) *mockPrivateKeyWriter {
	t.Helper()
	return &mockPrivateKeyWriter{name: filepath.Join(t.TempDir(), "key.pem")}
}

// stubPath creates a temporary directory with stub executables and returns the directory path.
func stubPath(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatalf("creating stub %s: %v", name, err)
		}
	}
	return dir
}

// --- CheckDependencies tests ---

func Test_CheckDependencies_BothPresent(t *testing.T) {
	dir := stubPath(t, "aws", "session-manager-plugin")
	t.Setenv("PATH", dir)

	svc := NewConnectService(&mockBastionService{}, &mockEC2ICSender{})
	if err := svc.CheckDependencies(); err != nil {
		t.Errorf("expected nil, got: %v", err)
	}
}

func Test_CheckDependencies_AWSMissing(t *testing.T) {
	dir := stubPath(t, "session-manager-plugin")
	t.Setenv("PATH", dir)

	svc := NewConnectService(&mockBastionService{}, &mockEC2ICSender{})
	err := svc.CheckDependencies()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "aws") {
		t.Errorf("error should reference 'aws', got: %v", err)
	}
}

func Test_CheckDependencies_PluginMissing(t *testing.T) {
	dir := stubPath(t, "aws")
	t.Setenv("PATH", dir)

	svc := NewConnectService(&mockBastionService{}, &mockEC2ICSender{})
	err := svc.CheckDependencies()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "session-manager-plugin") {
		t.Errorf("error should reference 'session-manager-plugin', got: %v", err)
	}
	if !strings.Contains(err.Error(), "https://") {
		t.Errorf("error should contain install URL, got: %v", err)
	}
}

// --- Prepare tests ---

func Test_Prepare_DescribeBastionError_Propagated(t *testing.T) {
	describeErr := errors.New("stack not found")

	svc := NewConnectService(&mockBastionService{
		DescribeBastionFn: func(_ context.Context, _ *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error) {
			return nil, describeErr
		},
	}, &mockEC2ICSender{})

	conn, err := svc.Prepare(context.Background(), &PrepareInput{
		BastionName:    "my-bastion",
		Region:         "ap-southeast-2",
		PrivateKeyFile: noopKeyFile(t),
	})
	if conn != nil {
		t.Error("expected nil Connection")
	}
	if !errors.Is(err, describeErr) {
		t.Errorf("expected describeErr, got: %v", err)
	}
}

func Test_Prepare_WaitForSSMReadyError_Propagated(t *testing.T) {
	ssmErr := errors.New("ssm timeout")

	svc := NewConnectService(&mockBastionService{
		DescribeBastionFn: func(_ context.Context, _ *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error) {
			return &bastion.DescribeBastionOutput{InstanceID: "i-abc001", AvailabilityZone: "ap-southeast-2a"}, nil
		},
		WaitForSSMReadyFn: func(_ context.Context, _ *bastion.WaitForSSMReadyInput) error {
			return ssmErr
		},
	}, &mockEC2ICSender{})

	conn, err := svc.Prepare(context.Background(), &PrepareInput{
		BastionName:    "my-bastion",
		Region:         "ap-southeast-2",
		PrivateKeyFile: noopKeyFile(t),
	})
	if conn != nil {
		t.Error("expected nil Connection")
	}
	if !errors.Is(err, ssmErr) {
		t.Errorf("expected ssmErr, got: %v", err)
	}
}

func Test_Prepare_KeyUploadError_Propagated(t *testing.T) {
	uploadErr := errors.New("ec2ic denied")

	svc := NewConnectService(&mockBastionService{
		DescribeBastionFn: func(_ context.Context, _ *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error) {
			return &bastion.DescribeBastionOutput{InstanceID: "i-abc001", AvailabilityZone: "ap-southeast-2a"}, nil
		},
	}, &mockEC2ICSender{
		SendSSHPublicKeyFn: func(_ context.Context, _ *ec2ic.SendSSHPublicKeyInput, _ ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error) {
			return nil, uploadErr
		},
	})

	conn, err := svc.Prepare(context.Background(), &PrepareInput{
		BastionName:    "my-bastion",
		Region:         "ap-southeast-2",
		PrivateKeyFile: noopKeyFile(t),
	})
	if conn != nil {
		t.Error("expected nil Connection")
	}
	if !errors.Is(err, uploadErr) {
		t.Errorf("expected uploadErr, got: %v", err)
	}
}

func Test_Prepare_OSUser_DefaultsToEC2User(t *testing.T) {
	var capturedInput *ec2ic.SendSSHPublicKeyInput

	svc := NewConnectService(&mockBastionService{
		DescribeBastionFn: func(_ context.Context, _ *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error) {
			return &bastion.DescribeBastionOutput{InstanceID: "i-abc001", AvailabilityZone: "ap-southeast-2a"}, nil
		},
	}, &mockEC2ICSender{
		SendSSHPublicKeyFn: func(_ context.Context, params *ec2ic.SendSSHPublicKeyInput, _ ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error) {
			capturedInput = params
			return nil, errors.New("stop here")
		},
	})

	_, _ = svc.Prepare(context.Background(), &PrepareInput{
		BastionName:    "my-bastion",
		Region:         "ap-southeast-2",
		OSUser:         "", // empty → should default to "ec2-user"
		PrivateKeyFile: noopKeyFile(t),
	})

	if capturedInput == nil {
		t.Fatal("SendSSHPublicKey was not called")
	}
	if *capturedInput.InstanceOSUser != "ec2-user" {
		t.Errorf("InstanceOSUser: want %q, got %q", "ec2-user", *capturedInput.InstanceOSUser)
	}
}

func Test_Prepare_OSUser_ExplicitValue(t *testing.T) {
	var capturedInput *ec2ic.SendSSHPublicKeyInput

	svc := NewConnectService(&mockBastionService{
		DescribeBastionFn: func(_ context.Context, _ *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error) {
			return &bastion.DescribeBastionOutput{InstanceID: "i-abc001", AvailabilityZone: "ap-southeast-2a"}, nil
		},
	}, &mockEC2ICSender{
		SendSSHPublicKeyFn: func(_ context.Context, params *ec2ic.SendSSHPublicKeyInput, _ ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error) {
			capturedInput = params
			return nil, errors.New("stop here")
		},
	})

	_, _ = svc.Prepare(context.Background(), &PrepareInput{
		BastionName:    "my-bastion",
		Region:         "ap-southeast-2",
		OSUser:         "ubuntu",
		PrivateKeyFile: noopKeyFile(t),
	})

	if capturedInput == nil {
		t.Fatal("SendSSHPublicKey was not called")
	}
	if *capturedInput.InstanceOSUser != "ubuntu" {
		t.Errorf("InstanceOSUser: want %q, got %q", "ubuntu", *capturedInput.InstanceOSUser)
	}
}

func Test_Prepare_Success_ReturnsConnectionWithSSHArgs(t *testing.T) {
	keyFile, err := os.CreateTemp("", "bastion-key-test-*.pem")
	if err != nil {
		t.Fatalf("creating temp key file: %v", err)
	}
	keyPath := keyFile.Name()
	defer func() { _ = os.Remove(keyPath) }()

	svc := NewConnectService(&mockBastionService{
		DescribeBastionFn: func(_ context.Context, _ *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error) {
			return &bastion.DescribeBastionOutput{InstanceID: "i-abc001", AvailabilityZone: "ap-southeast-2a"}, nil
		},
	}, &mockEC2ICSender{})

	conn, err := svc.Prepare(context.Background(), &PrepareInput{
		BastionName:    "my-bastion",
		Region:         "ap-southeast-2",
		PrivateKeyFile: keyFile,
	})

	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if conn == nil {
		t.Fatal("expected non-nil Connection")
	}
	if len(conn.SSHArgs) == 0 {
		t.Error("SSHArgs must be non-empty")
	}
}

// --- buildSSHArgs tests ---

func Test_buildSSHArgs_SSHMode(t *testing.T) {
	args := buildSSHArgs("i-abc001", "ap-southeast-2", "/tmp/key.pem", "ec2-user", nil)

	last := args[len(args)-1]
	if last != "ec2-user@i-abc001" {
		t.Errorf("target: want %q, got %q", "ec2-user@i-abc001", last)
	}

	assertArg(t, args, "-i", "/tmp/key.pem")
	assertArg(t, args, "-o", "StrictHostKeyChecking=no")
	assertArg(t, args, "-o", "UserKnownHostsFile=/dev/null")
	assertProxyCmd(t, args, "ap-southeast-2")

	for _, a := range args[:len(args)-1] {
		if a == "-D" || a == "-N" {
			t.Errorf("unexpected proxy flag %q in ssh mode args", a)
		}
	}
}

func Test_buildSSHArgs_ProxyMode(t *testing.T) {
	args := buildSSHArgs("i-abc001", "us-east-1", "/tmp/key.pem", "ec2-user", []string{"-D", "1080", "-N"})

	last := args[len(args)-1]
	if last != "ec2-user@i-abc001" {
		t.Errorf("target: want %q, got %q", "ec2-user@i-abc001", last)
	}

	assertArg(t, args, "-D", "1080")
	assertProxyCmd(t, args, "us-east-1")

	nIdx, targetIdx := -1, -1
	for i, a := range args {
		switch a {
		case "-N":
			nIdx = i
		case "ec2-user@i-abc001":
			targetIdx = i
		}
	}
	if nIdx < 0 {
		t.Fatal("-N not found in proxy mode args")
	}
	if nIdx >= targetIdx {
		t.Errorf("-N (index %d) must come before target (index %d)", nIdx, targetIdx)
	}
}

func Test_buildSSHArgs_RegionInProxyCommand(t *testing.T) {
	for _, region := range []string{"ap-southeast-2", "us-west-1", "eu-central-1"} {
		args := buildSSHArgs("i-000", region, "/k", "ec2-user", nil)
		assertProxyCmd(t, args, region)
	}
}

func Test_buildSSHArgs_CustomOSUser(t *testing.T) {
	args := buildSSHArgs("i-abc001", "us-east-1", "/tmp/key.pem", "ubuntu", nil)
	last := args[len(args)-1]
	if last != "ubuntu@i-abc001" {
		t.Errorf("target: want %q, got %q", "ubuntu@i-abc001", last)
	}
}

// assertArg checks that flag and value appear as consecutive elements in args.
func assertArg(t *testing.T, args []string, flag, value string) {
	t.Helper()
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag && args[i+1] == value {
			return
		}
	}
	t.Errorf("expected %q %q in args %v", flag, value, args)
}

// assertProxyCmd checks that a ProxyCommand option containing the given region is present.
func assertProxyCmd(t *testing.T, args []string, region string) {
	t.Helper()
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-o" && strings.HasPrefix(args[i+1], "ProxyCommand=") {
			if strings.Contains(args[i+1], "--region "+region) {
				return
			}
			t.Errorf("ProxyCommand %q does not contain --region %s", args[i+1], region)
			return
		}
	}
	t.Errorf("no ProxyCommand option found in args %v", args)
}
