package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/bernos/bastion/internal/bastion"
)

func Test_runUp_AMIParameterName_Passthrough(t *testing.T) {
	var captured *bastion.DeployBastionInput

	svc := &mockBastionService{
		DeployBastionFn: func(_ context.Context, input *bastion.DeployBastionInput) (*bastion.DeployBastionOutput, error) {
			captured = input
			return &bastion.DeployBastionOutput{}, nil
		},
	}

	cmd := NewUpCommand(svc, bytes.NewBufferString(""), &bytes.Buffer{}, &bytes.Buffer{})

	if err := cmd.Run(context.Background(), &UpInput{
		BastionName:      "my-bastion",
		Owner:            "owner",
		SubnetID:         "subnet-123",
		VPCID:            "vpc-123",
		AMIParameterName: "/custom/ami/path",
	}); err != nil {
		t.Fatal(err)
	}

	if captured.AMIParameterName != "/custom/ami/path" {
		t.Errorf("AMIParameterName: want %q, got %q", "/custom/ami/path", captured.AMIParameterName)
	}
}
