## 1. Config Layer

- [x] 1.1 Add `Tags map[string]string` field to `Config` struct in `internal/config/config.go` with mapstructure tag `tags`
- [x] 1.2 Add a mapstructure decode hook in `Initialize` that converts a string in `key=value,...` format to `map[string]string`, composing it with the default hooks via `mapstructure.ComposeDecodeHookFunc`; treat empty string as a no-op
- [x] 1.3 Add unit tests in `internal/config/config_test.go` covering: tags from YAML map, tags from env var string, tags from CLI flag, CLI flag precedence over env var, absent tags produces nil/empty map

## 2. CLI Flag

- [x] 2.1 Register `--tags` flag on the `up` command in `cmd/bastion/up.go` using `flags.TagMap` as the pflag value type
- [x] 2.2 Pass `cfg.Tags` into `UpInput.Tags` in the `RunE` closure in `cmd/bastion/up.go`

## 3. Command and Service Wiring

- [x] 3.1 Add `Tags map[string]string` field to `UpInput` in `internal/commands/up.go`
- [x] 3.2 Pass `input.Tags` into `DeployBastionInput.Tags` in `upCommand.Run`
- [x] 3.3 Add `Tags map[string]string` field to `DeployBastionInput` in `internal/bastion/bastion.go`
- [x] 3.4 In `bastionService.DeployBastion`, convert `input.Tags` (`map[string]string`) to `[]types.Tag` and set on `cloudformationservice.DeployInput.Tags`
- [x] 3.5 Add unit test in `internal/commands/up_test.go` asserting `Tags` passes through from `UpInput` to `DeployBastionInput`
- [x] 3.6 Add unit test in `internal/bastion/bastion_test.go` asserting user-supplied tags appear as stack-level tags in the captured `DeployInput`; assert nil tags produces empty `Tags` slice
