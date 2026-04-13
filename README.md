# maddock

An infrastructure-as-code tool written in Go that converges Linux machines to a desired state. Define your packages, config files, and services in YAML manifests, and Maddock ensures they're applied idempotently.

## Status

**Work in progress** -- Phases 1-3 are complete. The local agent is fully functional. Currently working on Phase 4 (gRPC transport). See [PLAN.md](PLAN.md) for the full roadmap.

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Core abstractions & project skeleton | Done |
| 2 | Resource implementations (package, file, service) | Done |
| 3 | YAML parser, engine & agent CLI | Done |
| 4 | gRPC transport | In progress |
| 5 | Server binary (push orchestration) | Not started |
| 6 | Polish & hardening | Not started |

## Architecture

- **Push mode**: `maddock-server` pushes catalogs to agents over gRPC
- **Local mode**: `maddock-agent apply manifest.yaml` for standalone use
- **Resources**: Packages (apt/dnf), Files (content/templates), Services (systemd)

## Prerequisites

- [mise](https://mise.jdx.dev/) for tooling management

## Getting started

```bash
# Install tools (Go, golangci-lint, protoc)
mise install

# Build
mise run build

# Test
mise run test

# Integration tests (requires Docker)
mise run test:integration

# Lint
mise run lint

# Generate protobuf code
mise run proto

# Run all checks (lint, build, test)
mise run check
```

## Project structure

```
cmd/
  agent/          # maddock-agent binary
  server/         # maddock-server binary
internal/
  resource/       # Resource interface, registry
  resources/      # Resource implementations (pkg, file, service)
  catalog/        # YAML manifest parser
  engine/         # Apply engine (check/apply loop)
  report/         # Run report types and formatters
  transport/      # gRPC server and client
  util/           # Command runner interface
```
