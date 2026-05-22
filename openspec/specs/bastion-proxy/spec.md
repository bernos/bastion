# bastion-proxy Specification

## Purpose
TBD - created by archiving change add-ssh-proxy-commands. Update Purpose after archive.
## Requirements
### Requirement: `proxy` command requires `--name` and `--region`
The `bastion proxy` command SHALL require both `--name` (or `BASTION_NAME` / `name` in config) and `--region` (or `BASTION_REGION` / `region` in config). If either is missing, the command SHALL exit with a non-zero code and a usage error before making any AWS API calls.

#### Scenario: Missing name flag
- **WHEN** `bastion proxy` is invoked without `--name` and no `BASTION_NAME` env var or config key is set
- **THEN** the command exits with a non-zero code and prints a usage error identifying the missing flag

#### Scenario: Missing region flag
- **WHEN** `bastion proxy` is invoked without `--region` and no `BASTION_REGION` env var or config key is set
- **THEN** the command exits with a non-zero code and prints a usage error identifying the missing flag

### Requirement: `proxy` command accepts an optional `--port` flag defaulting to 1080
The `bastion proxy` command SHALL accept a `--port` flag (integer). If not supplied, it SHALL default to `1080`. The port value is used as the local SOCKS5 listener port passed to `ssh -D`.

#### Scenario: Default port used when flag omitted
- **WHEN** `bastion proxy` is invoked without `--port`
- **THEN** the SOCKS5 proxy listens on local port 1080

#### Scenario: Custom port used when flag supplied
- **WHEN** `bastion proxy --port 8080` is invoked
- **THEN** the SOCKS5 proxy listens on local port 8080

### Requirement: `proxy` command verifies external dependencies before connecting
The `bastion proxy` command SHALL verify that both `aws` CLI and `session-manager-plugin` are available on `PATH` before making any AWS API calls. If either is missing, the command SHALL exit with a non-zero code and print a message indicating which binary is missing and how to install it.

#### Scenario: `aws` CLI not found on PATH
- **WHEN** `bastion proxy` is invoked and the `aws` binary is not found on `PATH`
- **THEN** the command exits with a non-zero code and prints an error referencing the missing `aws` CLI

#### Scenario: `session-manager-plugin` not found on PATH
- **WHEN** `bastion proxy` is invoked and `session-manager-plugin` is not found on `PATH`
- **THEN** the command exits with a non-zero code and prints an error referencing the missing plugin and its installation URL

### Requirement: `proxy` command resolves the bastion instance via `DescribeBastion`
The `bastion proxy` command SHALL call `BastionService.DescribeBastion` with the supplied name and region. If the stack is not found or returns an error, the command SHALL exit with a non-zero code and print the error.

#### Scenario: Bastion stack not found
- **WHEN** `bastion proxy --name foo --region ap-southeast-2` is invoked and no stack `foo-stack` exists
- **THEN** the command exits with a non-zero code and prints an error indicating the stack was not found

#### Scenario: Bastion instance resolved successfully
- **WHEN** the stack exists with `InstanceId` and `AvailabilityZone` outputs
- **THEN** `DescribeBastion` returns the instance ID and AZ and the command proceeds

### Requirement: `proxy` command waits for SSM agent readiness before connecting
Before uploading a key or starting the proxy, `bastion proxy` SHALL call `BastionService.WaitForSSMReady` with the resolved instance ID and region. If the agent does not come online within the timeout, the command SHALL exit with a non-zero code.

#### Scenario: SSM agent comes online within timeout
- **WHEN** the SSM agent reports `PingStatus == Online` within the polling window
- **THEN** the command proceeds to key upload

#### Scenario: SSM agent timeout
- **WHEN** the SSM agent does not come online within the timeout
- **THEN** the command exits with a non-zero code and prints the timeout error

### Requirement: `proxy` command generates an ephemeral ed25519 key pair and uploads the public key
The `bastion proxy` command SHALL generate an ed25519 key pair in memory. The public key SHALL be uploaded via `ec2instanceconnect:SendSSHPublicKey` with `InstanceOSUser` set to `ec2-user`. The private key SHALL be written to a temporary file with mode `0600`.

#### Scenario: Key uploaded successfully
- **WHEN** `SendSSHPublicKey` succeeds
- **THEN** the command prints the proxy status line and proceeds to start the `ssh` subprocess

#### Scenario: Key upload failure
- **WHEN** `SendSSHPublicKey` returns an error
- **THEN** the command exits with a non-zero code and prints the error without starting `ssh`

### Requirement: `proxy` command prints a status line before starting the proxy
Before starting the `ssh` subprocess, `bastion proxy` SHALL print a status line to stderr indicating the local port and instance ID in use.

#### Scenario: Status line printed before proxy starts
- **WHEN** key upload succeeds and the proxy is about to start
- **THEN** a line is printed to stderr containing the local port number and the instance ID, before `ssh` is started

### Requirement: `proxy` command runs `ssh` with SOCKS5 and SSM ProxyCommand flags
The `bastion proxy` command SHALL run the system `ssh` binary via `exec.Command` with:
- `-i <tempKeyFile>` pointing to the ephemeral private key
- `-D <port>` to open a local SOCKS5 listener on the specified port
- `-N` to suppress remote command execution
- `-o StrictHostKeyChecking=no`
- `-o UserKnownHostsFile=/dev/null`
- `-o ProxyCommand=aws ssm start-session --target %h --document-name AWS-StartSSHSession --parameters portNumber=22 --region <region>`
- Target: `ec2-user@<instanceID>`

`cmd.Stdin`, `cmd.Stdout`, and `cmd.Stderr` SHALL be set to `os.Stdin`, `os.Stdout`, and `os.Stderr`.

#### Scenario: Proxy runs until interrupted
- **WHEN** the proxy is running and the user sends SIGINT (Ctrl-C)
- **THEN** the signal is forwarded to the `ssh` subprocess, the proxy shuts down cleanly, and the temp key file is removed

#### Scenario: Proxy exits with non-zero code
- **WHEN** `ssh` exits with a non-zero exit code
- **THEN** `bastion proxy` exits with the same non-zero code and the temp key file is removed

### Requirement: `proxy` command forwards SIGINT and SIGTERM to the `ssh` subprocess
A goroutine SHALL listen for `os.Interrupt` and `syscall.SIGTERM` and forward each received signal to the `ssh` subprocess via `cmd.Process.Signal`. This ensures the proxy tears down cleanly when the parent process is interrupted or terminated.

#### Scenario: SIGINT forwarded to ssh
- **WHEN** the OS sends SIGINT to `bastion proxy`
- **THEN** the signal is forwarded to the `ssh` child process

### Requirement: `proxy` command removes the temporary private key file after the proxy ends
The temp key file SHALL be removed via `os.Remove` after `cmd.Wait()` returns, regardless of the `ssh` exit code.

#### Scenario: Key file cleaned up after proxy ends
- **WHEN** the proxy exits for any reason
- **THEN** the temp key file no longer exists on disk

