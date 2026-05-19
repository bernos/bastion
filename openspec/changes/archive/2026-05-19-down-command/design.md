## Context

`bastion up` deploys a CloudFormation stack named `<name>-stack`. There is no corresponding `down` command; users must tear down bastions manually. The `down.go` file in `cmd/bastion/` is a stub. `BastionService` and `CloudFormationService` have no delete operations.

## Goals / Non-Goals

**Goals:**
- Implement `bastion down --name <name>` to delete the bastion CloudFormation stack and wait for completion
- Keep the implementation consistent with `up`: same name convention, same config resolution order, JSON stdout on success, errors to stderr via Cobra

**Non-Goals:**
- Interactive confirmation prompt (callers can add this in scripts if needed)
- Deleting resources created outside the CloudFormation stack (e.g. SSH keys, logs)
- Listing or discovering stacks to pick from

## Decisions

### Stack name derivation
Stack name is derived as `<name>-stack`, identical to `DeployBastion`. No persistent state or stack lookup is needed — the convention is the contract.

*Alternative considered*: Accept a raw stack name flag. Rejected — it leaks the internal naming convention and is inconsistent with `up`.

### Config resolution
`--name` is bound to Viper via `cmd.Flags()`, exactly as in `up`. The root command's `PersistentPreRunE` calls `config.Initialize` for every subcommand, so `--name` / `BASTION_NAME` env var / `name:` in config file all resolve into `cfg.Name` automatically — no extra wiring needed in `down`.

### Behaviour when stack does not exist
Return an error: `stack "<name>-stack" does not exist`. A missing stack most likely indicates a typo in `--name`; silently succeeding would hide that mistake.

*Alternative considered*: Treat not-found as success (idempotent). Rejected — too easy to mask errors in automation.

### Where deletion logic lives
- `CloudFormationService` gains a `DeleteStack(ctx, stackName)` method — calls `cloudformation.DeleteStack` then blocks on `StackDeleteCompleteWaiter`.
- `BastionService` gains a `DeleteBastion(ctx, *DeleteBastionInput)` method — derives the stack name and delegates to `CloudFormationService.DeleteStack`.
- `CloudFormationClient` interface gains `DeleteStack` to keep the mock boundary consistent.

This mirrors the layering already established by `Deploy` / `DeployBastion`.

### Output format
Print a JSON object to stdout on success: `{"stackName": "..."}`. Consistent with `up`, composable in scripts.

## Risks / Trade-offs

- **Slow deletion**: CloudFormation stack deletion (EC2 instance, IAM role, instance profile, security group) typically takes 2–4 minutes. The command blocks — this is intentional and consistent with `up`.
- **DELETE_FAILED state**: If deletion fails mid-way (e.g. resource held by a dependency), CloudFormation leaves the stack in `DELETE_FAILED`. The waiter surfaces this as an error; the user must resolve it manually in the console.
