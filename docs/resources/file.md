# file

Manages a file's **content**, **owner**, **group**, and **mode**. The
resource name is the absolute path of the file on the target host.

Writes are **atomic**: Maddock writes content to a temporary file in
the same directory, sets permissions and ownership, then renames it
into place. Readers of the file never observe a partially-written
state.

## Attributes

| Attribute | Type   | Required | Description                                                                 |
|-----------|--------|----------|-----------------------------------------------------------------------------|
| `content` | string | one of   | Inline file content. Mutually exclusive with `source`.                      |
| `source`  | string | one of   | Path to a Go `text/template` file, rendered at parse time using `vars`.     |
| `vars`    | map    | no       | Template variables. Only used with `source`.                                |
| `owner`   | string | yes      | Unix username that should own the file.                                     |
| `group`   | string | yes      | Unix group name that should own the file.                                   |
| `mode`    | string | yes      | Octal permission string, e.g. `"0644"`. Must be quoted so YAML keeps it a string. |

Either `content` or `source` must be provided; supplying both is a
parse-time error.

## Behavior

The check phase stats the target path, then:

1. If the path does not exist, the resource is reported as **changed**
   with a single `path` difference, and apply creates the file.
2. Otherwise, Maddock compares:
   - The SHA-256 of the target file's bytes vs. the SHA-256 of the
     desired content (after template rendering if `source` was used).
   - The current owner and group (resolved by UID/GID) vs. the desired
     ones.
   - The current permission bits vs. the parsed `mode`.

   Each mismatch produces a separate `Difference` entry so reports can
   show exactly what changed.

The apply phase always writes a full new file (even if only the mode
needs changing) and then chmod/chown it. The temp-file-then-rename
pattern makes the operation atomic on the same filesystem.

## Requirements

- The agent needs write access to the target file's parent directory
  (for the temp file + rename), and the ability to `chown` (usually
  root). See [Architecture](/architecture) for notes on privilege.

## Examples

### Inline content

```yaml
- file:
    /etc/app.conf:
      content: |
        log_level=info
        port=8080
      owner: root
      group: root
      mode: "0644"
```

### Rendering from a Go template

Template file `templates/nginx.conf.tmpl`:

```text
worker_connections {{ .worker_connections }};
server_name {{ .server_name }};
```

Manifest:

```yaml
- file:
    /etc/nginx/nginx.conf:
      source: templates/nginx.conf.tmpl
      owner: root
      group: root
      mode: "0644"
      vars:
        worker_connections: 1024
        server_name: example.com
```

Template variables are accessed with the standard Go template syntax.
`vars` is passed in as a flat map; references look like:

```text
{{ .key }}
```

### Strict permissions

```yaml
- file:
    /etc/ssh/ssh_host_key:
      content: "..."
      owner: root
      group: root
      mode: "0600"
```

## Gotchas

- **Quote the mode.** `mode: 0644` is parsed by YAML as the integer
  420, which then fails to parse as an octal string. Always quote:
  `mode: "0644"`.
- **Templates are rendered at parse time**, not at apply time. Once
  the manifest is parsed, the content is fixed. Changing the template
  file or the `vars` map requires re-parsing the manifest (i.e.,
  running `apply` again).
- **Ownership is by name, not UID.** Maddock resolves `owner` and
  `group` at apply time. If the user does not exist on the host, the
  resource fails with a clear error during the apply phase.

## Known limitations

- No support for `state: absent` (deleting a file) yet. The file
  resource today is strictly "make this file exist with this
  content".
- No directory resource — parent directories must already exist.
- No backup/history of previous content; writes are destructive.
