## Why

There is no way to tear down a bastion host once it's deployed. Users must manually delete the CloudFormation stack via the AWS console or CLI, which is error-prone and inconsistent with the `bastion up` experience.

## What Changes

- Add `bastion down` CLI command that deletes the CloudFormation stack for a named bastion host and waits for deletion to complete.
- The command resolves `--name` via the same config precedence as `up`: `--name` CLI flag → `BASTION_NAME` env var → `name` key in config file (`./config.yaml` or `~/.config/bastion/config.yaml`).

## Capabilities

### New Capabilities

- `bastion-down`: The `bastion down` command — accepts `--name`, resolves the stack name, deletes the CloudFormation stack, waits for completion, and reports success or failure.

### Modified Capabilities

<!-- none -->

## Impact

- `cmd/bastion/down.go` — implement `NewDownCommand` (currently a stub)
- `internal/bastion/bastion.go` — add `DeleteBastion` method to `BastionService`
- `pkg/aws/cloudformationservice/service.go` — may need a `DeleteStack` / `WaitForStackDeletion` operation
- No breaking changes; no new dependencies
