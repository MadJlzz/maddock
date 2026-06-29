# Architecture

Maddock has two components:

- **Agent** — a long-lived daemon (or one-shot CLI) that runs on every
  target host. It knows how to check and apply resources against the
  local system.
- **Control plane** — an orchestration process that reads a central
  config, dispatches catalogs to agents over gRPC, and aggregates
  their reports.

## Push model

```
          YAML manifests
                │
                ▼
       ┌────────────────────┐    gRPC over mTLS   ┌──────────────────┐
       │maddock-controlplane│────ApplyCatalog────▶│   maddock-agent  │
       │                    │                     │                  │
       │  reads config,     │◄────stream reports──│  runs engine     │
       │  fans out          │                     │  against host    │
       └────────────────────┘                     └──────────────────┘
```

The control plane does not execute resources directly. It dials each agent,
forwards the catalog (serialized as protobuf messages), and streams
per-resource reports back to the operator.

## Security: mTLS on the push path

The push connection is always mutual TLS (TLS 1.3 minimum) — there is no
plaintext mode and no `--insecure` escape hatch. Both sides authenticate:

- The **agent** presents a server certificate and requires the control
  plane to present a client certificate signed by the same CA
  (`tls.RequireAndVerifyClientCert`).
- The **control plane** verifies the agent's certificate against that CA
  and checks the certificate SAN against the target's configured
  `hostname` (not the dialed `address`, which is typically an IP). An
  agent presenting a certificate for the wrong hostname is rejected.

### Trust model

A single CA, created by `maddock-controlplane init`, anchors trust:

- **CA** (`ca.crt` / `ca.key`) — lives in the control plane state
  directory. The private key never leaves the control plane.
- **Control plane certificate** (`controlplane.crt`) — issued at `init`,
  used as the client certificate when pushing.
- **Agent certificates** — issued per host with
  `maddock-controlplane cert issue --hostname <host>`. The CN/SAN is the
  hostname; the keypair is generated on the control plane and copied to
  the target.

Today certificate issuance and distribution are **manual** (issue on the
control plane, `scp` to the host). An automated agent join flow is
planned to replace the manual copy step; see
[the roadmap](https://github.com/MadJlzz/maddock/blob/main/PLAN.md).

## Local mode

The agent also runs without the control plane: `maddock-agent apply
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

## Design decisions

A few choices that shape the project, and the reasoning behind them.

**Config format: YAML.** Familiar to anyone coming from Ansible,
Kubernetes, or Puppet Hiera. Good enough for declarative resource
lists; we're not trying to invent a new language.

**Execution model: push first.** The control plane dials the agent
and pushes a catalog, rather than the agent polling. Push gives the
operator immediate feedback (streamed reports) and makes ad-hoc
"apply this manifest to these hosts now" a first-class operation. The
architecture leaves room to add a pull mode later (agent periodically
fetches its catalog), but push is what's implemented.

**Target OS: Linux only.** Resources target `apt`/`dnf` and
`systemd`. Cross-platform abstractions (e.g. a common package
interface for macOS/Windows) add significant complexity for a
learning-focused project and dilute the core lessons. The door stays
open to add more Linux families without redesign.

**Ordering: implicit.** Resources run in the order they appear in the
manifest. No `require` / `before` / `notify` semantics today —
ordering is the author's responsibility. This keeps the engine simple
and debuggable at the cost of making some patterns (reload on config
change) more awkward. See [the roadmap](https://github.com/MadJlzz/maddock/blob/main/PLAN.md) for planned notify/handler support.

**Wire format: protobuf with JSON-encoded attributes.** Catalog
messages are protobuf structs, but each resource's attribute map is
serialized as JSON bytes inside a `bytes attributes = 3` field. This
keeps the protobuf schema stable as new resource types are added —
the control plane doesn't need to be rebuilt when the agent grows a
new resource.
