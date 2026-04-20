# command

Runs an arbitrary shell command on the target host. The resource name
is a user-chosen identifier (e.g. `install-node`) used in reports; it
does not refer to a system object.

::: warning Idempotency is your responsibility
Unlike `package`, `file`, and `service`, the `command` resource has no
natural notion of "desired state". Without a guard attribute, every
apply will execute the command and report `CHANGED`. Use one of the
guards below — `creates`, `unless`, or `onlyif` — to make the resource
idempotent.
:::

::: warning Security
The command runs as the agent's user. In typical deployments that
means **root**. Treat manifests as privileged code; anyone who can
write a manifest effectively has a shell on every target host.
:::

## Attributes

| Attribute | Type   | Required | Description                                                                 |
|-----------|--------|----------|-----------------------------------------------------------------------------|
| `command` | string | yes      | The shell command to run. Executed via `/bin/sh -c "<command>"`.            |
| `creates` | string | no       | Skip the command if this path already exists.                               |
| `unless`  | string | no       | Skip the command if this shell command exits `0`.                           |
| `onlyif`  | string | no       | Skip the command if this shell command exits non-zero.                      |

Guards are combined with AND: the command runs only if **every**
provided guard indicates "run".

## Behavior

**Check** evaluates the guards in order:

1. If `creates` is set and the path exists, the resource is already
   satisfied (reported as `OK`).
2. If `unless` is set, Maddock runs it through `/bin/sh -c` and treats
   exit `0` as "already done".
3. If `onlyif` is set, Maddock runs it and treats exit non-zero as
   "precondition not met — skip".
4. Otherwise the resource reports `CHANGED` with a single `command`
   difference.

**Apply** runs the command through `/bin/sh -c "<command>"`. Exit
`0` → `CHANGED`. Non-zero exit → `FAILED`, with stderr included in the
error message.

With `--dry-run`, only the check phase runs. Resources that would
execute are reported as `SKIPPED` and the process exits with code `3`.

## Examples

### Idempotent via `creates`

The most common pattern: skip the command if a file it produces
already exists.

```yaml
- command:
    install-node:
      command: "curl -fsSL https://deb.nodesource.com/setup_lts.x | bash -"
      creates: /etc/apt/sources.list.d/nodesource.list
```

### Idempotent via `unless`

Useful when the "done" signal is a command's exit code rather than a
file on disk.

```yaml
- command:
    add-sudo-group:
      command: "usermod -aG sudo alice"
      unless: "id -nG alice | grep -qw sudo"
```

### Idempotent via `onlyif`

The inverse of `unless`: run only when a precondition is met.

```yaml
- command:
    reboot-if-needed:
      command: "systemctl reboot"
      onlyif: "test -f /var/run/reboot-required"
```

### Combining guards

Guards are AND-combined. Both must say "run" for the command to
execute.

```yaml
- command:
    provision-db:
      command: "psql -U postgres -c 'CREATE DATABASE app'"
      unless: "psql -U postgres -lqt | cut -d '|' -f 1 | grep -qw app"
      onlyif: "systemctl is-active --quiet postgresql"
```

## Gotchas

- **Shell semantics apply.** Pipes, redirects, and environment
  variable expansion all work because the command is invoked via
  `/bin/sh -c`. On most Linux distributions `/bin/sh` is
  bash-compatible, but scripts that rely on bash-only features should
  invoke bash explicitly: `command: "bash -c '...'"`.
- **No `daemon-reload` or restart hook.** Commands do not automatically
  reload systemd or restart services. Model that explicitly with a
  subsequent `service` resource or another `command`.
- **`creates` is not a trigger.** It only controls whether the command
  runs this time; it does not delete the path or refresh it.

## Known limitations

- No `cwd` / working directory (commands run in the agent's CWD).
- No `user` / sudo-style user switching (commands run as the agent's user).
- No `environment` / custom env vars (inherits the agent's environment).

These are planned additions; see the project roadmap.
