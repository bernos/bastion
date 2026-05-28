## Why

Operators deploying bastions across multiple environments need to label resources for cost allocation, environment segregation, and ownership tracking. Today the stack only applies `Name` and `Owner` tags, giving users no way to add their own tags without modifying the CloudFormation template.

## What Changes

- Add `Tags` field to `Config` struct (`map[string]string`), mapped to config key `tags`
- Register a `--tags` CLI flag on the `up` command using the existing `TagMap` cobra flag type (`cmd/bastion/flags/tagmap.go`), accepting `key=value,key2=value2` format
- Support `BASTION_TAGS` env var using the same comma-separated format
- Flow tags through the full call stack: `Config` → `UpInput` → `DeployBastionInput` → CloudFormation `Tags` parameter on the stack
- Apply user-supplied tags to all tagged resources in `stack.yaml` (in addition to existing `Name` and `Owner` tags)

## Capabilities

### New Capabilities

- `custom-resource-tags`: Config param, CLI flag, and env var for supplying arbitrary key-value tags; tags flow from user input through to the CloudFormation stack deployment call.

### Modified Capabilities

- `bastion-cfn-stack`: The stack now accepts and forwards user-supplied tags to all tagged EC2 resources, extending the existing `Name`/`Owner` tagging requirement.

## Impact

- `internal/config/config.go` — new `Tags` field; default handling (nil map → empty, no default tags injected)
- `cmd/bastion/up.go` — new `--tags` flag using `flags.TagMap`; bind to viper; pass into `UpInput`
- `internal/commands/up.go` — pass `Tags` from `UpInput` to `DeployBastionInput`
- `internal/bastion/bastion.go` — pass tags to CloudFormation `Deploy` call
- `internal/bastion/stack.yaml` — new `Tags` parameter (type `CommaDelimitedList` or CloudFormation tag list); apply to instance and security group resources
- `pkg/aws/cloudformationservice/` — may need to forward tags as stack-level tags on the `CreateChangeSet` / `CreateStack` calls, depending on implementation approach
- No breaking changes; `Tags` is optional and defaults to empty
