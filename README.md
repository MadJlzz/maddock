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

## Usage

Try a manifest locally:

```bash
maddock-agent apply manifest.yaml
```

To push from a control plane to remote agents, the wire is mutual TLS, so
first bootstrap a CA and certificates:

```bash
# Control plane: create the CA + control plane cert
maddock-controlplane init

# Issue a cert per agent host (hostname must match the push config)
maddock-controlplane cert issue --hostname web-1 --output ./web-1/

# On the agent host, with the copied certs:
maddock-agent serve --ca-cert ca.crt --cert web-1.crt --key web-1.key

# Back on the control plane:
maddock-controlplane push
```

See [docs/installation.md](docs/installation.md#bootstrap-tls) for the full
bootstrap walkthrough and [docs/architecture.md](docs/architecture.md#security-mtls-on-the-push-path)
for the trust model.
