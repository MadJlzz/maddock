# Architecture

Maddock has two components:

- **Agent** — a long-lived daemon (or one-shot CLI) that runs on every
  target host. It knows how to check and apply resources against the
  local system.
- **Server** — an orchestration process that reads a central config,
  dispatches catalogs to agents over gRPC, and aggregates their
  reports.

## Push model

```
          YAML manifests
                │
                ▼
       ┌─────────────────┐        gRPC        ┌──────────────────┐
       │  maddock-server │ ─────ApplyCatalog──▶│   maddock-agent  │
       │                 │                     │                  │
       │  reads config,  │ ◄───stream reports──│  runs engine     │
       │  fans out       │                     │  against host    │
       └─────────────────┘                     └──────────────────┘
```

The server does not execute resources directly. It dials each agent,
forwards the catalog (serialized as protobuf messages), and streams
per-resource reports back to the operator.

## Local mode

The agent also runs without a server: `maddock-agent apply
manifest.yaml` parses the YAML in process and runs the same engine
used for gRPC pushes. This is the recommended path for trying Maddock
out.

## Resource lifecycle

For every resource in a catalog, the engine runs two phases:

**Check**
: Query the current state. Compare against the desired state. Produce
zero or more `Difference` objects describing attributes that disagree.

**Apply**
: If differences were found (and `--dry-run` was not passed), converge
the resource to the desired state.

The `ResourceError` type carries the phase along with the error, so
failures during check are distinguishable from failures during apply.
Reports include the phase in JSON output.

## Idempotency

The core invariant: applying the same manifest twice produces changes
on the first run and `OK` on the second. Every resource implementation
enforces this through the check-then-apply pattern. `--dry-run` runs
the check phase only.
