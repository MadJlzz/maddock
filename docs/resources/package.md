# package

Manages a system package through the host's native package manager.

## Supported package managers

Maddock auto-detects the package manager at agent startup by probing
`$PATH` in this order:

1. `dnf` — Fedora, Rocky, Alma, RHEL 8+.
2. `apt` — Debian, Ubuntu.

If neither is found, the agent fails fast at startup with a clear
error. There is no way to override the choice today; the expectation
is that each target host has exactly one relevant package manager.

## Attributes

| Attribute | Type   | Required | Description                                  |
|-----------|--------|----------|----------------------------------------------|
| `state`   | string | yes      | `present` to install, `absent` to remove.    |

## Behavior

- **Check** queries the package database:
  - `dpkg-query --status <pkg>` on apt.
  - `rpm --query <pkg>` on dnf.

  If the binary exits `0`, the package is considered installed.

- **Apply** invokes the manager non-interactively:
  - `apt-get install --yes <pkg>` / `apt-get remove --yes <pkg>`.
  - `dnf install --assumeyes <pkg>` / `dnf remove --assumeyes <pkg>`.

  Apply only runs if check shows a mismatch, so running the same
  manifest twice is idempotent.

## Requirements

- The agent must run as a user that can call the package manager
  (typically root). `apt-get` and `dnf` both need write access to
  system paths.
- On Debian/Ubuntu targets, `apt-get update` is **not** run
  automatically — make sure the package index is recent enough to
  resolve the packages you reference.

## Examples

### Install a package

```yaml
- package:
    htop:
      state: present
```

### Remove a package

```yaml
- package:
    vim-tiny:
      state: absent
```

### Several packages in one manifest

```yaml
- package:
    curl:
      state: present
    htop:
      state: present
    git:
      state: present
```

Each entry under `package:` is an independent resource, so the report
shows one line per package.

## Known limitations

- **No version pinning.** Only `present` / `absent` are supported. If
  you need a specific version, that's a future improvement.
- **No package groups or virtual packages.** Only concrete package
  names are tested.
- **No repository management.** Adding a custom repo is outside the
  scope of the resource; use your base-image or configuration
  management of choice.
