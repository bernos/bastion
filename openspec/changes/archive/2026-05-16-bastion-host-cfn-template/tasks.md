## 1. CloudFormation Template

- [x] 1.1 Rewrite `internal/bastion/stack.yaml` — add `BastionRole` (IAM role with EC2 trust + `AmazonSSMManagedInstanceCore`), `BastionInstanceProfile`, `BastionSecurityGroup` (no ingress), and `BastionInstance` (private subnet, no public IP, instance profile attached)
- [x] 1.2 Add CFN parameters: `SubnetId`, `VpcId`, `InstanceType` (default `t3.micro`), `AmiId` (default AL2 SSM path), `BastionName`, `Owner`
- [x] 1.3 Add `Name` and `Owner` tags to the EC2 instance and security group

## 2. Config & CLI

- [x] 2.1 Add `SubnetID` and `VPCID` string fields to `internal/config/config.go` `Config` struct
- [x] 2.2 Bind `SubnetID` and `VPCID` as required CLI flags in `cmd/bastion/up.go`

## 3. Parameter Wiring

- [x] 3.1 Update `bastionService.DeployBastion` in `internal/bastion/bastion.go` to pass all six CFN parameters from `DeployBastionInput` to `cloudformationservice.Deploy`
- [x] 3.2 Verify `cloudformationservice.DeployInput` supports a `Parameters` field (add it if missing)

## 4. Tests

- [x] 4.1 Update or add unit tests in `internal/bastion/` to assert the correct parameters are forwarded to the CloudFormation service mock
- [x] 4.2 Update the integration test in `internal/bastion/` to pass `SubnetID` and `VPCID` (can use a LocalStack-created VPC/subnet or placeholder values)
