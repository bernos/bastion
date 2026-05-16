## Why

The existing `internal/bastion/stack.yaml` CloudFormation template is incomplete — the EC2 instance is commented out, the security group references a missing parameter (`MyIP`), and no parameters are forwarded from `DeployBastionInput` to the stack. Without a working template the CLI cannot actually provision a bastion host.

## What Changes

- Replace the stub `stack.yaml` with a complete CloudFormation template that provisions a bastion host with no public IP address, accessed exclusively via SSM Session Manager and EC2 Instance Connect Endpoint
- Add IAM instance role and instance profile with `AmazonSSMManagedInstanceCore` for SSM Session Manager and EC2 Instance Connect permissions
- Add a security group with no inbound rules (all access is outbound / service-side via SSM and EICE)
- Add EC2 instance in a private subnet with no public IP, wired to the supplied subnet, VPC, AMI, and instance type
- Wire `DeployBastionInput` fields (SubnetID, VPCID, InstanceType, AMIParameterName, Owner) to CloudFormation stack parameters in `bastionService.DeployBastion`
- Extend `Config` and the `up` command flags to capture required inputs (SubnetID, VPCID)

## Capabilities

### New Capabilities

- `bastion-cfn-stack`: CloudFormation template and parameter wiring that provisions an EC2 bastion host with no public IP, IAM instance profile (SSM + EC2 Instance Connect), and no inbound security group rules

### Modified Capabilities

<!-- No existing specs to modify -->

## Impact

- `internal/bastion/stack.yaml` — full rewrite
- `internal/bastion/bastion.go` — pass CloudFormation parameters in `Deploy` call
- `internal/config/config.go` — add SubnetID, VPCID fields
- `cmd/bastion/up.go` — expose new config fields as CLI flags
- No external API or schema changes
