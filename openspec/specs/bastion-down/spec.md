## Purpose
The `bastion down` command tears down a named bastion host by deleting its CloudFormation stack and waiting for deletion to complete.
## Requirements
### Requirement: `down` command deletes the bastion CloudFormation stack
The `bastion down` command SHALL resolve the bastion name from the first of: `--name` CLI flag, `BASTION_NAME` environment variable, or `name` key in a config file (`./config.yaml` or `~/.config/bastion/config.yaml`). It SHALL derive the stack name as `<name>-stack`, delete the CloudFormation stack, and wait for deletion to complete before exiting.

#### Scenario: Successful deletion
- **WHEN** `bastion down --name <name>` is run and the stack `<name>-stack` exists
- **THEN** the CloudFormation stack is deleted, the command waits for `DELETE_COMPLETE`, and exits with code 0

#### Scenario: Name resolved from config file
- **WHEN** `bastion down` is run without `--name` and a config file contains `name: <name>`
- **THEN** the command uses `<name>` to derive the stack name and proceeds with deletion

#### Scenario: Name resolved from environment variable
- **WHEN** `bastion down` is run without `--name` and `BASTION_NAME=<name>` is set in the environment
- **THEN** the command uses `<name>` to derive the stack name and proceeds with deletion

#### Scenario: Stack does not exist
- **WHEN** `bastion down --name <name>` is run and no stack named `<name>-stack` exists
- **THEN** the command exits with a non-zero code and prints an error message identifying the missing stack

### Requirement: `down` command prints JSON output on success
The `down` command SHALL write a JSON object to stdout on success containing a `stackName` field with the name of the deleted stack.

#### Scenario: JSON printed after successful deletion
- **WHEN** `bastion down` completes successfully
- **THEN** stdout contains a single JSON object with a string field `stackName`

### Requirement: `CloudFormationService` exposes a `DeleteStack` operation
`CloudFormationService` SHALL expose a `DeleteStack(ctx context.Context, stackName string) error` method that initiates stack deletion and blocks until the stack reaches `DELETE_COMPLETE` or returns an error if deletion fails.

#### Scenario: Stack deleted and waiter completes
- **WHEN** `CloudFormationService.DeleteStack` is called with the name of an existing stack
- **THEN** the underlying `cloudformation.DeleteStack` API is called and the method blocks until `StackDeleteCompleteWaiter` resolves successfully

#### Scenario: Stack does not exist
- **WHEN** `CloudFormationService.DeleteStack` is called with a stack name that does not exist
- **THEN** the method returns an error

### Requirement: `BastionService` exposes a `DeleteBastion` operation
`BastionService` SHALL expose a `DeleteBastion(ctx context.Context, input *DeleteBastionInput) error` method. `DeleteBastionInput` SHALL contain `BastionName string` and `Region string` fields. The method SHALL derive the stack name as `<BastionName>-stack`, verify it exists, and delegate deletion to `CloudFormationService.DeleteStack`.

#### Scenario: Bastion deleted successfully
- **WHEN** `BastionService.DeleteBastion` is called with a `BastionName` whose corresponding stack exists and a valid `Region`
- **THEN** `CloudFormationService.DeleteStack` is called with `<BastionName>-stack` and the method returns nil on success

#### Scenario: Stack does not exist
- **WHEN** `BastionService.DeleteBastion` is called with a `BastionName` whose stack does not exist
- **THEN** the method returns a descriptive error without calling `CloudFormationService.DeleteStack`

### Requirement: `down` command requires `--region`
The `bastion down` command SHALL require that region is resolvable from `--region`, `BASTION_REGION`, or the `region` config key. If region cannot be resolved, the command SHALL exit with a non-zero code and a usage error before making any AWS API calls.

#### Scenario: Region resolved from CLI flag
- **WHEN** `bastion down --name foo --region ap-southeast-2` is invoked
- **THEN** `ap-southeast-2` is used for all AWS operations in that invocation

#### Scenario: Region resolved from environment variable
- **WHEN** `BASTION_REGION=us-east-1` is set and `bastion down --name foo` is invoked
- **THEN** `us-east-1` is used for all AWS operations

#### Scenario: Region not resolvable
- **WHEN** `bastion down --name foo` is invoked without `--region` and without `BASTION_REGION` or a config `region` key
- **THEN** the command exits with a non-zero code and a usage error before making any AWS API calls

