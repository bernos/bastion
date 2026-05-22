# explicit-region Specification

## Purpose
TBD - created by archiving change add-ssh-proxy-commands. Update Purpose after archive.
## Requirements
### Requirement: All commands accept `--region` resolved from flag, env var, or config
Every `bastion` subcommand (`up`, `down`, `ssh`, `proxy`) SHALL accept a `--region` flag. Region SHALL also be resolvable from the `BASTION_REGION` environment variable or the `region` key in a config file (`./config.yaml` or `~/.config/bastion/config.yaml`), following the same resolution order as all other config values (flag → env var → config file).

#### Scenario: Region resolved from CLI flag
- **WHEN** any command is invoked with `--region ap-southeast-2`
- **THEN** `ap-southeast-2` is used as the region for all AWS operations in that invocation

#### Scenario: Region resolved from environment variable
- **WHEN** `BASTION_REGION=us-east-1` is set and a command is invoked without `--region`
- **THEN** `us-east-1` is used as the region for all AWS operations in that invocation

#### Scenario: Region resolved from config file
- **WHEN** a config file contains `region: eu-west-1` and a command is invoked without `--region` and without `BASTION_REGION`
- **THEN** `eu-west-1` is used as the region for all AWS operations in that invocation

### Requirement: `--region` is a required input on all commands
Every `bastion` subcommand SHALL require that region is resolvable. If region cannot be resolved from any source (flag, env var, config file), the command SHALL exit with a non-zero code and a usage error before making any AWS API calls.

#### Scenario: Region not set on any command
- **WHEN** any command is invoked without `--region`, without `BASTION_REGION`, and without a config file containing `region`
- **THEN** the command exits with a non-zero code and prints a usage error identifying the missing region

### Requirement: Resolved region is passed to AWS SDK client construction
The resolved region SHALL be passed to `awsconfig.LoadDefaultConfig` via `awsconfig.WithRegion` in `dependencies.AwsConfig`. This ensures all AWS SDK clients created from that config use the explicitly supplied region, with no fallback to the ambient SDK default chain.

#### Scenario: SDK clients use the supplied region
- **WHEN** `--region ap-southeast-2` is supplied
- **THEN** all CloudFormation, SSM, and EC2 Instance Connect API calls are directed to `ap-southeast-2`

### Requirement: Service method input structs carry a `Region` field
`DeployBastionInput`, `DeleteBastionInput`, `DescribeBastionInput`, and `WaitForSSMReadyInput` SHALL each include a `Region string` field. Callers SHALL populate this field with the resolved region.

#### Scenario: Region present in DeployBastionInput
- **WHEN** `BastionService.DeployBastion` is called
- **THEN** the input struct contains a non-empty `Region` field matching the region supplied to the command

#### Scenario: Region present in DeleteBastionInput
- **WHEN** `BastionService.DeleteBastion` is called
- **THEN** the input struct contains a non-empty `Region` field

#### Scenario: Region present in DescribeBastionInput
- **WHEN** `BastionService.DescribeBastion` is called
- **THEN** the input struct contains a non-empty `Region` field

#### Scenario: Region present in WaitForSSMReadyInput
- **WHEN** `BastionService.WaitForSSMReady` is called
- **THEN** the input struct contains a non-empty `Region` field

