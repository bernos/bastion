## Why

After deploying a bastion, operators need a way to SSH into it via EC2 Instance Connect. The `SendSSHPublicKey` API uploads a temporary public key to the instance, but this step currently requires a manual out-of-band call. Making it part of `bastion up` closes that gap and keeps the deploy-and-connect flow self-contained.

## What Changes

- New `public-key-path` configuration parameter on the `up` command, backed by the standard resolution chain: `--public-key-path` CLI flag → `BASTION_PUBLIC_KEY_PATH` env var → `public-key-path` in config YAML.
- If `public-key-path` is set, the file is validated before the deploy: the command checks for PEM private key markers and attempts to parse the content as an authorized-keys format public key, rejecting it early with a clear error if either check fails.
- After a successful deploy the `up` command waits until the SSM agent on the instance reports `Online` via `ssm:DescribeInstanceInformation`, then calls `ec2instanceconnect:SendSSHPublicKey`.
- The SSM readiness wait and key upload are both gated on `public-key-path` being set; omitting it preserves the current deploy-only behaviour.

## Capabilities

### New Capabilities

- `ssh-key-upload`: Upload a public SSH key to the deployed bastion instance via EC2 Instance Connect (`SendSSHPublicKey`) as part of the `up` flow.

### Modified Capabilities

- `bastion-cfn-stack`: `DeployBastionOutput` already exposes `InstanceID` and `AvailabilityZone`, which the new capability depends on — no requirement changes needed.

## Impact

- `cmd/bastion/up.go` — new `--public-key-path` flag, call key-upload after deploy
- `internal/config/config.go` — new `PublicKeyPath` field
- `internal/bastion/bastion.go` — extend `BastionService`/`DeployBastionInput` or add new method
- `go.mod` / `go.sum` — `aws-sdk-go-v2/service/ec2instanceconnect` added; `aws-sdk-go-v2/service/ssm` and `golang.org/x/crypto` already indirect deps, promoted to direct use; both AWS clients injected into `BastionService` via minimal local interfaces
