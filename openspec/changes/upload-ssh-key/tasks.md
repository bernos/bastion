## 1. Feature branch

- [x] 1.1 Create and check out `feature/upload-ssh-key` branch

## 2. SDK dependency

- [x] 2.1 Add `github.com/aws/aws-sdk-go-v2/service/ec2instanceconnect` to `go.mod` / `go.sum` (`go get`)

## 3. Config

- [x] 3.1 Add `PublicKeyPath string` field to `Config` struct in `internal/config/config.go` with `mapstructure:"public-key-path"` tag
- [x] 3.2 Add unit test covering `public-key-path` resolution from flag, env var, and YAML in `internal/config/config_test.go`

## 4. BastionService

- [x] 4.1 Add `PublicKeyContent string` field to `DeployBastionInput` in `internal/bastion/bastion.go`
- [x] 4.2 Define unexported `ec2InstanceConnectClient` interface (single `SendSSHPublicKey` method) and `ssmClient` interface (single `DescribeInstanceInformation` method) in `internal/bastion/bastion.go`; add both fields to `bastionService` and update `NewBastionService` to accept them
- [x] 4.3 Add private `waitForSSMReady(ctx, instanceID)` helper on `bastionService` that polls `DescribeInstanceInformation` every 10 seconds, returns nil when `PingStatus == Online`, and returns an error after 2 minutes
- [x] 4.4 In `DeployBastion`, when `PublicKeyContent` is non-empty: call `waitForSSMReady` then call `SendSSHPublicKey`; propagate any error from either call
- [x] 4.5 Update unit tests in `internal/bastion/bastion_test.go` — add mocks for both interfaces and cases covering: SSM online before timeout → key uploaded; SSM timeout → error returned, key not uploaded; no key content → neither SSM nor EC2IC called; EC2IC error propagated

## 5. Dependencies wiring

- [x] 5.1 In `internal/dependencies/dependencies.go`, construct `*ec2instanceconnect.Client` and `*ssm.Client` from the AWS config and pass both to `NewBastionService`

## 6. up command

- [x] 6.1 Add `--public-key-path` flag (non-required) to `NewUpCommand` in `cmd/bastion/up.go` and bind it via Viper
- [x] 6.2 At the start of `RunE`, if `cfg.PublicKeyPath` is non-empty, read the file contents and return an error immediately if the file cannot be read
- [x] 6.3 Validate the file content: check for `PRIVATE KEY` substring (return a specific error) then call `golang.org/x/crypto/ssh.ParseAuthorizedKey` (return a parse error if it fails)
- [x] 6.4 Pass the validated key contents as `PublicKeyContent` in `DeployBastionInput`

## 7. Verification

- [x] 7.1 Run `make lint` and fix any issues
- [x] 7.2 Run `make test` and confirm all unit tests pass
- [x] 7.3 Run `make test-integration` and confirm existing integration tests still pass
