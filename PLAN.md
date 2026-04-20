# Maddock Roadmap

This is the living roadmap — what's planned but not yet built. Anything that was in the original build plan (core abstractions, resource implementations, gRPC transport, server binary, JSON output, structured logging) is now shipped; see the code under `internal/` and the documentation at `docs/` for the current state.

The items below are ranked by how often they block real manifests, based on an exercise porting a non-trivial provisioning role to Maddock. Each entry has a rough design sketch and a size estimate (XS / S / M / L).

---

## Near-term resources

### 1. `user` + `group` resources (L)

Provisioning any multi-user host without them is essentially impossible — they are the single biggest gap for real-world use.

**`user`** — attributes `name`, `uid?`, `shell?`, `groups?`, `home?`, `system?` (bool), `state: present|absent`. Check parses `/etc/passwd` (and `/etc/group` for supplementary groups). Apply uses `useradd` / `usermod` / `userdel`.

**`group`** — attributes `name`, `gid?`, `state: present|absent`. Check parses `/etc/group`. Apply uses `groupadd` / `groupdel`.

### 2. `cron` resource (S)

Scheduled maintenance tasks are common and today require a `file` + `/etc/cron.d/` workaround where idempotency is file-level rather than entry-level.

**Sketch:** `name`, `minute`, `hour`, `day?`, `month?`, `weekday?`, `user?`, `command`, `state`. Apply writes a one-line file under `/etc/cron.d/<name>`. Check diffs a canonical representation (ignoring whitespace).

### 3. `authorized_key` resource (M)

SSH key deployment per user. Depends on the `user` resource landing first.

**Sketch:** `user`, `key` (single string or list), `state`, `path?` (defaults to `~user/.ssh/authorized_keys`). Check reads the file, looks for exact key lines. Apply appends/removes. Handles `~user/.ssh` creation with correct mode (700) and file mode (600).

---

## Architectural work

### 4. Facts & conditional resources (L, architectural)

Real-world manifests often need to branch on OS family, distribution, virtualization type, etc. Today Maddock has no facts and no way to express conditions.

**Sketch:** the `Ping` RPC already returns hostname + version — extend it to return `facts: {os_family, distribution, distribution_version, virtualization, kernel, arch, hostname}`. The server evaluates a per-resource `when:` expression against those facts before packaging the catalog for that target. Keep the expression language tiny (`os_family == "Debian"`) to avoid reimplementing Jinja.

---

## Resource refinements

### 5. Native `sysctl` resource (S)

Workaroundable with `file` + `command` today. A native resource would own `/etc/sysctl.d/<name>.conf` and the reload, and report `CHANGED` with the specific keys that changed.

**Sketch:** `values: map[string]string`, `filename?` (default `99-maddock.conf`). Apply writes atomically, runs `sysctl -p`.

### 6. Native `hostname` resource (XS)

Wrapper around the current file + `hostnamectl` pair.

**Sketch:** `name`. Check runs `hostname`. Apply runs `hostnamectl set-hostname <name>` and updates `/etc/hostname`.

### 7. `apt_repository` + `apt_key` (M)

Common for installing third-party packages (Docker, Node, etc.). First-class resources would handle the signing-key dance, the `sources.list.d` file, and `apt-get update` in one unit. Today these are achievable via `file` + `command`, just awkwardly.

---

## Deferred

Items we acknowledge but aren't prioritizing. Workarounds exist; implementation cost is high relative to value.

### Notify / handlers (L, architectural)

The "reload service only when its config actually changed" pattern is impossible today. We approximate with guards like `onlyif: sshd -t`, which reload on every apply. That's noisy but correct.

A proper implementation needs engine changes (track which resources reported `CHANGED`, run a deferred handler pass) and catalog syntax (top-level `handlers:` block or a `notify:` attribute). The protobuf and report formats already carry enough information.

For now, the `command` resource with guards covers 90% of the use case. Revisit if real manifests consistently hit the ceiling.

### Pass `CheckResult` to `Apply()`

Change the `Resource` interface to `Apply(ctx, *CheckResult)` so resources can skip unchanged attributes instead of unconditionally reapplying everything. The engine already has the `CheckResult` — it just doesn't forward it.

### Privilege separation

Split the agent into an unprivileged network-facing process (gRPC listener) and a short-lived privileged helper that only runs during apply. The helper would be forked/exec'd with the validated catalog, apply it, and exit. This is defense in depth, not a hard boundary — if an attacker controls the catalog content, they still have root *within the constraints of the apply engine*. The real value is that the privileged helper can reject anything outside known resource types and validated inputs, so a compromised network process doesn't grant a full root shell.
