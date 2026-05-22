package main

import (
	"context"
	"errors"
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

	err := runConnect(testCmd(), "my-bastion", "ap-southeast-2", svc, &mockEC2ICSender{}, nil)
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

	err := runConnect(testCmd(), "my-bastion", "ap-southeast-2", svc, &mockEC2ICSender{}, nil)
	if !errors.Is(err, ssmErr) {
		t.Errorf("expected ssmErr, got: %v", err)
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

	err := runConnect(testCmd(), "my-bastion", "ap-southeast-2", svc, sender, nil)
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

	_ = runConnect(testCmd(), "my-bastion", "ap-southeast-2", svc, sender, nil)

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
