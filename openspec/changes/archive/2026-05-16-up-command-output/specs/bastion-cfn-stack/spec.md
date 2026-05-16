## ADDED Requirements

### Requirement: Stack outputs expose instance ID and availability zone
The CloudFormation stack SHALL declare `InstanceId` and `AvailabilityZone` as Outputs, set to `!Ref BastionInstance` and `!GetAtt BastionInstance.AvailabilityZone` respectively.

#### Scenario: Stack outputs are present after deploy
- **WHEN** the bastion CloudFormation stack reaches `CREATE_COMPLETE` or `UPDATE_COMPLETE`
- **THEN** `DescribeStacks` returns an `Outputs` list that includes entries with `OutputKey` equal to `InstanceId` and `AvailabilityZone`, each with a non-empty `OutputValue`

### Requirement: DeployBastionOutput includes instance ID and availability zone
`DeployBastionOutput` SHALL expose `InstanceID` and `AvailabilityZone` string fields populated from the corresponding CloudFormation stack outputs after a successful deploy.

#### Scenario: Fields populated after successful deploy
- **WHEN** `BastionService.DeployBastion` returns without error
- **THEN** `DeployBastionOutput.InstanceID` is a non-empty EC2 instance ID string (e.g. `i-0abc1234`) and `DeployBastionOutput.AvailabilityZone` is a non-empty AZ string (e.g. `ap-southeast-2a`)

#### Scenario: Fields empty when stack lacks outputs
- **WHEN** `BastionService.DeployBastion` returns without error but the stack has no outputs (e.g. an old stack that predates this change)
- **THEN** `DeployBastionOutput.InstanceID` and `DeployBastionOutput.AvailabilityZone` are empty strings, and no error is returned

### Requirement: `up` command prints JSON output
The `up` command SHALL write a JSON object to stdout on success containing `stackName`, `instanceId`, and `availabilityZone` fields.

#### Scenario: JSON printed after successful deploy
- **WHEN** `bastion up` completes successfully
- **THEN** stdout contains a single JSON object with string fields `stackName`, `instanceId`, and `availabilityZone`
