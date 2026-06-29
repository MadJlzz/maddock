# maddock-controlplane

The orchestration binary. Reads a control plane config that maps targets to
manifests, then pushes each manifest to the corresponding agent over
gRPC.

## Global flags

`--log-level`
: `debug`, `info`, `warn`, or `error`. Default: `info`.

`--state-dir`
: Path to the control plane state directory, which holds the CA and the
control plane's own certificate. Defaults to
`/var/lib/maddock-controlplane` when running as root, otherwise
`~/.local/share/maddock-controlplane`.

`--version`
: Print the version and exit.

## init

Initialize a new control plane: create the state directory, generate the
CA, and issue the control plane's own certificate. Run this once before
issuing agent certificates or pushing.

```sh
maddock-controlplane init [--state-dir DIR]
```

This writes `ca.crt`, `ca.key`, `controlplane.crt`, and
`controlplane.key` into the state directory. It refuses to run if the
directory is already initialized.

## cert issue

Issue an agent certificate by hand, signed by the control plane CA. The
keypair is generated locally on the control plane; copy the output onto
the target host for `maddock-agent serve`.

```sh
maddock-controlplane cert issue --hostname web-1 --output ./web-1/ [--ttl 24h]
```

### Flags

`--hostname`
: Agent hostname. Used as the certificate CN/SAN **and** must match the
`hostname` field of the corresponding target in the push config, since
the control plane verifies the agent's certificate SAN against it.
Default: `agent`.

`--ttl`
: Certificate validity duration. Default: `24h`.

`--output`
: Directory to write `ca.crt`, `<hostname>.crt`, and `<hostname>.key`
into (created if missing). Default: `./`.

> This subcommand is a manual stopgap. An automated agent join flow is
> planned to replace the copy-certs-by-hand step.

## push

Push catalogs to one or more agents.

```sh
maddock-controlplane push [--config controlplane.yaml] [--dry-run] [--target HOST] [--parallel N] [--output text|json]
```

Push connects to each agent over mutual TLS, loading the CA and the
control plane's certificate from `--state-dir`. Run `init` first. The
connection verifies the agent's certificate SAN against the target's
`hostname` (not its `address`), so an agent presenting a certificate
issued for a different hostname is rejected with a TLS error.

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
