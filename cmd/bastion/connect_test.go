package main

import (
	"context"
	"errors"
	"testing"

	"github.com/bernos/bastion/internal/connect"
	"github.com/spf13/cobra"
)

// mockConnectService satisfies connect.ConnectService for tests.
type mockConnectService struct {
	CheckDependenciesFn func() error
	PrepareFn           func(context.Context, *connect.PrepareInput) (*connect.Connection, error)
}

func (m *mockConnectService) CheckDependencies() error {
	if m.CheckDependenciesFn != nil {
		return m.CheckDependenciesFn()
	}
	return nil
}

func (m *mockConnectService) Prepare(ctx context.Context, input *connect.PrepareInput) (*connect.Connection, error) {
	if m.PrepareFn != nil {
		return m.PrepareFn(ctx, input)
	}
	return &connect.Connection{SSHArgs: []string{"-V"}}, nil
}

func testCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	return cmd
}

func Test_runConnect_CheckDependencies_Error_Propagated(t *testing.T) {
	depErr := errors.New("aws not found")

	svc := &mockConnectService{
		CheckDependenciesFn: func() error { return depErr },
	}

	err := runConnect(testCmd(), svc, &connect.PrepareInput{BastionName: "my-bastion", Region: "ap-southeast-2"}, nil)
	if !errors.Is(err, depErr) {
		t.Errorf("expected depErr, got: %v", err)
	}
}

func Test_runConnect_Prepare_Error_Propagated(t *testing.T) {
	prepErr := errors.New("stack not found")

	svc := &mockConnectService{
		PrepareFn: func(_ context.Context, _ *connect.PrepareInput) (*connect.Connection, error) {
			return nil, prepErr
		},
	}

	err := runConnect(testCmd(), svc, &connect.PrepareInput{BastionName: "my-bastion", Region: "ap-southeast-2"}, nil)
	if !errors.Is(err, prepErr) {
		t.Errorf("expected prepErr, got: %v", err)
	}
}

func Test_runConnect_SSH_PassesBastionNameAndRegion(t *testing.T) {
	var capturedInput *connect.PrepareInput

	svc := &mockConnectService{
		PrepareFn: func(_ context.Context, input *connect.PrepareInput) (*connect.Connection, error) {
			capturedInput = input
			return nil, errors.New("stop here")
		},
	}

	_ = runConnect(testCmd(), svc, &connect.PrepareInput{BastionName: "my-bastion", Region: "ap-southeast-2"}, nil)

	if capturedInput == nil {
		t.Fatal("Prepare was not called")
	}
	if capturedInput.BastionName != "my-bastion" {
		t.Errorf("BastionName: want %q, got %q", "my-bastion", capturedInput.BastionName)
	}
	if capturedInput.Region != "ap-southeast-2" {
		t.Errorf("Region: want %q, got %q", "ap-southeast-2", capturedInput.Region)
	}
}

func Test_runConnect_Proxy_PassesExtraSSHArgs(t *testing.T) {
	var capturedInput *connect.PrepareInput
	extraArgs := []string{"-D", "1080", "-N"}

	svc := &mockConnectService{
		PrepareFn: func(_ context.Context, input *connect.PrepareInput) (*connect.Connection, error) {
			capturedInput = input
			return nil, errors.New("stop here")
		},
	}

	_ = runConnect(testCmd(), svc, &connect.PrepareInput{
		BastionName:  "my-bastion",
		Region:       "ap-southeast-2",
		ExtraSSHArgs: extraArgs,
	}, nil)

	if capturedInput == nil {
		t.Fatal("Prepare was not called")
	}
	if len(capturedInput.ExtraSSHArgs) != 3 {
		t.Fatalf("ExtraSSHArgs: want 3 elements, got %d", len(capturedInput.ExtraSSHArgs))
	}
	if capturedInput.ExtraSSHArgs[0] != "-D" || capturedInput.ExtraSSHArgs[1] != "1080" || capturedInput.ExtraSSHArgs[2] != "-N" {
		t.Errorf("ExtraSSHArgs: want [-D 1080 -N], got %v", capturedInput.ExtraSSHArgs)
	}
}

func Test_runConnect_OnReady_NotCalledOnPrepareError(t *testing.T) {
	onReadyCalled := false

	svc := &mockConnectService{
		PrepareFn: func(_ context.Context, _ *connect.PrepareInput) (*connect.Connection, error) {
			return nil, errors.New("prepare failed")
		},
	}

	_ = runConnect(testCmd(), svc, &connect.PrepareInput{BastionName: "my-bastion", Region: "ap-southeast-2"}, func() {
		onReadyCalled = true
	})

	if onReadyCalled {
		t.Error("onReady should not be called when Prepare fails")
	}
}
