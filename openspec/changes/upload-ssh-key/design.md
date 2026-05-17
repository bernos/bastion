## Context

The `bastion up` command deploys a CloudFormation stack and returns the instance ID and AZ. Operators then connect via EC2 Instance Connect (EICE) or SSM. `SendSSHPublicKey` uploads a temporary public key to the instance (60-second lifetime), enabling password-less SSH over EICE without pre-baking keys. This step is currently manual.

The `cloudformationservice` wrapper under `pkg/aws/` earns its keep by orchestrating a multi-step protocol (create changeset → execute → poll for completion). `BastionService` in `internal/bastion/` holds that service and adds its own business logic on top.

## Goals / Non-Goals

**Goals:**
- Add optional `--public-key-path` flag (+ env var + config YAML) to `up`
- After a successful deploy, if the flag is set: wait for the SSM agent to come online, then call `SendSSHPublicKey`
- Both the SSM wait and key upload are gated on `public-key-path`; omitting it leaves the current flow unchanged

**Non-Goals:**
- Making `--os-user` configurable (default to `ec2-user` for Amazon Linux 2; can be added later)
- Persisting or rotating keys (AWS handles the 60-second expiry)

## Decisions

### Where to call SendSSHPublicKey

**Decision:** Extend `BastionService.DeployBastion` — add `PublicKeyContent` (raw bytes read by the `up` command) to `DeployBastionInput`. If non-empty, the service calls the EC2IC wrapper after a successful deploy.

**Rationale:** The service already owns the full deploy lifecycle. Keeping the AWS call inside `BastionService` means the `up` command stays a thin orchestrator. The cmd reads the file from disk (its responsibility: CLI I/O) and passes the content to the service (AWS calls: service responsibility).

**Alternative considered:** A separate `BastionService.UploadSSHKey` method called from the cmd. Rejected because it leaks orchestration into the cmd layer and forces the cmd to know about instance IDs explicitly.

### Public key validation

**Decision:** Two checks in sequence, both in the `up` command before the deploy starts:
1. Scan the raw file content for the substring `PRIVATE KEY` — this matches all common PEM private key headers (`BEGIN RSA PRIVATE KEY`, `BEGIN OPENSSH PRIVATE KEY`, `BEGIN EC PRIVATE KEY`, `BEGIN PRIVATE KEY`). Return a specific error directing the operator to use the `.pub` file.
2. Call `golang.org/x/crypto/ssh.ParseAuthorizedKey` on the content. This parses the standard authorized-keys line format and rejects anything that isn't a valid public key.

**Rationale:** Private key detection is the most important check — the error message can be specific and actionable ("looks like a private key; did you mean `.pub`?"). `ParseAuthorizedKey` then catches all other malformed input without us needing to enumerate key types or validate base64 manually. Both `golang.org/x/crypto` and `ssm` are already indirect deps; no new modules needed.

### SSM readiness check before key upload

**Decision:** After deploy and before `SendSSHPublicKey`, call `ssm:DescribeInstanceInformation` with an `InstanceIds` filter and poll until `PingStatus == Online`. Poll every 10 seconds with a total timeout of 2 minutes. Only run this check when `PublicKeyContent` is non-empty.

**Rationale:** CloudFormation marks the stack `CREATE_COMPLETE` once the EC2 instance enters the `running` state, but the SSM agent registers 30–60 seconds later. Calling `SendSSHPublicKey` before the agent is online doesn't fail, but the key expires before the instance is connectable — making the upload useless. Gating the key upload on SSM readiness ensures the 60-second key lifetime starts when the instance is actually reachable.

**Alternative considered:** Always waiting for SSM regardless of key upload. Rejected because it adds latency to plain `bastion up` runs with no benefit.

### No service wrapper for EC2 Instance Connect or SSM

**Decision:** Use `ec2instanceconnect.Client` and `ssm.Client` directly — no wrapper packages. Define two minimal unexported interfaces in `internal/bastion/` (one per client, one method each) so both dependencies can be mocked in unit tests.

**Rationale:** Neither client needs orchestration logic — EC2IC is one call, SSM is one call inside a polling loop. A pass-through wrapper package would add a package, interface, mock, and tests for code that does nothing but forward arguments. Local interfaces give the same testability at a fraction of the surface area.

### OS user

**Decision:** Hardcode `ec2-user` as the SSH OS user for now.

**Rationale:** All bastions use Amazon Linux 2 (see stack.yaml AMI param). Exposing this as config adds surface area for no current benefit.

### Config key name

**Decision:** `public-key-path` (kebab-case), mapping to `BASTION_PUBLIC_KEY_PATH` env var and `PublicKeyPath` Go struct field.

**Rationale:** Consistent with existing `subnet-id` / `vpc-id` naming. Viper's key replacer (`-` → `_`) handles env var mapping automatically.

## Risks / Trade-offs

- **60-second key lifetime** → Operators must SSH immediately after `bastion up`. This matches the EC2 Instance Connect design intent; document it in help text.
- **File not found / invalid key** → Surface as a clear error before the deploy, so the operator can fix the path without waiting for CloudFormation. Validate path and read the file at cmd start, not after deploy.
- **SSM timeout** → If the agent never comes online (e.g. missing SSM endpoint, IAM misconfiguration), the command will block for 2 minutes then return an error. The error message should direct the operator to check SSM Fleet Manager.
- **SDK dependency** → `ec2instanceconnect` needs `go get`; `ssm` is already an indirect dep and just needs promoting to direct use.

## Migration Plan

Purely additive — no existing behaviour changes. Deploy as a normal feature branch PR.
