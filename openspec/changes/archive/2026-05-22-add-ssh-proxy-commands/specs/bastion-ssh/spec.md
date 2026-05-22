## ADDED Requirements

### Requirement: `ssh` command requires `--name` and `--region`
The `bastion ssh` command SHALL require both `--name` (or `BASTION_NAME` / `name` in config) and `--region` (or `BASTION_REGION` / `region` in config). If either is missing, the command SHALL exit with a non-zero code and a usage error before making any AWS API calls.

#### Scenario: Missing name flag
- **WHEN** `bastion ssh` is invoked without `--name` and no `BASTION_NAME` env var or config key is set
- **THEN** the command exits with a non-zero code and prints a usage error identifying the missing flag

#### Scenario: Missing region flag
- **WHEN** `bastion ssh` is invoked without `--region` and no `BASTION_REGION` env var or config key is set
- **THEN** the command exits with a non-zero code and prints a usage error identifying the missing flag

### Requirement: `ssh` command verifies external dependencies before connecting
The `bastion ssh` command SHALL verify that both `aws` CLI and `session-manager-plugin` are available on `PATH` before making any AWS API calls. If either is missing, the command SHALL exit with a non-zero code and print a message indicating which binary is missing and how to install it.

#### Scenario: `aws` CLI not found on PATH
- **WHEN** `bastion ssh` is invoked and the `aws` binary is not found on `PATH`
- **THEN** the command exits with a non-zero code and prints an error referencing the missing `aws` CLI

#### Scenario: `session-manager-plugin` not found on PATH
- **WHEN** `bastion ssh` is invoked and `session-manager-plugin` is not found on `PATH`
- **THEN** the command exits with a non-zero code and prints an error referencing the missing plugin and its installation URL

#### Scenario: Both dependencies present
- **WHEN** both `aws` and `session-manager-plugin` are on `PATH`
- **THEN** the command proceeds to the connection flow without error

### Requirement: `ssh` command resolves the bastion instance via `DescribeBastion`
The `bastion ssh` command SHALL call `BastionService.DescribeBastion` with the supplied name and region to obtain the instance ID and availability zone. If `DescribeBastion` returns an error (e.g. stack not found), the command SHALL exit with a non-zero code and print the error.

#### Scenario: Bastion stack not found
- **WHEN** `bastion ssh --name foo --region ap-southeast-2` is invoked and no CloudFormation stack named `foo-stack` exists
- **THEN** the command exits with a non-zero code and prints an error indicating the stack was not found

#### Scenario: Bastion instance resolved successfully
- **WHEN** `bastion ssh --name foo --region ap-southeast-2` is invoked and `foo-stack` exists with `InstanceId` and `AvailabilityZone` outputs
- **THEN** `DescribeBastion` returns the instance ID and AZ and the command proceeds

### Requirement: `ssh` command waits for SSM agent readiness before connecting
Before uploading a key or opening a connection, `bastion ssh` SHALL call `BastionService.WaitForSSMReady` with the resolved instance ID and region. If the SSM agent does not come online within the timeout, the command SHALL exit with a non-zero code and print the timeout error.

#### Scenario: SSM agent comes online within timeout
- **WHEN** the SSM agent on the target instance reports `PingStatus == Online` within the polling window
- **THEN** the command proceeds to key upload

#### Scenario: SSM agent timeout
- **WHEN** the SSM agent does not come online within the `WaitForSSMReady` timeout
- **THEN** the command exits with a non-zero code and prints the timeout error

### Requirement: `ssh` command generates an ephemeral ed25519 key pair and uploads the public key
The `bastion ssh` command SHALL generate an ed25519 key pair in memory. The public key SHALL be uploaded via `ec2instanceconnect:SendSSHPublicKey` with `InstanceOSUser` set to `ec2-user`. The private key SHALL be written to a temporary file with mode `0600`.

#### Scenario: Key uploaded successfully
- **WHEN** `SendSSHPublicKey` succeeds
- **THEN** the command proceeds to exec the `ssh` binary

#### Scenario: Key upload failure
- **WHEN** `SendSSHPublicKey` returns an error
- **THEN** the command exits with a non-zero code and prints the error without starting `ssh`

### Requirement: `ssh` command execs the system `ssh` binary with SSM as ProxyCommand
The `bastion ssh` command SHALL run the system `ssh` binary via `exec.Command` with:
- `-i <tempKeyFile>` pointing to the ephemeral private key
- `-o StrictHostKeyChecking=no`
- `-o UserKnownHostsFile=/dev/null`
- `-o ProxyCommand=aws ssm start-session --target %h --document-name AWS-StartSSHSession --parameters portNumber=22 --region <region>`
- Target: `ec2-user@<instanceID>`

`cmd.Stdin`, `cmd.Stdout`, and `cmd.Stderr` SHALL be set to `os.Stdin`, `os.Stdout`, and `os.Stderr` respectively so the SSH session has access to the operator's terminal.

#### Scenario: SSH session runs and exits cleanly
- **WHEN** `ssh` exits with code 0
- **THEN** `bastion ssh` exits with code 0 and the temp key file is removed

#### Scenario: SSH exits with non-zero code
- **WHEN** `ssh` exits with a non-zero exit code
- **THEN** `bastion ssh` exits with the same non-zero code and the temp key file is removed

### Requirement: `ssh` command removes the temporary private key file after the session ends
The temp key file SHALL be removed via `os.Remove` after `cmd.Wait()` returns, regardless of the `ssh` exit code.

#### Scenario: Key file cleaned up after normal exit
- **WHEN** the SSH session ends for any reason
- **THEN** the temp key file no longer exists on disk
