## 1. New `internal/connect` package

- [ ] 1.1 Create `internal/connect/connect.go` — define `ConnectService` interface, `PrepareInput`, and `Connection` types
- [ ] 1.2 Create `internal/connect/connectservice.go` — implement `connectService` struct with `NewConnectService(bastionSvc, ec2icSender)` constructor
- [ ] 1.3 Implement `CheckDependencies()` — move `checkSSMDependencies` logic from `cmd/bastion/connect.go`
- [ ] 1.4 Implement `Prepare()` — move and recompose `runConnect` + `generateEphemeralKeyPair` + `buildSSHArgs` logic from `cmd/bastion/`
- [ ] 1.5 Write unit tests in `internal/connect/connect_test.go` — cover all scenarios from `connect-service` spec (CheckDependencies, Prepare success/error paths, OnReady ordering, SSH arg assembly, default OSUser)

## 2. Wire `ConnectService` into `dependencies`

- [ ] 2.1 Add `connectService` field to `Dependencies` struct in `internal/dependencies/dependencies.go`
- [ ] 2.2 Add `ConnectService(ctx) (connect.ConnectService, error)` factory method that composes `BastionService` and `EC2InstanceConnectClient`

## 3. Update `ssh` and `proxy` commands

- [ ] 3.1 Rewrite `NewSSHCommand` RunE to use `deps.ConnectService` — call `CheckDependencies`, then `Prepare`, then exec `ssh` with `conn.SSHArgs`; defer `os.Remove(conn.KeyPath)`
- [ ] 3.2 Rewrite `NewProxyCommand` RunE to use `deps.ConnectService` — same flow with `ExtraSSHArgs: ["-D", port, "-N"]` and `OnReady` printing the status line
- [ ] 3.3 Delete `cmd/bastion/connect.go`
- [ ] 3.4 Remove `runConnect`, `buildSSHArgs`, `generateEphemeralKeyPair`, `checkSSMDependencies`, and `ec2icSender` from `cmd/bastion/ssh.go`

## 4. Update cmd-layer tests

- [ ] 4.1 Replace `connect_test.go` in `cmd/bastion/` with a mock-based `ConnectService` test that verifies `ssh` and `proxy` commands call `CheckDependencies` and `Prepare` correctly
- [ ] 4.2 Remove tests for `buildSSHArgs` and `runConnect` from `cmd/bastion/` (they move to `internal/connect/connect_test.go`)

## 5. Validation

- [ ] 5.1 Run `make lint` and fix any issues
- [ ] 5.2 Run `make test` and confirm all unit tests pass
