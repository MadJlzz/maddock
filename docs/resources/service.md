# service

Manages a `systemd` unit: whether it's running and whether it's enabled
at boot. The resource name is the unit name (the part before `.service`
— the `.service` suffix is implied).

## Attributes

| Attribute | Type   | Required | Description                                 |
|-----------|--------|----------|---------------------------------------------|
| `state`   | string | yes      | `running` to start, `stopped` to stop.      |
| `enabled` | bool   | yes      | `true` to enable at boot, `false` to disable. |

Both attributes are always checked; you cannot manage only one of
them today.

## Behavior

Check runs two `systemctl` subcommands:

- `systemctl is-active <name>` — active ↔ running.
- `systemctl is-enabled <name>` — parses the output to distinguish
  `enabled`, `enabled-runtime`, `disabled`, etc. Anything producing an
  "enabled" variant is treated as enabled.

Apply only invokes the transitions needed:

- If running state differs: `systemctl start` or `systemctl stop`.
- If enablement differs: `systemctl enable` or `systemctl disable`.

If both are already in the desired state, apply is a no-op — same
idempotent property as the other resources.

## Requirements

- `systemd` must be the init system. Maddock does not support
  SysV-init, OpenRC, runit, etc.
- The agent must have permission to manage the service. For
  system units that means running as root.
- The unit must already exist on the host. This resource does **not**
  create `.service` files; pair it with the
  [file](./file) resource if you need to drop a
  unit file into `/etc/systemd/system/` first.

## Examples

### Run and enable a daemon

```yaml
- service:
    nginx:
      state: running
      enabled: true
```

### Stop and disable

```yaml
- service:
    apache2:
      state: stopped
      enabled: false
```

### Full pattern: install, configure, run

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

The order matters: the package provides the unit file, the file
resource drops the config, and the service resource starts the daemon
last.

## Gotchas

- **No `daemon-reload`.** If you drop a new unit file via the file
  resource, systemd may not pick up the change until a reload. Today,
  Maddock does not run `systemctl daemon-reload` for you. This is a
  known gap and will be addressed as a future improvement, likely via
  a dedicated attribute or a separate resource type.
- **Configuration changes do not restart services.** The service
  resource only cares about *running or not*. If a config file
  changes but the service is already running, Maddock will report
  `OK` for the service. A future `notify`/`subscribe` mechanism will
  fix this.

## Known limitations

- User units (`systemctl --user`) are not supported.
- Masked units are treated the same as disabled units. Unmasking is
  not implemented.
- No support for templated units (e.g. `getty@tty1.service`) beyond
  passing the full name as the resource name.
