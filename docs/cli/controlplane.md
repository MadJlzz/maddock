# maddock-controlplane

The orchestration binary. Reads a control plane config that maps targets to
manifests, then pushes each manifest to the corresponding agent over
gRPC.

## Global flags

`--log-level`
: `debug`, `info`, `warn`, or `error`. Default: `info`.

`--version`
: Print the version and exit.

## push

Push catalogs to one or more agents.

```sh
maddock-controlplane push [--config controlplane.yaml] [--dry-run] [--target HOST] [--parallel N] [--output text|json]
```

### Flags

`--config`
: Path to the control plane config. Default: `controlplane.yaml`.

`--dry-run`
: Run check-only against every agent; no changes applied.

`--target`
: Only push to the target with this hostname.

`--parallel`
: Maximum concurrent pushes. Default: `4`.

`--output`
: `text` (default) or `json`.

### Exit codes

Exit codes mirror the agent's, aggregated across hosts:

- `0` — all hosts converged.
- `2` — any host had a transport error or a failed resource.
- `3` — any host had pending changes in dry-run mode.

## Control plane config

```yaml
controlplane:
  listen: ":9500"

targets:
  - hostname: web1.example.com
    address: 10.0.1.10:9600
    manifest: manifests/webserver.yaml
  - hostname: db1.example.com
    address: 10.0.1.20:9600
    manifest: manifests/dbserver.yaml
```

Manifest paths that are not absolute are resolved relative to the
config file's directory, so configs remain portable.
