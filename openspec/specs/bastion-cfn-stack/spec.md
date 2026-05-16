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
The CloudFormation stack SHALL attach an IAM role that grants the `ec2-instance-connect:SendSSHPublicKey` action so that EC2 Instance Connect Endpoint can be used for shell access.

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

### Requirement: Stack resources are tagged with Name and Owner
The EC2 instance and security group SHALL be tagged with `Name` (set to `BastionName`) and `Owner` (set to the `Owner` parameter).

#### Scenario: Tags applied on stack creation
- **WHEN** the CloudFormation stack is deployed with a BastionName and Owner
- **THEN** the EC2 instance and security group have matching `Name` and `Owner` tags
