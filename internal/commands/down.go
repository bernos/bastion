package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/bernos/bastion/internal/bastion"
)

type DownInput struct {
	BastionName string
}

type downCommand struct {
	bastionService bastion.BastionService
	stdin          io.Reader
	stdout         io.Writer
	stderr         io.Writer
}

func NewDownCommand(svc bastion.BastionService, stdin io.Reader, stdout io.Writer, stderr io.Writer) *downCommand {
	return &downCommand{
		bastionService: svc,
		stdin:          stdin,
		stdout:         stdout,
		stderr:         stderr,
	}
}

func (c *downCommand) Run(ctx context.Context, input *DownInput) error {

	if err := c.bastionService.DeleteBastion(ctx, &bastion.DeleteBastionInput{
		BastionName: input.BastionName,
	}); err != nil {
		return err
	}

	return json.NewEncoder(c.stdout).Encode(struct {
		StackName string `json:"stackName"`
	}{
		StackName: fmt.Sprintf("%s-stack", input.BastionName),
	})

}
