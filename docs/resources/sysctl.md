# sysctl

Manages kernel parameters by writing a drop-in file under
`/etc/sysctl.d/` and reloading it via `sysctl -p`. The resource name
is a user-chosen identifier (e.g. `disable-ipv6`); the parameters
themselves are listed under the `values` attribute.

## Prerequisites

- `sysctl` must be available on `PATH` (provided by `procps` on
  Debian/Ubuntu and `procps-ng` on RHEL/Fedora).
- The agent must be able to write to `/etc/sysctl.d/` and to
  `/proc/sys`. In practice that means running as root.

## Attributes

| Attribute  | Type              | Required | Description                                                                                  |
|------------|-------------------|----------|----------------------------------------------------------------------------------------------|
| `values`   | map[string]string | yes      | Kernel parameters and their desired runtime values. At least one entry is required.          |
| `filename` | string            | no       | Filename written under `/etc/sysctl.d/`. Defaults to `99-maddock.conf`.                      |

## Behavior

**Check** runs `sysctl --values <key>` for each key in `values` and
compares the trimmed output to the desired value. It also reads the
managed file under `/etc/sysctl.d/<filename>` and compares its
content to a deterministic, alphabetically-sorted rendering of
`values`. Any drift is reported:

- One `Difference` per kernel parameter whose runtime value does not
  match (with the parameter name as the attribute).
- One `Difference` with attribute `file` if the on-disk file is
  missing or its content differs.

**Apply** runs Check first. If nothing differs, the resource reports
`OK` without touching disk or running any command. Otherwise it
writes the rendered file atomically (via temp file + rename) with
mode `0644`, then runs `sysctl -p /etc/sysctl.d/<filename>` to load
the new values. A non-zero exit from `sysctl -p` is treated as a
failure.

## Examples

### Disable IPv6 on all interfaces

```yaml
- sysctl:
    disable-ipv6:
      values:
        net.ipv6.conf.all.disable_ipv6:     "1"
        net.ipv6.conf.lo.disable_ipv6:      "1"
        net.ipv6.conf.default.disable_ipv6: "1"
```

### Use a custom drop-in filename

```yaml
- sysctl:
    network-tuning:
      filename: "60-network.conf"
      values:
        net.core.somaxconn:        "4096"
        net.ipv4.tcp_max_syn_backlog: "4096"
```

## Gotchas

- **Always quote values.** YAML parses bare `1` or `0` as integers,
  but the resource expects strings. Quote (`"1"`) to avoid a parse
  error.
- **The file is rewritten in full on every change.** Do not edit
  `/etc/sysctl.d/<filename>` by hand — the next Apply will overwrite
  it. If you need parameters Maddock isn't managing, put them in a
  different drop-in file.
- **Other drop-in files take precedence by name order.** `sysctl -p`
  loads only this resource's file, so values set in higher-numbered
  drop-ins (e.g. `99-other.conf` if you name yours `60-foo.conf`)
  may override what Maddock sets at boot.
- **Literal string comparison.** Runtime drift is detected by
  comparing the trimmed output of `sysctl --values <key>` against
  the desired value byte-for-byte. Multi-value parameters
  (e.g. `net.ipv4.tcp_rmem`) must match the exact whitespace
  `sysctl` prints back.

## Known limitations

- One drop-in file per resource. To split parameters across multiple
  files, declare multiple `sysctl` resources with different
  `filename`s.
- The `--system` reload (which would pick up changes from every
  drop-in directory) is not used; only the resource's own file is
  reloaded.
