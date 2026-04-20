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
`maddock-server`.

```sh
maddock-agent serve [--listen :9600]
```

### Flags

`--listen`
: Address and port to bind, default `:9600`.

The server stays running until killed.
