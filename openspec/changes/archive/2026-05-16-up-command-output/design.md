## Context

The `up` command currently prints only the CloudFormation stack name on success. Operators need the EC2 instance ID and availability zone immediately after deploy to connect via SSM or EC2 Instance Connect Endpoint.

`stack.yaml` already has an `Outputs` section with `InstanceId`. `AvailabilityZone` is missing. `cloudformationservice.DeployOutput` does not expose stack outputs — it only returns `CreateChangeSetInput` and `Changes`. `DeployBastionOutput` has only `StackName`.

The change touches three layers: the CFN template, the service layer (`cloudformationservice` + `bastion`), and the CLI (`up.go`).

## Goals / Non-Goals

**Goals:**
- Expose `InstanceId` and `AvailabilityZone` from the CloudFormation stack as `Outputs`
- Surface those values on `DeployBastionOutput`
- Print them as JSON from `up.go`

**Non-Goals:**
- Exposing other stack outputs (security group ID, IAM role ARN, etc.)
- Adding a dedicated `status` or `describe` command
- Changing the `DeployOutput` contract in `cloudformationservice` beyond what's needed here

## Decisions

### Fetch stack outputs inside `cloudformationservice.Deploy`

`Deploy` already calls `executeChangeSetAndWait`, which internally calls `DescribeStacks` and returns the result. The return value is currently discarded. We can capture it and populate `DeployOutput.Outputs []types.Output`.

For the empty-changeset path (no changes, changeset deleted), the stack is already in its final state, so a single `DescribeStacks` call at that point also works.

**Alternative considered**: Add a `GetStackOutputs(ctx, stackName)` method to `CloudFormationService` and call it from `bastion.go`. Rejected: it adds a second round-trip and leaks stack-describe logic into the bastion layer.

### `bastion.go` reads outputs by key

`DeployBastionOutput` gains two string fields: `InstanceID` and `AvailabilityZone`. `DeployBastion` iterates `DeployOutput.Outputs` and matches on `OutputKey == "InstanceId"` and `OutputKey == "AvailabilityZone"`. Missing keys are left as empty string (not an error — old stacks may not have them).

### `up.go` uses `encoding/json` for output

Replace `fmt.Printf("Bastion deployed. Stack: %s\n", ...)` with a `json.NewEncoder(os.Stdout).Encode(...)` call on an anonymous struct with `stackName`, `instanceId`, and `availabilityZone` fields (camelCase JSON keys). This is machine-parseable and extensible.

## Risks / Trade-offs

- **Empty changeset path**: When Deploy detects no changes and deletes the changeset, it now does an extra `DescribeStacks` call. This is one extra API call per no-op deploy, which is acceptable.
- **Missing outputs on old stacks**: Stacks deployed before this change lack `AvailabilityZone`. The empty-string fallback in `bastion.go` handles this gracefully; the JSON output will show `""`.
- **`DeployOutput` is a public type**: Adding `Outputs []types.Output` is additive and backward-compatible. Callers that don't read the new field are unaffected.

## Migration Plan

No migration required. The CloudFormation template change adds an `Output` — this is backward-compatible; CloudFormation does not require consumers to read outputs.

Existing stacks are updated on the next `bastion up` run (change-set strategy picks up the template diff).
