## 1. CloudFormation Service Layer

- [x] 1.1 Add `DeleteStack` to the `CloudFormationClient` interface in `pkg/aws/cloudformationservice/service.go`
- [x] 1.2 Add `DeleteStack(ctx context.Context, stackName string) error` to `CloudFormationService` interface
- [x] 1.3 Implement `cloudFormationService.DeleteStack`: call `client.DeleteStack`, then block on `StackDeleteCompleteWaiter`
- [x] 1.4 Add unit tests for `DeleteStack` in `service_test.go` (success and not-found cases)

## 2. Bastion Service Layer

- [x] 2.1 Add `DeleteBastionInput` struct and `DeleteBastion` method signature to `BastionService` interface in `internal/bastion/bastion.go`
- [x] 2.2 Implement `bastionService.DeleteBastion`: derive stack name, check existence via `StackExists`, delegate to `CloudFormationService.DeleteStack`
- [x] 2.3 Add unit tests for `DeleteBastion` in `bastion_test.go` (success, stack not found)

## 3. CLI Command

- [x] 3.1 Implement `NewDownCommand` in `cmd/bastion/down.go`: declare `--name` flag via `cmd.Flags().String(...)` (Viper picks it up automatically via `PersistentPreRunE`), read `cfg.Name`, call `BastionService.DeleteBastion`, print JSON `{"stackName": "..."}` to stdout on success
- [x] 3.2 Add unit tests for `NewDownCommand` covering success and error paths

## 4. Integration Test

- [x] 4.1 Add an integration test in `internal/bastion/bastion_integration_test.go` that deploys a stack with `DeployBastion` then tears it down with `DeleteBastion` against LocalStack
