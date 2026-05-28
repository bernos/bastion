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

func Test_runUp_Tags_Passthrough(t *testing.T) {
	var captured *bastion.DeployBastionInput

	svc := &mockBastionService{
		DeployBastionFn: func(_ context.Context, input *bastion.DeployBastionInput) (*bastion.DeployBastionOutput, error) {
			captured = input
			return &bastion.DeployBastionOutput{}, nil
		},
	}

	cmd := NewUpCommand(svc, bytes.NewBufferString(""), &bytes.Buffer{}, &bytes.Buffer{})

	if err := cmd.Run(context.Background(), &UpInput{
		BastionName: "my-bastion",
		Owner:       "owner",
		SubnetID:    "subnet-123",
		VPCID:       "vpc-123",
		Tags:        map[string]string{"env": "prod", "team": "platform"},
	}); err != nil {
		t.Fatal(err)
	}

	if captured.Tags["env"] != "prod" {
		t.Errorf("Tags[env]: want %q, got %q", "prod", captured.Tags["env"])
	}
	if captured.Tags["team"] != "platform" {
		t.Errorf("Tags[team]: want %q, got %q", "platform", captured.Tags["team"])
	}
}
