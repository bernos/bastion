package commands

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/bernos/bastion/internal/bastion"
)

type mockBastionService struct {
	DeleteBastionFn   func(context.Context, *bastion.DeleteBastionInput) error
	DeployBastionFn   func(context.Context, *bastion.DeployBastionInput) (*bastion.DeployBastionOutput, error)
	DescribeBastionFn func(context.Context, *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error)
	WaitForSSMReadyFn func(context.Context, *bastion.WaitForSSMReadyInput) error
}

func (m *mockBastionService) DeleteBastion(ctx context.Context, input *bastion.DeleteBastionInput) error {
	return m.DeleteBastionFn(ctx, input)
}

func (m *mockBastionService) DeployBastion(ctx context.Context, input *bastion.DeployBastionInput) (*bastion.DeployBastionOutput, error) {
	if m.DeployBastionFn != nil {
		return m.DeployBastionFn(ctx, input)
	}
	return &bastion.DeployBastionOutput{}, nil
}

func (m *mockBastionService) DescribeBastion(ctx context.Context, input *bastion.DescribeBastionInput) (*bastion.DescribeBastionOutput, error) {
	if m.DescribeBastionFn != nil {
		return m.DescribeBastionFn(ctx, input)
	}
	return &bastion.DescribeBastionOutput{}, nil
}

func (m *mockBastionService) WaitForSSMReady(ctx context.Context, input *bastion.WaitForSSMReadyInput) error {
	if m.WaitForSSMReadyFn != nil {
		return m.WaitForSSMReadyFn(ctx, input)
	}
	return nil
}

func Test_runDown_Success(t *testing.T) {
	var capturedInput *bastion.DeleteBastionInput

	svc := &mockBastionService{
		DeleteBastionFn: func(_ context.Context, input *bastion.DeleteBastionInput) error {
			capturedInput = input
			return nil
		},
	}

	cmd := NewDownCommand(svc, bytes.NewBufferString(""), &bytes.Buffer{}, &bytes.Buffer{})

	if err := cmd.Run(context.Background(), &DownInput{
		BastionName: "my-bastion",
	}); err != nil {
		t.Fatal(err)
	}

	if capturedInput.BastionName != "my-bastion" {
		t.Errorf("BastionName: want %q, got %q", "my-bastion", capturedInput.BastionName)
	}

}

func Test_runDown_ServiceError_Propagated(t *testing.T) {
	deleteErr := errors.New("stack not found")

	svc := &mockBastionService{
		DeleteBastionFn: func(_ context.Context, _ *bastion.DeleteBastionInput) error {
			return deleteErr
		},
	}

	cmd := NewDownCommand(svc, bytes.NewBufferString(""), &bytes.Buffer{}, &bytes.Buffer{})

	err := cmd.Run(context.Background(), &DownInput{
		BastionName: "missing",
	})

	if !errors.Is(err, deleteErr) {
		t.Errorf("expected deleteErr to be wrapped in returned error, got: %v", err)
	}
}
