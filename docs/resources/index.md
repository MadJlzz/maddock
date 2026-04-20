# Resources

A manifest is an ordered list of resources. Each resource has a `type`,
a `name`, and a set of attributes. The engine processes them in the
order they appear in the YAML, which also determines the order in which
the report displays them.

```yaml
name: webserver

resources:
  - package:
      nginx:
        state: present

  - file:
      /etc/nginx/nginx.conf:
        source: templates/nginx.conf.tmpl
        owner: root
        group: root
        mode: "0644"
        vars:
          worker_connections: 1024

  - service:
      nginx:
        state: running
        enabled: true
```

## Shape of a resource

Each list item is a single-key map whose key is the resource **type**:

```yaml
- <type>:
    <name>:
      <attributes>
```

- The resource `name` identifies the instance — a package name, a file
  path, a service name. Two resources of the same type and name are
  not allowed in a single manifest.
- Attributes are type-specific. Unknown attributes are ignored at parse
  time; required-but-missing attributes cause a parse error.

## Available resource types

- [package](./package) — install or remove system packages via apt/dnf.
- [file](./file) — manage file content, templates, ownership, and mode.
- [service](./service) — start, stop, enable, or disable systemd units.
- [command](./command) — run an arbitrary shell command with idempotency guards.

## Ordering and dependencies

Resources run in the order they appear. Maddock does not currently
have explicit dependencies or `notify`/`subscribe` semantics — if a
config file needs to be in place *before* a service starts, put it
earlier in the list.

A typical pattern:

1. `package` — install the software.
2. `file` — drop config files into place.
3. `service` — start and enable the daemon.

## Check vs apply

Every resource implements two phases:

- **Check** — query the current state, compare to desired, return a
  list of differences.
- **Apply** — converge to the desired state.

With `--dry-run`, only check runs; any resource with pending changes
is reported as `SKIPPED` and the process exits with code `3`.
