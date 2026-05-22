## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: `BastionService` exposes a `DeleteBastion` operation
`BastionService` SHALL expose a `DeleteBastion(ctx context.Context, input *DeleteBastionInput) error` method. `DeleteBastionInput` SHALL contain `BastionName string` and `Region string` fields. The method SHALL derive the stack name as `<BastionName>-stack`, verify it exists, and delegate deletion to `CloudFormationService.DeleteStack`.

#### Scenario: Bastion deleted successfully
- **WHEN** `BastionService.DeleteBastion` is called with a `BastionName` whose corresponding stack exists and a valid `Region`
- **THEN** `CloudFormationService.DeleteStack` is called with `<BastionName>-stack` and the method returns nil on success

#### Scenario: Stack does not exist
- **WHEN** `BastionService.DeleteBastion` is called with a `BastionName` whose stack does not exist
- **THEN** the method returns a descriptive error without calling `CloudFormationService.DeleteStack`
