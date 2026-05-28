### Requirement: Bastion instance has no public IP address
The CloudFormation stack SHALL launch the EC2 instance into a private subnet with `AssociatePublicIpAddress` set to `false`. No Elastic IP SHALL be allocated or associated.

#### Scenario: Instance launched without public IP
- **WHEN** the `bastion up` command is run with a valid private SubnetId and VpcId
- **THEN** the resulting EC2 instance has no public IP address and no Elastic IP associated

### Requirement: Bastion instance is accessible via SSM Session Manager
The CloudFormation stack SHALL attach an IAM instance profile that includes the `AmazonSSMManagedInstanceCore` managed policy, enabling SSM Session Manager access without inbound security group rules.

#### Scenario: SSM agent can register with the service
- **WHEN** the bastion EC2 instance is running in a subnet with outbound internet access or SSM VPC endpoints
- **THEN** the instance appears as a managed instance in SSM Fleet Manager and sessions can be started via `aws ssm start-session`

### Requirement: Bastion instance is accessible via EC2 Instance Connect
The CloudFormation stack SHALL provision the instance so that it is reachable via an EC2 Instance Connect Endpoint (EICE). No instance-side IAM permissions are required for EICE — `ec2-instance-connect:SendSSHPublicKey` and `ec2-instance-connect:OpenTunnel` are caller permissions, not instance permissions. The instance role SHALL NOT include these actions.

#### Scenario: EC2 Instance Connect Endpoint shell access
- **WHEN** an EC2 Instance Connect Endpoint exists in the target VPC and the caller has `ec2-instance-connect:OpenTunnel` permission
- **THEN** the operator can open a shell to the bastion using the AWS CLI without any inbound security group rules

### Requirement: Security group has no inbound rules
The CloudFormation stack SHALL create a security group with no ingress rules. All inbound access is brokered by SSM and EC2 Instance Connect Endpoint.

#### Scenario: Security group created with empty ingress
- **WHEN** the CloudFormation stack is deployed
- **THEN** the bastion security group has zero ingress rules and unrestricted egress

### Requirement: CloudFormation stack accepts runtime parameters
The CloudFormation stack SHALL accept `SubnetId`, `VpcId`, `InstanceType`, `AmiId`, `BastionName`, and `Owner` as parameters. `InstanceType` SHALL default to `t3.micro`. `AmiId` SHALL default to the latest Amazon Linux 2 SSM path.

#### Scenario: Stack deployed with required parameters
- **WHEN** `bastionService.DeployBastion` is called with a populated `DeployBastionInput`
- **THEN** the CloudFormation stack receives all six parameters and the instance is placed in the correct subnet and VPC

#### Scenario: Stack deployed with defaults for optional parameters
- **WHEN** `DeployBastionInput.InstanceType` is empty
- **THEN** the stack uses `t3.micro` as the instance type

### Requirement: Stack resources are tagged with Name and Owner plus any user-supplied tags
The EC2 instance and security group SHALL be tagged with `Name` (set to `BastionName`) and `Owner` (set to the `Owner` parameter) via resource-level tags in the CloudFormation template. In addition, any user-supplied tags SHALL be applied as CloudFormation stack-level tags on the `CreateChangeSet` call; CloudFormation SHALL propagate these stack-level tags to all supported resources in the stack.

#### Scenario: Built-in tags applied on stack creation
- **WHEN** the CloudFormation stack is deployed with a BastionName and Owner
- **THEN** the EC2 instance and security group have matching `Name` and `Owner` tags

#### Scenario: User-supplied tags applied as stack-level tags
- **WHEN** `DeployBastionInput.Tags` contains `{"env": "prod", "team": "platform"}`
- **THEN** the CloudFormation `CreateChangeSet` call includes stack-level tags with those key-value pairs and the EC2 instance receives both the template tags and the stack-level tags

#### Scenario: No user tags — stack-level tags list is empty
- **WHEN** `DeployBastionInput.Tags` is nil or empty
- **THEN** the CloudFormation `CreateChangeSet` call receives an empty stack-level `Tags` slice and only the template-defined `Name` and `Owner` tags are present on resources

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
