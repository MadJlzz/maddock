# maddock

An infrastructure-as-code tool written in Go that converges Linux machines to a desired state. Define your packages, config files, and services in YAML manifests, and Maddock ensures they're applied idempotently.

## Status

**Work in progress** -- Phase 1 (core abstractions) is complete. See [PLAN.md](PLAN.md) for the full roadmap.

## Architecture

- **Push mode**: `maddock-server` pushes catalogs to agents over gRPC
- **Local mode**: `maddock-agent apply manifest.yaml` for standalone use
- **Resources**: Packages (apt/dnf), Files (content/templates), Services (systemd)

## Prerequisites

- [mise](https://mise.jdx.dev/) for tooling management

## Getting started

```bash
# Install tools (Go, golangci-lint)
mise install

# Build
mise run build

# Test
mise run test

# Lint
mise run lint
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
