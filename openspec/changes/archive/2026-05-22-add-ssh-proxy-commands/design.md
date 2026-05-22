## Context

The bastion CLI uses CloudFormation to deploy a private EC2 instance accessible via SSM Session Manager (AmazonSSMManagedInstanceCore policy) and EC2 Instance Connect. The instance has no public IP and no inbound security group rules. Today the `up` command optionally uploads an SSH public key via EC2 Instance Connect as a convenience, but this couples infrastructure provisioning with connection concerns and forces a 60-second key window to start immediately at deploy time.

SSH access to the instance works via SSM's `AWS-StartSSHSession` document as a ProxyCommand — the SSM agent on the instance proxies TCP port 22 through the SSM WebSocket channel, so no inbound rules or public network path are required.

## Goals / Non-Goals

**Goals:**
- `bastion ssh` opens an interactive SSH session to a named bastion
- `bastion proxy` starts a local SOCKS5 proxy through a named bastion
- Key upload is atomic with connection — the 60-second window opens exactly when the connection starts
- Region is explicit everywhere — no ambient SDK region resolution
- `up` is a pure infrastructure command with no connection concerns

**Non-Goals:**
- Native Go SSH implementation (we exec the system `ssh` binary)
- EC2 Instance Connect Endpoint (EICE) transport — SSM only for now
- SSH config file generation (`~/.ssh/config` ProxyCommand stanza)
- Support for user-supplied key pairs — ephemeral keys only

## Decisions

### Use `exec.Command` (not `syscall.Exec`) for SSH

`syscall.Exec` replaces the current process, so deferred cleanup never runs. The ephemeral private key would remain in `/tmp` until the OS cleans it. With `exec.Command` and `cmd.Stdin/Stdout/Stderr = os.Stdin/Stdout/Stderr`, the `ssh` binary receives the real terminal file descriptors and handles its own PTY allocation and terminal mode setup — behavior is equivalent to `syscall.Exec` in practice, and cleanup is straightforward after `cmd.Wait()`.

### Ephemeral ed25519 key pair per connection

Generating a key pair in memory per connection eliminates all key management concerns — no `--private-key-path` flag, no pre-existing key infrastructure required. The private key is written to a `0600` temp file immediately before exec and removed after `cmd.Wait()`. The public key is valid for 60 seconds; by the time the 60-second window expires, the SSH handshake has already completed and the session is established.

Alternatives considered:
- **User-provided key** (`--private-key-path`): requires users to manage a key pair; no benefit in this ephemeral-access model
- **ssh-agent injection**: avoids the temp file entirely but requires implementing the ssh-agent protocol or shelling out to `ssh-agent`/`ssh-add`; significant complexity for marginal gain

### SSM Session Manager as SSH transport

All session activity is logged in CloudTrail automatically. The instance requires no inbound rules and no public IP. EC2 Instance Connect Endpoint (EICE) was considered but requires a VPC endpoint resource not provisioned by this project, and has weaker audit trail by default.

The ProxyCommand:
```
aws ssm start-session \
  --target %h \
  --document-name AWS-StartSSHSession \
  --parameters portNumber=22 \
  --region <region>
```

This requires `aws` CLI and `session-manager-plugin` on the operator's machine. The commands will verify both are present (via `exec.LookPath`) before proceeding and return a clear error if either is missing.

### Region as explicit required input

Region is read from `--region` (or `BASTION_REGION` / `region` in config) and flows into `dependencies.LoadDefaultConfig` via `awsconfig.WithRegion`. All service method input structs carry a `Region` field. This ensures the ProxyCommand subprocess and the Go SDK calls use the same region from the same source — no risk of the two diverging under a named AWS profile or mixed environment.

### `DescribeBastion` as the canonical name → instance resolver

The `ssh` and `proxy` commands identify the target by `--name`, consistent with `up` and `down`. `BastionService.DescribeBastion` looks up the stack `<name>-stack` via `CloudFormationService.GetStackOutputs` and returns `InstanceID`, `AvailabilityZone`, and `Region`. This keeps commands stateless — no local state file — and avoids requiring users to look up or track instance IDs.

### `WaitForSSMReady` promoted to a public interface method

The SSM readiness poll is needed by `ssh` and `proxy` but not by `up` (which no longer uploads keys). Promoting it to the `BastionService` interface as `WaitForSSMReady(ctx, *WaitForSSMReadyInput) error` keeps the logic in one place and makes it testable via the existing mock pattern. Input struct: `{ InstanceID, Region string }`.

### Connection flow for both `ssh` and `proxy`

```
DescribeBastion(name, region)
  → InstanceID, AZ, Region

WaitForSSMReady(InstanceID, Region)
  → polls until PingStatus == Online (2-minute timeout)

generate ed25519 key pair (in memory)
SendSSHPublicKey(InstanceID, AZ, publicKey)
  → 60-second window opens

write privateKey → os.CreateTemp (mode 0600)
exec.Command("ssh", ...args)
  cmd.Stdin/Stdout/Stderr = os.Stdin/Stdout/Stderr
  cmd.Wait()

os.Remove(tempKeyFile)
```

`ssh` and `proxy` differ only in the ssh flags passed:
- `ssh`: interactive session (no extra flags)
- `proxy`: `-D <port> -N` (SOCKS5, no shell)

For `proxy`, a goroutine forwards `SIGINT`/`SIGTERM` to the child process so the proxy shuts down cleanly on Ctrl-C.

## Risks / Trade-offs

- **`aws` CLI + `session-manager-plugin` required at runtime** → Commands check for both via `exec.LookPath` at startup and return a clear installation hint if missing
- **Temp key file on disk** → mode `0600`, random name in `/tmp`, removed after `cmd.Wait()`; key is expired (60s) before the session typically ends, so exposure window is negligible
- **SSM agent must be online** → `WaitForSSMReady` polls for up to 2 minutes; if the bastion was just deployed, this may add latency on first connect
- **Breaking change to `up` and `down`** → `--region` becomes required; teams using these commands without `--region` will need to update scripts or set `BASTION_REGION`

## Migration Plan

1. Users with `--public-key-path` in scripts or CI: remove the flag; key upload now happens automatically via `bastion ssh` or `bastion proxy`
2. All callers of `bastion up` and `bastion down`: add `--region <region>` (or set `BASTION_REGION` in the environment)
3. No infrastructure changes required — the CloudFormation stack is unchanged
