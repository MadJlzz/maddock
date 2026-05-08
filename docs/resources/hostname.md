# hostname

Sets the system's static hostname via `hostnamectl`. The resource name
is a user-chosen identifier; the actual hostname is specified through
the `name` attribute.

## Prerequisites

- **systemd** must be the init system. This resource uses `hostnamectl`
  under the hood, which is part of the `systemd` suite. It will not
  work on systems using SysV-init, OpenRC, or other init systems.
- The agent must have permission to change the hostname. In practice
  that means running as root.

## Attributes

| Attribute | Type   | Required | Description                        |
|-----------|--------|----------|------------------------------------|
| `name`    | string | yes      | The desired static hostname.       |

## Behavior

**Check** runs `hostnamectl --static` and compares the output to the
desired `name`. If they differ, a single `name` difference is reported.

**Apply** runs `hostnamectl hostname <name>` to set the static
hostname. A non-zero exit code is treated as a failure.

## Examples

### Set the hostname

```yaml
- hostname:
    set-hostname:
      name: web1.example.com
```

### Full pattern: hostname, hosts file, and DNS

```yaml
name: host-identity

resources:
  - hostname:
      set-hostname:
        name: web1.example.com

  - file:
      /etc/hostname:
        content: "web1.example.com\n"
        owner: root
        group: root
        mode: "0644"

  - file:
      /etc/hosts:
        content: |
          127.0.0.1 localhost
          127.0.1.1 web1.example.com web1
        owner: root
        group: root
        mode: "0644"
```

## Gotchas

- **Only the static hostname is managed.** `hostnamectl` distinguishes
  between static, transient, and pretty hostnames. This resource only
  reads and sets the static hostname.
- **`/etc/hostname` is not updated.** `hostnamectl hostname` persists
  the change to `/etc/hostname` on most distributions, but if you need
  to guarantee the file content (e.g. for tools that read it directly),
  pair this resource with a [file](./file) resource as shown above.

## Known limitations

- No support for setting the pretty or transient hostname.
- No support for non-systemd systems.
