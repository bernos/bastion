## Context

The existing `stack.yaml` is a stub with commented-out resources and a broken parameter reference. `bastionService.DeployBastion` calls `cloudformationservice.Deploy` but passes no stack parameters, so the template cannot be parameterised at all. The CLI's `up` command doesn't yet expose subnet/VPC inputs.

Access model: no SSH key pair, no public IP. Operators connect via SSM Session Manager (outbound from the SSM agent to the SSM service) and EC2 Instance Connect Endpoint (EICE, provisioned separately per-VPC). Both require only outbound connectivity from the instance — no inbound security group rules are needed.

## Goals / Non-Goals

**Goals:**
- Complete `stack.yaml` with all resources needed to launch a private bastion instance
- Forward `DeployBastionInput` fields as CloudFormation stack parameters
- Extend `Config` and `up` flags to surface subnet/VPC inputs to the caller

**Non-Goals:**
- Provisioning VPC endpoints for SSM (assumed: NAT gateway or pre-existing endpoints in the target VPC)
- Provisioning the EC2 Instance Connect Endpoint (per-VPC resource managed outside this stack)
- Instance hardening, OS patching automation, or CloudWatch agent setup

## Decisions

### CloudFormation resources

The template includes:

| Resource | Type | Notes |
|---|---|---|
| `BastionRole` | `AWS::IAM::Role` | EC2 trust policy; `AmazonSSMManagedInstanceCore` managed policy for SSM + EICE |
| `BastionInstanceProfile` | `AWS::IAM::InstanceProfile` | Wraps `BastionRole` |
| `BastionSecurityGroup` | `AWS::EC2::SecurityGroup` | No ingress rules; default egress (all outbound) retained for SSM/EICE |
| `BastionInstance` | `AWS::EC2::Instance` | Private subnet, `AssociatePublicIpAddress: false`, IamInstanceProfile attached |

No Elastic IP — the instance is private and unreachable from the internet by design.

### Parameter strategy

CFN parameters map 1-to-1 to `DeployBastionInput` fields. The Go `Deploy` call is extended to pass them as `cloudformationservice.Parameter` entries. This keeps the template self-describing and avoids hardcoding values in the binary.

Parameters:

| CFN Parameter | Source field | Notes |
|---|---|---|
| `SubnetId` | `DeployBastionInput.SubnetID` | Private subnet to place the instance |
| `VpcId` | `DeployBastionInput.VPCID` | VPC for the security group |
| `InstanceType` | `DeployBastionInput.InstanceType` | Default `t3.micro` |
| `AmiId` | `DeployBastionInput.AMIParameterName` | SSM parameter path resolved by CFN |
| `BastionName` | `DeployBastionInput.BastionName` | Used for Name tag |
| `Owner` | `DeployBastionInput.Owner` | Used for Owner tag |

### Config / CLI wiring

`Config` gains `SubnetID` and `VPCID` string fields (required). `up.go` binds them as required flags. `AMIParameterName` and `InstanceType` get defaults (`/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2` and `t3.micro`) so they remain optional.

## Risks / Trade-offs

- **SSM connectivity depends on VPC egress** — if the subnet has no NAT gateway and no SSM VPC endpoints, the SSM agent can't reach the service. → Mitigation: document the prerequisite; adding VPC endpoints is out of scope for this change.
- **EICE is a separate resource** — the template doesn't provision the EC2 Instance Connect Endpoint. → Mitigation: document that EICE must exist in the VPC before `ec2-instance-connect` CLI commands will work.
- **Amazon Linux 2 EOL** — the default AMI SSM path targets AL2. → Mitigation: the parameter is caller-overridable; migration to AL2023 is a future task.

## Open Questions

- Should `Owner` default to the caller's IAM identity (from STS `GetCallerIdentity`) rather than being a required input? Currently modelled as an optional string with no default.
