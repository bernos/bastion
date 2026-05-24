## Why

The connection logic in `cmd/bastion/connect.go` and `ssh.go` — key generation, EC2 Instance Connect upload, SSH arg assembly, and dependency checking — lives in the CLI layer where it can't be tested without running real AWS calls or spawning processes. Moving it into a dedicated service makes the orchestration independently testable and removes business logic from the command layer.

## What Changes

- **New** `internal/connect/` package exposing a `ConnectService` interface
- `ConnectService.Prepare` handles: dependency checking, `DescribeBastion`, `WaitForSSMReady`, ephemeral key generation, EC2IC key upload, temp file creation, and SSH arg assembly — returning a `*Connection` the caller uses to exec `ssh`
- `cmd/bastion/connect.go` and the shared logic in `ssh.go` (`runConnect`, `buildSSHArgs`, `generateEphemeralKeyPair`, `checkSSMDependencies`, `ec2icSender` interface) are removed from the `cmd` package
- `NewSSHCommand` and `NewProxyCommand` are updated to construct a `ConnectService` via `dependencies` and call `Prepare`, then exec `ssh` with the returned args
- Unit tests for `ConnectService.Prepare` are added; `connect_test.go` in `cmd` is updated to use the new service mock

## Capabilities

### New Capabilities
- `connect-service`: The `ConnectService` interface and `Prepare` method that orchestrates all pre-connection steps and returns ready-to-use SSH arguments and a temp key path

### Modified Capabilities
<!-- No external behavior changes; bastion-ssh and bastion-proxy specs are unchanged -->

## Impact

- `cmd/bastion/connect.go` is removed or gutted
- `cmd/bastion/ssh.go` loses `runConnect`, `buildSSHArgs`, `generateEphemeralKeyPair`, `checkSSMDependencies`, `ec2icSender`
- New package: `internal/connect/` (service interface + implementation)
- `internal/dependencies/dependencies.go` gains a `ConnectService` factory method
- Existing `connect_test.go` tests move into the new package; cmd-layer tests are simplified to mock `ConnectService`
- No changes to external CLI behaviour or AWS API calls
