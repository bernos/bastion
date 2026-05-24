## Context

The connection flow currently lives across two files in `cmd/bastion/`:

- `connect.go` — `generateEphemeralKeyPair`, `checkSSMDependencies`, `buildSSHArgs`
- `ssh.go` — `runConnect` (orchestrator), `ec2icSender` interface

`runConnect` is the core routine used by both `ssh` and `proxy`. It calls `BastionService.DescribeBastion` → `WaitForSSMReady` → generates and uploads an ephemeral key → writes a temp file → assembles SSH args. The cmd layer then execs `ssh` with those args. Because all of this sits in `package main`, it cannot be imported by other packages and is tested only through integration-style tests that stub at the service boundary.

## Goals / Non-Goals

**Goals:**
- Move all pre-exec logic into a new `internal/connect/` package behind a `ConnectService` interface
- Make the service unit-testable without running real AWS calls or spawning processes
- Simplify `NewSSHCommand` and `NewProxyCommand` to: construct service → call `Prepare` → exec
- Keep exec, stdin/stdout/stderr wiring, and signal forwarding in the cmd layer (process management is not business logic)

**Non-Goals:**
- Changing any external CLI behaviour or AWS API calls
- Moving process exec into the service (would make it impossible to unit-test without spawning `ssh`)
- Combining `ConnectService` with `BastionService` (they have separate lifecycles and concerns)

## Decisions

### ConnectService interface

```go
// internal/connect/connect.go

type PrivateKeyWriter interface {
    io.WriteCloser
    Chmod(mode fs.FileMode) error
    Name() string
}

type PrepareInput struct {
    BastionName    string
    Region         string
    OSUser         string           // defaults to "ec2-user" if empty
    ExtraSSHArgs   []string         // e.g. ["-D", "1080", "-N"] for proxy mode
    PrivateKeyFile PrivateKeyWriter // caller-created file that receives the private key
}

type Connection struct {
    SSHArgs []string // full argument list ready to pass to exec.Command("ssh", ...)
}

type ConnectService interface {
    CheckDependencies() error
    Prepare(ctx context.Context, input *PrepareInput) (*Connection, error)
}
```

`Prepare` encapsulates: `DescribeBastion` → `WaitForSSMReady` → key generation → EC2IC upload → write private key to `input.PrivateKeyFile` → `buildSSHArgs`. It returns a `*Connection` the cmd layer uses to exec `ssh`.

**Why not split Prepare into two steps?** The cmd layer has no use for partial state (instance ID, AZ) independently. One call returning exec-ready args keeps the command handlers minimal.

**Why keep exec in cmd?** Signal forwarding, stdin/stdout/stderr wiring, and `cmd.Wait()` are process-management concerns that belong with Cobra. Keeping them in cmd also means `Prepare` has a clean return path and is straightforwardly testable.

### Package layout

```
internal/connect/
    connect.go         — ConnectService interface, PrepareInput, Connection types
    connectservice.go  — connectService struct and method implementations
    connect_test.go    — unit tests for ConnectService
```

Follows the pattern of `internal/bastion/` (interface + impl in the same package, tests alongside).

### Dependencies

`connectService` takes `BastionService` and an EC2IC client as constructor arguments, matching the existing pattern in `NewBastionService`. No new AWS clients are needed.

```go
func NewConnectService(bastionSvc bastion.BastionService, ec2ic ec2icSender) ConnectService
```

`dependencies.ConnectService(ctx)` constructs it by composing the existing `BastionService` and `EC2InstanceConnectClient` factories — no new AWS credentials or config are required.

### What moves out of cmd

| Current location | Destination |
|---|---|
| `generateEphemeralKeyPair()` | `internal/connect/connectservice.go` |
| `checkSSMDependencies()` | `internal/connect/connectservice.go` (becomes `CheckDependencies`) |
| `buildSSHArgs()` | `internal/connect/connectservice.go` |
| `runConnect()` | replaced by `connectSvc.Prepare()` call in each command |
| `ec2icSender` interface | `internal/connect/connectservice.go` |
| `connect.go`, relevant parts of `ssh.go` | deleted |

### What stays in cmd

- `os.CreateTemp` to create the key file, passed in as `input.PrivateKeyFile`
- `defer os.Remove(keyFile.Name())` to clean up after `cmd.Wait()`
- `exec.Command("ssh", conn.SSHArgs...)` with stdin/stdout/stderr passthrough
- Signal forwarding goroutine and context-cancellation escalation
- `onReady()` callback (e.g. proxy status message) called between `Prepare` and exec
- Cobra flag wiring

## Risks / Trade-offs

- **`buildSSHArgs` tests move** — existing tests in `cmd/bastion/connect_test.go` that test `buildSSHArgs` will move to `internal/connect/connect_test.go`. The cmd-layer tests will be simplified to mock `ConnectService` at the interface boundary.
