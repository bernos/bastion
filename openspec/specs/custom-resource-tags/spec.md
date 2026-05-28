### Requirement: Tags can be specified in the config file as a YAML map
The `Config` struct SHALL include a `Tags` field of type `map[string]string` with mapstructure key `tags`. When a config file is present and contains a `tags` key whose value is a YAML mapping, those key-value pairs SHALL be decoded into `Config.Tags`.

#### Scenario: Tags loaded from config file
- **WHEN** the config file contains `tags:\n  env: prod\n  team: platform`
- **THEN** `Config.Tags` equals `{"env": "prod", "team": "platform"}`

#### Scenario: Absent tags key results in empty map
- **WHEN** the config file contains no `tags` key
- **THEN** `Config.Tags` is nil or empty and no error is returned

### Requirement: Tags can be specified via the --tags CLI flag
The `up` command SHALL register a `--tags` flag using the `flags.TagMap` pflag type, accepting a comma-separated list of `key=value` pairs. The flag value SHALL be decoded into `Config.Tags` and take precedence over the config file.

#### Scenario: Tags set via flag
- **WHEN** `bastion up --tags env=staging,team=ops ...` is run
- **THEN** `Config.Tags` equals `{"env": "staging", "team": "ops"}`

#### Scenario: Invalid flag format returns error
- **WHEN** `bastion up --tags notavalidtag ...` is run
- **THEN** cobra returns a usage error before execution begins

### Requirement: Tags can be specified via the BASTION_TAGS environment variable
The `Initialize` config function SHALL honour the `BASTION_TAGS` environment variable using the same `key=value,...` format as the CLI flag. The env var value SHALL be parsed into `Config.Tags` and take precedence over the config file but not over the CLI flag.

#### Scenario: Tags loaded from env var
- **WHEN** `BASTION_TAGS=env=prod,owner=sre` is set and no `--tags` flag is provided
- **THEN** `Config.Tags` equals `{"env": "prod", "owner": "sre"}`

#### Scenario: CLI flag takes precedence over env var
- **WHEN** `BASTION_TAGS=env=prod` is set and `--tags env=staging` is also provided
- **THEN** `Config.Tags` equals `{"env": "staging"}`

### Requirement: Tags flow from Config through to CloudFormation stack deployment
`Config.Tags` SHALL be propagated through `UpInput.Tags`, `DeployBastionInput.Tags`, and finally converted to `[]types.Tag` and set on `cloudformationservice.DeployInput.Tags` before the stack deploy call. An empty or nil `Tags` map SHALL result in no stack-level tags being added.

#### Scenario: Tags passed to CloudFormation deploy
- **WHEN** `DeployBastionInput.Tags` is `{"env": "prod"}`
- **THEN** the CloudFormation `CreateChangeSet` call includes a tag with `Key="env"` and `Value="prod"`

#### Scenario: Nil tags produce no stack-level tags
- **WHEN** `DeployBastionInput.Tags` is nil or empty
- **THEN** the CloudFormation `CreateChangeSet` call receives an empty `Tags` slice
