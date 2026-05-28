## 1. Config Layer

- [ ] 1.1 Add `Tags map[string]string` field to `Config` struct in `internal/config/config.go` with mapstructure tag `tags`
- [ ] 1.2 Add a mapstructure decode hook in `Initialize` that converts a string in `key=value,...` format to `map[string]string`, composing it with the default hooks via `mapstructure.ComposeDecodeHookFunc`; treat empty string as a no-op
- [ ] 1.3 Add unit tests in `internal/config/config_test.go` covering: tags from YAML map, tags from env var string, tags from CLI flag, CLI flag precedence over env var, absent tags produces nil/empty map

## 2. CLI Flag

- [ ] 2.1 Register `--tags` flag on the `up` command in `cmd/bastion/up.go` using `flags.TagMap` as the pflag value type
- [ ] 2.2 Pass `cfg.Tags` into `UpInput.Tags` in the `RunE` closure in `cmd/bastion/up.go`

## 3. Command and Service Wiring

- [ ] 3.1 Add `Tags map[string]string` field to `UpInput` in `internal/commands/up.go`
- [ ] 3.2 Pass `input.Tags` into `DeployBastionInput.Tags` in `upCommand.Run`
- [ ] 3.3 Add `Tags map[string]string` field to `DeployBastionInput` in `internal/bastion/bastion.go`
- [ ] 3.4 In `bastionService.DeployBastion`, convert `input.Tags` (`map[string]string`) to `[]types.Tag` and set on `cloudformationservice.DeployInput.Tags`
- [ ] 3.5 Add unit test in `internal/commands/up_test.go` asserting `Tags` passes through from `UpInput` to `DeployBastionInput`
- [ ] 3.6 Add unit test in `internal/bastion/bastion_test.go` asserting user-supplied tags appear as stack-level tags in the captured `DeployInput`; assert nil tags produces empty `Tags` slice
