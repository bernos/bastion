## 1. CloudFormation Template

- [ ] 1.1 Add `AvailabilityZone` output to `internal/bastion/stack.yaml` using `!GetAtt BastionInstance.AvailabilityZone`

## 2. CloudFormation Service Layer

- [ ] 2.1 Add `Outputs []types.Output` field to `cloudformationservice.DeployOutput`
- [ ] 2.2 Capture `DescribeStacks` result in `executeChangeSetAndWait` return and populate `DeployOutput.Outputs` in the happy path of `Deploy`
- [ ] 2.3 Call `DescribeStacks` and populate `DeployOutput.Outputs` in the empty-changeset path of `Deploy`

## 3. Bastion Service Layer

- [ ] 3.1 Add `InstanceID` and `AvailabilityZone` string fields to `DeployBastionOutput`
- [ ] 3.2 After calling `cloudFormationService.Deploy`, iterate `DeployOutput.Outputs` to populate `DeployBastionOutput.InstanceID` and `DeployBastionOutput.AvailabilityZone` by matching on `OutputKey`

## 4. CLI

- [ ] 4.1 Replace `fmt.Printf` in `up.go` with `json.NewEncoder(os.Stdout).Encode` on an anonymous struct with `stackName`, `instanceId`, and `availabilityZone` fields

## 5. Tests

- [ ] 5.1 Update `cloudformationservice` unit tests to assert `DeployOutput.Outputs` is populated from `DescribeStacks` after a successful deploy
- [ ] 5.2 Update `bastion` unit tests to assert `DeployBastionOutput.InstanceID` and `AvailabilityZone` are set from mock stack outputs
- [ ] 5.3 Update or add integration test to verify the full flow returns instance ID and AZ
