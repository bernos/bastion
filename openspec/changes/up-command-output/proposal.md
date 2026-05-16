## Why

After `bastion up` succeeds, the operator has no way to know the EC2 instance ID or availability zone without opening the AWS console or running a separate CLI command. These values are needed to connect via SSM or EC2 Instance Connect Endpoint.

## What Changes

- The CloudFormation stack gains two `Outputs`: `InstanceId` and `AvailabilityZone`
- `DeployBastionOutput` gains `InstanceID` and `AvailabilityZone` string fields populated from those stack outputs
- The `up` command prints a JSON object containing `stackName`, `instanceId`, and `availabilityZone` instead of a plain text line

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `bastion-cfn-stack`: Stack must expose `InstanceId` and `AvailabilityZone` as CloudFormation Outputs so the CLI can surface them to the operator

## Impact

- `internal/bastion/bastion_cfn_template.yaml` — add `Outputs` section
- `internal/bastion/bastion.go` — populate new fields on `DeployBastionOutput` by reading stack outputs
- `cmd/bastion/up.go` — replace plain-text `fmt.Printf` with `json.Marshal` output
