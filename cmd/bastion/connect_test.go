package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	ec2ic "github.com/aws/aws-sdk-go-v2/service/ec2instanceconnect"
	"github.com/bernos/bastion/internal/bastion"
	"github.com/spf13/cobra"
)

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

func testCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	return cmd
}

func Test_runConnect_DescribeBastionError_Propagated(t *testing.T) {
	describeErr := errors.New("stack not found")

	svc := &mockBastionService{
		DescribeBastionFn: func(_ context.Context, _ *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error) {
			return nil, describeErr
		},
	}

	err := runConnect(testCmd(), "my-bastion", "ap-southeast-2", svc, &mockEC2ICSender{}, nil, nil)
	if !errors.Is(err, describeErr) {
		t.Errorf("expected describeErr, got: %v", err)
	}
}

func Test_runConnect_WaitForSSMReadyError_Propagated(t *testing.T) {
	ssmErr := errors.New("ssm timeout")

	svc := &mockBastionService{
		DescribeBastionFn: func(_ context.Context, _ *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error) {
			return &bastion.DescribeBastionOutput{InstanceID: "i-abc001", AvailabilityZone: "ap-southeast-2a"}, nil
		},
		WaitForSSMReadyFn: func(_ context.Context, _ *bastion.WaitForSSMReadyInput) error {
			return ssmErr
		},
	}

	err := runConnect(testCmd(), "my-bastion", "ap-southeast-2", svc, &mockEC2ICSender{}, nil, nil)
	if !errors.Is(err, ssmErr) {
		t.Errorf("expected ssmErr, got: %v", err)
	}
}

func Test_runConnect_OnReady_CalledAfterSSMReady(t *testing.T) {
	var seq []string

	svc := &mockBastionService{
		DescribeBastionFn: func(_ context.Context, _ *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error) {
			return &bastion.DescribeBastionOutput{InstanceID: "i-abc001", AvailabilityZone: "ap-southeast-2a"}, nil
		},
		WaitForSSMReadyFn: func(_ context.Context, _ *bastion.WaitForSSMReadyInput) error {
			seq = append(seq, "ssm")
			return nil
		},
	}
	sender := &mockEC2ICSender{
		SendSSHPublicKeyFn: func(_ context.Context, _ *ec2ic.SendSSHPublicKeyInput, _ ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error) {
			seq = append(seq, "upload")
			return nil, errors.New("stop here")
		},
	}
	onReady := func() { seq = append(seq, "ready") }

	_ = runConnect(testCmd(), "my-bastion", "ap-southeast-2", svc, sender, nil, onReady)

	want := []string{"ssm", "ready", "upload"}
	if strings.Join(seq, ",") != strings.Join(want, ",") {
		t.Errorf("call order: want %v, got %v", want, seq)
	}
}

func Test_runConnect_OnReady_NotCalledOnSSMError(t *testing.T) {
	onReadyCalled := false

	svc := &mockBastionService{
		DescribeBastionFn: func(_ context.Context, _ *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error) {
			return &bastion.DescribeBastionOutput{InstanceID: "i-abc001", AvailabilityZone: "ap-southeast-2a"}, nil
		},
		WaitForSSMReadyFn: func(_ context.Context, _ *bastion.WaitForSSMReadyInput) error {
			return errors.New("ssm timeout")
		},
	}

	_ = runConnect(testCmd(), "my-bastion", "ap-southeast-2", svc, &mockEC2ICSender{}, nil, func() {
		onReadyCalled = true
	})

	if onReadyCalled {
		t.Error("onReady should not be called when WaitForSSMReady fails")
	}
}

func Test_runConnect_KeyUploadError_Propagated(t *testing.T) {
	uploadErr := errors.New("ec2ic denied")

	svc := &mockBastionService{
		DescribeBastionFn: func(_ context.Context, _ *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error) {
			return &bastion.DescribeBastionOutput{InstanceID: "i-abc001", AvailabilityZone: "ap-southeast-2a"}, nil
		},
		WaitForSSMReadyFn: func(_ context.Context, _ *bastion.WaitForSSMReadyInput) error {
			return nil
		},
	}
	sender := &mockEC2ICSender{
		SendSSHPublicKeyFn: func(_ context.Context, _ *ec2ic.SendSSHPublicKeyInput, _ ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error) {
			return nil, uploadErr
		},
	}

	err := runConnect(testCmd(), "my-bastion", "ap-southeast-2", svc, sender, nil, nil)
	if !errors.Is(err, uploadErr) {
		t.Errorf("expected uploadErr, got: %v", err)
	}
}

func Test_runConnect_KeyUpload_UsesCorrectInstanceDetails(t *testing.T) {
	var capturedInput *ec2ic.SendSSHPublicKeyInput

	svc := &mockBastionService{
		DescribeBastionFn: func(_ context.Context, _ *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error) {
			return &bastion.DescribeBastionOutput{InstanceID: "i-abc001", AvailabilityZone: "ap-southeast-2a"}, nil
		},
		WaitForSSMReadyFn: func(_ context.Context, _ *bastion.WaitForSSMReadyInput) error {
			return nil
		},
	}
	sender := &mockEC2ICSender{
		SendSSHPublicKeyFn: func(_ context.Context, params *ec2ic.SendSSHPublicKeyInput, _ ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error) {
			capturedInput = params
			// Return an error to stop before exec — we only care about the upload call.
			return nil, errors.New("stop here")
		},
	}

	_ = runConnect(testCmd(), "my-bastion", "ap-southeast-2", svc, sender, nil, nil)

	if capturedInput == nil {
		t.Fatal("SendSSHPublicKey was not called")
	}
	if *capturedInput.InstanceId != "i-abc001" {
		t.Errorf("InstanceId: want %q, got %q", "i-abc001", *capturedInput.InstanceId)
	}
	if *capturedInput.AvailabilityZone != "ap-southeast-2a" {
		t.Errorf("AvailabilityZone: want %q, got %q", "ap-southeast-2a", *capturedInput.AvailabilityZone)
	}
	if *capturedInput.InstanceOSUser != "ec2-user" {
		t.Errorf("InstanceOSUser: want %q, got %q", "ec2-user", *capturedInput.InstanceOSUser)
	}
}

func Test_buildSSHArgs_SSHMode(t *testing.T) {
	args := buildSSHArgs("i-abc001", "ap-southeast-2", "/tmp/key.pem", nil)

	last := args[len(args)-1]
	if last != "ec2-user@i-abc001" {
		t.Errorf("target: want %q, got %q", "ec2-user@i-abc001", last)
	}

	assertArg(t, args, "-i", "/tmp/key.pem")
	assertArg(t, args, "-o", "StrictHostKeyChecking=no")
	assertArg(t, args, "-o", "UserKnownHostsFile=/dev/null")
	assertProxyCmd(t, args, "ap-southeast-2")

	// No extra args — target must immediately follow the fixed flags.
	for _, a := range args[:len(args)-1] {
		if a == "-D" || a == "-N" {
			t.Errorf("unexpected proxy flag %q in ssh mode args", a)
		}
	}
}

func Test_buildSSHArgs_ProxyMode(t *testing.T) {
	args := buildSSHArgs("i-abc001", "us-east-1", "/tmp/key.pem", []string{"-D", "1080", "-N"})

	last := args[len(args)-1]
	if last != "ec2-user@i-abc001" {
		t.Errorf("target: want %q, got %q", "ec2-user@i-abc001", last)
	}

	assertArg(t, args, "-D", "1080")
	assertProxyCmd(t, args, "us-east-1")

	// -N must appear before the target
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
		args := buildSSHArgs("i-000", region, "/k", nil)
		assertProxyCmd(t, args, region)
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
