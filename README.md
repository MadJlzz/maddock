# maddock

An infrastructure-as-code tool written in Go that converges Linux machines to a desired state. Define your packages,
config files, and services in YAML manifests, and Maddock ensures they're applied idempotently.

## Prerequisites

- [mise](https://mise.jdx.dev/) for tooling management
- [go](https://go.dev/)

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
