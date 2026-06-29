# maddock-agent

The agent binary. Runs on every host that Maddock manages.

## Global flags

`--log-level`
: `debug`, `info`, `warn`, or `error`. Default: `info`. Logs are
written to stderr; primary output (reports) goes to stdout.

`--version`
: Print the version and exit.

## apply

Apply a manifest to the local host.

```sh
maddock-agent apply [--dry-run] [--output text|json] <manifest.yaml>
```

### Flags

`--dry-run`
: Run the check phase only. Resources with pending changes are
reported as `SKIPPED`. Exit code is `3` if any changes are pending.

`--output`
: `text` (default, human-readable) or `json` (machine-readable).

### Exit codes

- `0` — catalog converged; all resources are `OK` or `CHANGED`.
- `2` — one or more resources failed.
- `3` — `--dry-run` found pending changes.

## serve

Run the agent as a gRPC server, ready to accept catalogs pushed by
`maddock-controlplane`.

The wire is always mutual TLS — there is no plaintext mode. The agent
presents its own certificate and **requires** the control plane to
present a certificate signed by the same CA, so the three certificate
flags are all mandatory.

```sh
maddock-agent serve \
  --listen :9600 \
  --ca-cert ./ca.crt \
  --cert ./web-1.crt \
  --key ./web-1.key
```

### Flags

`--listen`
: Address and port to bind, default `:9600`.

`--ca-cert`
: Path to the CA certificate. Used as the trust anchor to verify the
control plane's client certificate. **Required.**

`--cert`
: Path to the agent's own server certificate. **Required.**

`--key`
: Path to the agent's own private key. **Required.**

The certificate, key, and CA are produced on the control plane with
`maddock-controlplane cert issue` and copied to the host (see
[maddock-controlplane](controlplane.md)). The agent stays running until
killed.
