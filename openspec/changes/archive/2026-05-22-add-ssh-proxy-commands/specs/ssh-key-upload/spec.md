## REMOVED Requirements

### Requirement: `up` command accepts an optional public key path
**Reason**: Key upload is being removed from `up`. The `up` command is now a pure CloudFormation deploy command. Ephemeral key generation and upload happen atomically in `bastion ssh` and `bastion proxy` at connection time.
**Migration**: Remove `--public-key-path` from any `bastion up` invocations. Use `bastion ssh` or `bastion proxy` to connect — these commands handle key upload automatically.

### Requirement: Public key file is validated before deploy
**Reason**: Removed along with `--public-key-path`. Key validation is no longer needed in `up`.
**Migration**: No action required. `bastion ssh` and `bastion proxy` generate ephemeral keys internally; no user-supplied key file is needed.

### Requirement: SSM agent readiness is confirmed before key upload
**Reason**: SSM readiness is no longer checked during `up`. It is now checked by `bastion ssh` and `bastion proxy` via `BastionService.WaitForSSMReady` before the connection is established.
**Migration**: No action required. The wait still happens — it has moved to the connection commands.

### Requirement: SSH public key is uploaded after SSM agent is confirmed online
**Reason**: Key upload is no longer part of the deploy flow. `bastion ssh` and `bastion proxy` generate an ephemeral key and upload it immediately before opening the connection.
**Migration**: No action required. Key upload still happens — it has moved to the connection commands.
