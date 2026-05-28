## Context

The CloudFormation service layer (`pkg/aws/cloudformationservice`) already has a `Tags []types.Tag` field on `DeployInput` that is forwarded directly to the CloudFormation `CreateChangeSet` API call as stack-level tags. CloudFormation propagates stack-level tags to all supported resources automatically. No changes to `stack.yaml` are needed.

The main challenge is config layer wiring: Viper's `Unmarshal` handles `map[string]string` cleanly from YAML, but when the value arrives as a string (env var `BASTION_TAGS=key=value,key2=value2` or a custom pflag type), mapstructure cannot convert it to `map[string]string` without help.

## Goals / Non-Goals

**Goals:**
- Accept user-supplied tags via config file (YAML map), `--tags` CLI flag, and `BASTION_TAGS` env var
- Flow tags to CloudFormation as stack-level tags (propagated to all stack resources)
- No changes to `stack.yaml` or its parameters

**Non-Goals:**
- Applying tags selectively to specific resources within the stack
- Validating tag keys/values against AWS limits (50 tag max, key/value length)
- Merging user tags with built-in `Name`/`Owner` tags (those remain as resource-level tags in the template; stack-level tags are additive)

## Decisions

### Use stack-level tags, not a CloudFormation template parameter

`DeployInput.Tags []types.Tag` is already forwarded to `CreateChangeSet`. Stack-level tags are propagated by CloudFormation to all supported resources and appear alongside the template-defined `Name`/`Owner` tags — no template changes required.

**Alternative considered:** Add a `Tags` parameter to `stack.yaml` as a `CommaDelimitedList` and iterate over it with `Fn::Split` + dynamic resource tag blocks. Rejected: CloudFormation's YAML support for dynamic tag lists is verbose, fragile, and limits the parameter to 4096 characters.

### Decode string → `map[string]string` via mapstructure hook

Register a custom `mapstructure.DecodeHookFuncType` in the `v.Unmarshal` call that detects when the source type is `string` and the target is `map[string]string`, then parses using the same `key=value,key2=value2` logic as `TagMap.Set`. This handles the env var case transparently within Viper's existing flow.

**Alternative considered:** Post-unmarshal manual check — inspect `cfg.Tags`, then re-read the raw env var string and parse it. Rejected: duplicates Viper's precedence logic and requires manually re-implementing source priority (flag > env > file).

### Register `--tags` flag using the `flags.TagMap` pflag type

`TagMap` implements `pflag.Value`. Registering it as a pflag on the `up` command lets Viper pick up the value via `BindPFlags`. The decode hook converts the stored string representation to `map[string]string` during unmarshal.

## Risks / Trade-offs

- **mapstructure hook ordering**: The hook must be composed with the existing default hooks via `mapstructure.ComposeDecodeHookFunc` so it doesn't break other field decoding. → Use `mapstructure.StringToTimeDurationHookFunc()` and `mapstructure.StringToSliceHookFunc(",")` as the base, prepend the custom hook.
- **Empty-string env var**: `BASTION_TAGS=` should result in an empty map, not an error. The hook must treat an empty string as a no-op. → Handle explicitly in the hook.
- **Stack tag limits**: AWS allows a maximum of 50 tags per stack (including those added by AWS itself). No validation is added now; document the limit in `--tags` flag help text.

## Migration Plan

No migration needed. The `Tags` field is optional and defaults to an empty map, leaving existing deployments unchanged. Stack-level tags on an existing stack are additive on the next `UPDATE` changeset.
