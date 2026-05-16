# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Bastion is a CLI tool that provisions EC2 bastion hosts on AWS via CloudFormation. Built with Cobra (CLI) and Viper (config).

## Build & Run

```bash
go build -o bastion ./cmd/bastion
./bastion up
```

## Lint

```bash
golangci-lint run ./...
```

Install: `mise use -g golangci-lint` (or see golangci-lint docs). Config is in `.golangci.yml`.

## Testing

Unit tests (no external dependencies):
```bash
go test ./...
```

Integration tests (require Docker — spin up LocalStack containers):
```bash
go test -tags integration ./...
```

Single test:
```bash
go test -run TestName ./path/to/package -v
```

Integration tests are gated with `//go:build integration` and use testcontainers-go + LocalStack.

### Testing approach

- Unit tests use mocks (interfaces defined per-package); add these first.
- Integration tests via LocalStack for critical paths only (e.g., CloudFormation deploy flow).
- Test helpers live in `testhelpers/` (LocalStack container setup, utilities).

## Project Structure

```
cmd/bastion/          CLI entry points (main.go, up.go, down.go)
internal/bastion/     Core BastionService — CloudFormation template embedded via //go:embed
internal/config/      Viper/Cobra config wiring
internal/dependencies/ AWS client construction (STS, CloudFormation)
pkg/aws/              Public AWS service wrappers (one package per service)
testhelpers/          Shared test utilities (LocalStack setup)
```

New AWS service wrappers go in `pkg/aws/<servicename>/` — see `pkg/aws/cloudformationservice/` as the canonical example.

## Configuration

Config resolution order: `BASTION_*` env vars → config file → defaults. Config file locations: `./config.yaml` or `~/.config/bastion/config.yaml`.

AWS credentials use the SDK v2 default chain (env vars, `~/.aws/credentials`, instance profile).

## Conventions

- Feature branches (`feature/`, `fix/`, etc.) with PRs to `main`.
- CloudFormation templates embedded in Go binaries with `//go:embed`.
- Change-set strategy for CloudFormation: detect create vs. update automatically.
