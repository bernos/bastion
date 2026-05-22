## Why

The bastion CLI can deploy and tear down a bastion host, but provides no way to actually connect to one. Users must manually construct `ssh` invocations with the correct ProxyCommand, upload keys by hand, and look up instance IDs — defeating the purpose of a purpose-built tool. Connecting to the bastion should be a first-class operation.

## What Changes

- **New**: `bastion ssh` command — uploads an ephemeral key, waits for SSM readiness, and opens an interactive SSH session to the bastion via SSM Session Manager as ProxyCommand
- **New**: `bastion proxy` command — same flow, but starts a local SOCKS5 proxy (`ssh -D`) instead of an interactive shell; defaults to port 1080
- **New**: `BastionService.DescribeBastion` — resolves a bastion name to its instance ID, AZ, and region by reading CloudFormation stack outputs
- **New**: `BastionService.WaitForSSMReady` — promoted from a private function to a public interface method, accepting a `WaitForSSMReadyInput` struct
- **New**: `CloudFormationService.GetStackOutputs` — reads outputs from an existing stack without deploying
- **BREAKING**: `--region` flag added as a required parameter to all commands (`up`, `down`, `ssh`, `proxy`); region is no longer resolved from the AWS SDK default chain
- **BREAKING**: `--public-key-path` flag removed from `up`; `up` becomes a pure CloudFormation deploy command
- **BREAKING**: `DeployBastionInput.PublicKeyContent` removed; `Region` added
- **BREAKING**: `DeleteBastionInput.Region` added

## Capabilities

### New Capabilities

- `bastion-ssh`: `bastion ssh` command — ephemeral key generation, SSM readiness wait, EC2 Instance Connect key upload, interactive SSH session via SSM ProxyCommand
- `bastion-proxy`: `bastion proxy` command — same connection flow as `bastion ssh`, but establishes a local SOCKS5 proxy via `ssh -D` instead of an interactive shell
- `explicit-region`: AWS region is a required explicit input to all commands and service method input structs; no ambient SDK region resolution

### Modified Capabilities

- `ssh-key-upload`: all existing requirements removed — key upload is no longer part of the `up` command; the capability is superseded by the connection flow in `bastion-ssh` and `bastion-proxy`
- `bastion-down`: gains a required `--region` flag (and `BASTION_REGION` env var / `region` config key)

## Impact

- `cmd/bastion/up.go` — remove `--public-key-path` flag, key validation, and `PublicKeyContent` field usage
- `cmd/bastion/down.go` — add `--region` flag
- `cmd/bastion/ssh.go` — new file
- `cmd/bastion/proxy.go` — new file
- `internal/bastion/bastion.go` — add `DescribeBastion`, promote `WaitForSSMReady`, strip key upload from `DeployBastion`
- `internal/config/config.go` — add `Region`, remove `PublicKeyPath`
- `internal/dependencies/dependencies.go` — pass `cfg.Region` to `LoadDefaultConfig`
- `pkg/aws/cloudformationservice/service.go` — add `GetStackOutputs`
- New dependency: `golang.org/x/crypto/ed25519` (ephemeral key generation, likely already transitive via `golang.org/x/crypto/ssh`)
- External runtime dependency: `aws` CLI + `session-manager-plugin` must be present on the operator's machine
