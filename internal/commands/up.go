package commands

import (
	"context"
	"encoding/json"
	"io"

	"github.com/bernos/bastion/internal/bastion"
)

type UpInput struct {
	BastionName string
	Owner       string
	SubnetID    string
	VPCID       string
}

type upCommand struct {
	bastionService bastion.BastionService
	stdin          io.Reader
	stdout         io.Writer
	stderr         io.Writer
}

func NewUpCommand(svc bastion.BastionService, stdin io.Reader, stdout io.Writer, stderr io.Writer) Command[*UpInput] {
	return &upCommand{
		bastionService: svc,
		stdin:          stdin,
		stdout:         stdout,
		stderr:         stderr,
	}
}

func (c *upCommand) Run(ctx context.Context, input *UpInput) error {

	out, err := c.bastionService.DeployBastion(ctx, &bastion.DeployBastionInput{
		BastionName: input.BastionName,
		Owner:       input.Owner,
		SubnetID:    input.SubnetID,
		VPCID:       input.VPCID,
	})

	if err != nil {
		return err
	}

	return json.NewEncoder(c.stdout).Encode(struct {
		StackName        string `json:"stackName"`
		InstanceID       string `json:"instanceId"`
		AvailabilityZone string `json:"availabilityZone"`
	}{
		StackName:        out.StackName,
		InstanceID:       out.InstanceID,
		AvailabilityZone: out.AvailabilityZone,
	})
}
