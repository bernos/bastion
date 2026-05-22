## 1. Config and dependencies

- [x] 1.1 Add `Region string` field to `internal/config/config.go` and remove `PublicKeyPath string`
- [x] 1.2 Update `dependencies.go` to store `cfg *config.Config` and pass `cfg.Region` to `awsconfig.LoadDefaultConfig` via `awsconfig.WithRegion`

## 2. CloudFormation service — read path

- [x] 2.1 Add `GetStackOutputs(ctx context.Context, stackName string) (map[string]string, error)` to `CloudFormationService` interface and `cloudformationservice` package
- [x] 2.2 Add unit tests for `GetStackOutputs` using the existing mock pattern

## 3. BastionService — refactor and new methods

- [x] 3.1 Remove `PublicKeyContent` from `DeployBastionInput`; add `Region string`
- [x] 3.2 Remove SSM wait and `SendSSHPublicKey` call from `DeployBastion`; remove `waitForSSMReady` private method
- [x] 3.3 Add `Region string` to `DeleteBastionInput`
- [x] 3.4 Add `DescribeBastion(ctx context.Context, input *DescribeBastionInput) (*DescribeBastionOutput, error)` to `BastionService` interface and implementation (`DescribeBastionInput`: `BastionName`, `Region`; `DescribeBastionOutput`: `InstanceID`, `AvailabilityZone`, `Region`)
- [x] 3.5 Add `WaitForSSMReady(ctx context.Context, input *WaitForSSMReadyInput) error` to `BastionService` interface and implementation (`WaitForSSMReadyInput`: `InstanceID`, `Region`)
- [x] 3.6 Update `BastionService` mock to include `DescribeBastion` and `WaitForSSMReady`
- [x] 3.7 Update unit tests for `DeployBastion` to reflect removal of key upload and SSM wait
- [x] 3.8 Add unit tests for `DescribeBastion` and `WaitForSSMReady`

## 4. `up` command cleanup

- [x] 4.1 Remove `--public-key-path` flag, key validation logic, and `PublicKeyContent` usage from `cmd/bastion/up.go`
- [x] 4.2 Add `--region` required flag to `up` command and wire into `Config.Region`
- [x] 4.3 Update `up` command tests to reflect removed flag and added region

## 5. `down` command update

- [x] 5.1 Add `--region` required flag to `down` command and wire into `Config.Region`
- [x] 5.2 Pass `Region` in `DeleteBastionInput` when calling `BastionService.DeleteBastion`
- [x] 5.3 Update `down` command tests

## 6. Shared connection helper

- [x] 6.1 Implement ephemeral ed25519 key generation (returns private key PEM bytes and SSH-format public key string)
- [x] 6.2 Implement `checkDependencies() error` — uses `exec.LookPath` to verify `aws` and `session-manager-plugin` are on PATH, returns a clear error with install hints if either is missing
- [x] 6.3 Implement `buildSSHArgs(instanceID, region, keyFile string, extraArgs []string) []string` — assembles the full ssh argument list with ProxyCommand, StrictHostKeyChecking, UserKnownHostsFile, and target

## 7. `ssh` command

- [x] 7.1 Create `cmd/bastion/ssh.go` with `NewSSHCommand(cfg *config.Config) (*cobra.Command, error)`
- [x] 7.2 Wire `--name` (required) and `--region` (required) flags
- [x] 7.3 Implement `RunE`: check dependencies → DescribeBastion → WaitForSSMReady → generate key pair → upload public key → write private key to temp file (0600) → exec.Command(ssh, args) with stdin/stdout/stderr passthrough → Wait → Remove temp file
- [x] 7.4 Register `ssh` command in `cmd/bastion/main.go`
- [x] 7.5 Write unit tests for the ssh command RunE using mocked BastionService

## 8. `proxy` command

- [x] 8.1 Create `cmd/bastion/proxy.go` with `NewProxyCommand(cfg *config.Config) (*cobra.Command, error)`
- [x] 8.2 Wire `--name` (required), `--region` (required), and `--port` (default 1080) flags
- [x] 8.3 Implement `RunE`: check dependencies → DescribeBastion → WaitForSSMReady → generate key pair → upload public key → write private key to temp file (0600) → print status line to stderr → exec.Command(ssh -D <port> -N, args) with stdin/stdout/stderr passthrough → goroutine to forward SIGINT/SIGTERM → Wait → Remove temp file
- [x] 8.4 Register `proxy` command in `cmd/bastion/main.go`
- [x] 8.5 Write unit tests for the proxy command RunE using mocked BastionService

## 9. Integration and validation

- [x] 9.1 Run `make lint` and fix any issues
- [x] 9.2 Run `make test` and confirm all unit tests pass
- [x] 9.3 Run `make test-integration` and confirm CloudFormation integration tests pass
- [ ] 9.4 Manually verify `bastion ssh` opens an interactive session against a live bastion
- [ ] 9.5 Manually verify `bastion proxy` starts a SOCKS5 proxy and forwards traffic through the bastion
