# connect-service Specification

## Purpose
TBD - created by archiving change extract-connect-service. Update Purpose after archive.
## Requirements
### Requirement: `ConnectService` exposes a `CheckDependencies` method
`ConnectService` SHALL expose a `CheckDependencies() error` method that verifies both the `aws` CLI and `session-manager-plugin` binaries are available on `PATH`. If either is missing the method SHALL return an error containing the name of the missing binary and a URL to its installation documentation.

#### Scenario: Both binaries present
- **WHEN** `aws` and `session-manager-plugin` are both found on `PATH`
- **THEN** `CheckDependencies` returns nil

#### Scenario: `aws` CLI missing
- **WHEN** `aws` is not found on `PATH`
- **THEN** `CheckDependencies` returns a non-nil error referencing the missing `aws` CLI

#### Scenario: `session-manager-plugin` missing
- **WHEN** `session-manager-plugin` is not found on `PATH`
- **THEN** `CheckDependencies` returns a non-nil error referencing the missing plugin and its install URL

### Requirement: `ConnectService` exposes a `Prepare` method that returns exec-ready SSH args
`ConnectService` SHALL expose a `Prepare(ctx context.Context, input *PrepareInput) (*Connection, error)` method. `PrepareInput` SHALL contain `BastionName string`, `Region string`, `OSUser string` (defaults to `"ec2-user"` when empty), `ExtraSSHArgs []string`, and `PrivateKeyFile PrivateKeyWriter` (a caller-created file that receives the private key). `Connection` SHALL contain `SSHArgs []string`.

`Prepare` SHALL perform the following steps in order:
1. Call `BastionService.DescribeBastion` to resolve the instance ID and availability zone
2. Call `BastionService.WaitForSSMReady` with the resolved instance ID
3. Generate an ephemeral ed25519 key pair
4. Upload the public key via `ec2instanceconnect:SendSSHPublicKey` using `OSUser` as `InstanceOSUser`
5. Write the private key to `input.PrivateKeyFile` with mode `0600` and close it
6. Assemble and return exec-ready SSH arguments

If any step fails, `Prepare` SHALL return a non-nil error and no `Connection`. The caller is responsible for creating `PrivateKeyFile` before calling `Prepare` and removing it after the SSH process exits.

#### Scenario: Successful prepare
- **WHEN** all steps succeed
- **THEN** `Prepare` returns a `*Connection` with a non-empty `SSHArgs` slice and a nil error

#### Scenario: `DescribeBastion` fails
- **WHEN** `BastionService.DescribeBastion` returns an error (e.g. stack not found)
- **THEN** `Prepare` returns nil and the error from `DescribeBastion`; no key is generated or uploaded

#### Scenario: `WaitForSSMReady` times out
- **WHEN** `BastionService.WaitForSSMReady` returns a timeout error
- **THEN** `Prepare` returns nil and the timeout error; no key is generated or uploaded

#### Scenario: Key upload fails
- **WHEN** `SendSSHPublicKey` returns an error
- **THEN** `Prepare` returns nil and the upload error

#### Scenario: `OSUser` defaults to `ec2-user`
- **WHEN** `PrepareInput.OSUser` is empty
- **THEN** `InstanceOSUser` sent to `SendSSHPublicKey` is `"ec2-user"`

### Requirement: `Prepare` assembles SSH args with SSM as ProxyCommand
The `SSHArgs` returned by `Prepare` SHALL include:
- `-i <keyPath>` pointing to the temp private key file
- `-o StrictHostKeyChecking=no`
- `-o UserKnownHostsFile=/dev/null`
- `-o ProxyCommand=aws ssm start-session --target %h --document-name AWS-StartSSHSession --parameters portNumber=22 --region <region>`
- Any args from `PrepareInput.ExtraSSHArgs` before the target
- Final argument: `<OSUser>@<instanceID>`

#### Scenario: SSH mode args (no extra args)
- **WHEN** `PrepareInput.ExtraSSHArgs` is nil or empty
- **THEN** `SSHArgs` contains the standard flags and `<OSUser>@<instanceID>` as the final element

#### Scenario: Proxy mode args
- **WHEN** `PrepareInput.ExtraSSHArgs` is `["-D", "1080", "-N"]`
- **THEN** `SSHArgs` contains those args between the fixed flags and the target

#### Scenario: Region embedded in ProxyCommand
- **WHEN** `PrepareInput.Region` is `"ap-southeast-2"`
- **THEN** the `ProxyCommand` option in `SSHArgs` contains `--region ap-southeast-2`

