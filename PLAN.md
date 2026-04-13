# Maddock — Infrastructure as Code Tool in Go

## Context

Build a learning-focused IaC tool (like Puppet/Salt/Chef) in Go. The goal is to get a target machine to a **"ready" state** — all packages installed, config files in place with correct content/permissions, and services running/enabled as specified. This is a greenfield project in an empty repo.

**Decisions made:**
- Config: YAML manifests
- Execution: Client-server, **push mode first** (server → agents via gRPC), architecture ready for pull mode
- Target OS: Linux only (apt/dnf, systemd)
- Resources: Packages, Files & templates, Services
- Ordering: Implicit (YAML order)

---

## Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│                         MADDOCK SYSTEM                               │
│                                                                      │
│  ┌─────────────────────┐            ┌──────────────────────────────┐ │
│  │   maddock-server     │   gRPC    │   maddock-agent (per host)   │ │
│  │                      │ ────────► │                              │ │
│  │  • Reads manifests   │  push     │  ┌────────────────────────┐  │ │
│  │  • Resolves targets  │  catalog  │  │     Apply Engine       │  │ │
│  │  • Pushes catalogs   │           │  │                        │  │ │
│  │  • Collects reports  │ ◄──────── │  │  Catalog (ordered)     │  │ │
│  │                      │  stream   │  │    │                   │  │ │
│  └────────┬─────────────┘  reports  │  │    ▼                   │  │ │
│           │                         │  │  Resource Runner       │  │ │
│           │ reads                   │  │  ┌───┐ ┌───┐ ┌───┐     │  │ │
│           ▼                         │  │  │Pkg│ │Fil│ │Svc│     │  │ │
│  ┌──────────────────────┐           │  │  └───┘ └───┘ └───┘     │  │ │
│  │   YAML Manifests     │           │  └────────────────────────┘  │ │
│  │  • host/target map   │           │           │                  │ │
│  │  • resource lists    │           │           ▼                  │ │
│  │  • template files    │           │  ┌────────────────────────┐  │ │
│  └──────────────────────┘           │  │  Report (per run)      │  │ │
│                                     │  └────────────────────────┘  │ │
│  Also usable standalone:            └──────────────────────────────┘ │
│  $ maddock-agent apply manifest.yaml   (local mode)                  │
└──────────────────────────────────────────────────────────────────────┘

Resource Abstraction:                  Data Flow (push):

  Resource (interface)                 YAML ─► Server parses
      │                                          │
      ├── PackageResource (apt/dnf)              ▼
      ├── FileResource (content/tmpl)        Catalog (protobuf)
      ├── ServiceResource (systemd)              │
      └── ... (future)                    gRPC push to agent
                                                 │
  Each resource implements:                      ▼
    Check() → is current == desired?       Engine.Apply()
    Apply() → converge to desired          [Check → Apply] per resource
                                                 │
                                                 ▼
                                           Report streamed back
```

---

## Project Structure

```
maddock/
  go.mod
  cmd/
    agent/main.go               # maddock-agent binary
    server/main.go              # maddock-server binary
  internal/
    resource/
      resource.go               # Resource interface, State, Difference types
      registry.go               # Type registry (Register + Parse)
    resources/
      pkg/pkg.go                # Package resource (apt/dnf)
      file/file.go              # File resource (content, templates, perms)
      service/service.go        # Service resource (systemd)
    catalog/
      catalog.go                # Catalog type (ordered resource list)
      parser.go                 # YAML → Catalog
    engine/
      engine.go                 # Apply engine (Check/Apply loop, dry-run)
    report/
      report.go                 # Report types, text + JSON formatters
    transport/
      proto/maddock.proto       # Protobuf service definition
      server.go                 # gRPC server (runs inside agent)
      client.go                 # gRPC client (used by server binary)
    util/
      exec.go                   # Command runner interface + mock
  testdata/
    webserver.yaml              # Sample manifests
```

---

## Phase 1 — Core Abstractions & Project Skeleton

**Goal:** Compilable project with the central `Resource` interface and registry.

### Steps

1. `go mod init github.com/jklaer/maddock` + create directory structure
2. Define `internal/resource/resource.go`:
   - `Resource` interface with `Type()`, `Name()`, `Check(ctx) (*CheckResult, error)`, `Apply(ctx) (*ApplyResult, error)`
   - `State` enum: `OK`, `Changed`, `Failed`, `Skipped`
   - `Difference` struct: `Attribute`, `Current`, `Desired`
   - `CheckResult` and `ApplyResult` structs
3. Define `internal/resource/registry.go`:
   - `Register(kind string, parseFunc)` — called from each resource's `init()`
   - `Parse(kind, name, attrs map[string]any) (Resource, error)` — used by YAML parser
4. Define `internal/util/exec.go`:
   - `Commander` interface: `Run(ctx, name, args...) (stdout, stderr, exitCode, error)`
   - `RealCommander` (wraps `os/exec`) and `MockCommander` (for tests)
5. Stub `cmd/agent/main.go` and `cmd/server/main.go` (just `func main()`)

**Test:** `go build ./...` compiles. Unit test the registry with a dummy resource.

---

## Phase 2 — Resource Implementations

**Goal:** The three core resources, each idempotent and testable via mock Commander.

### 2a. Package Resource (`internal/resources/pkg/`)

- `Manager` interface: `IsInstalled(ctx, pkg) (bool, version, error)`, `Install(ctx, pkg) error`, `Remove(ctx, pkg) error`
- `AptManager` — uses `dpkg-query` to check, `apt-get install -y` to install
- `DnfManager` — uses `rpm -q` to check, `dnf install -y` to install
- Auto-detect which manager at startup (check binary existence)
- **Idempotency:** `Check` queries package status → only `Apply` if not in desired state
- YAML attrs: `state: present | absent`

### 2b. File Resource (`internal/resources/file/`)

- Supports `content` (inline string) OR `source` (Go template file path) + `vars`
- Template rendering via `text/template` at parse time
- **Check:** compare sha256 of existing vs desired content, compare owner/group/mode
- **Apply:** write to temp file → chmod → chown → atomic rename
- YAML attrs: `content`, `source`, `vars`, `owner`, `group`, `mode`

### 2c. Service Resource (`internal/resources/service/`)

- **Check:** `systemctl is-active <name>` + `systemctl is-enabled <name>`
- **Apply:** `systemctl start|stop` + `systemctl enable|disable` as needed
- YAML attrs: `state: running | stopped`, `enabled: true | false`

**Test:** Unit tests with MockCommander for all three. No root required.

---

## Phase 3 — YAML Parser, Engine & Agent CLI

**Goal:** Working `maddock-agent apply manifest.yaml` — the first runnable deliverable.

### YAML Manifest Format

```yaml
name: webserver

resources:
  - package:
      nginx:
        state: present
      curl:
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

`resources` is an ordered YAML list. Each item is a single-key map (key = resource type). Under each type, keys are resource names.

### Steps

1. `internal/catalog/parser.go` — parse YAML → walk list → call `resource.Parse()` per item → build `Catalog`
2. `internal/engine/engine.go` — iterate catalog in order, call `Check()`, call `Apply()` if needed (skip in dry-run), build `Report`
3. `internal/report/report.go` — `Report` struct with `ResourceReport` per resource, text formatter, summary counts
4. Wire `cmd/agent/main.go`:
   - `maddock-agent apply <manifest.yaml>` — parse, engine.Apply, print report
   - `--dry-run` flag — Check-only mode
   - `--log-level` flag
   - Exit codes: 0=converged, 1=error, 2=failed resources, 3=dry-run found changes

### Expected Output

```
$ sudo maddock-agent apply webserver.yaml

Maddock Agent — applying: webserver
════════════════════════════════════
[1/4] package:nginx .............. CHANGED (installed 1.24.0)
[2/4] file:/etc/nginx/nginx.conf . CHANGED (content, mode)
[3/4] file:/var/www/index.html ... OK
[4/4] service:nginx .............. CHANGED (started, enabled)

Summary: 3 changed, 1 ok, 0 failed | 12.4s
```

**Test:** End-to-end on a VM/container. Run twice — second run should show all OK (idempotency proof).

---

## Phase 4 — gRPC Transport

**Goal:** Agent can listen for pushed catalogs over gRPC.

### Protobuf (`internal/transport/proto/maddock.proto`)

```protobuf
syntax = "proto3";
package maddock.v1;
option go_package = "github.com/jklaer/maddock/internal/transport/proto";

service AgentService {
  // Server pushes a catalog; agent streams back per-resource reports.
  rpc ApplyCatalog(CatalogRequest) returns (stream ResourceReportMsg);
  rpc Ping(PingRequest) returns (PingResponse);
}

message CatalogRequest {
  string manifest_name = 1;
  bool dry_run = 2;
  repeated ResourceMsg resources = 3;
}

message ResourceMsg {
  string type = 1;        // "package", "file", "service"
  string name = 2;        // "nginx", "/etc/nginx/nginx.conf"
  bytes attributes = 3;   // JSON-encoded attributes map
}

message ResourceReportMsg {
  string type = 1;
  string name = 2;
  State state = 3;
  repeated DifferenceMsg changes = 4;
  string error = 5;
  int64 duration_ms = 6;
}

enum State {
  STATE_OK = 0;
  STATE_CHANGED = 1;
  STATE_FAILED = 2;
  STATE_SKIPPED = 3;
}

message DifferenceMsg {
  string attribute = 1;
  string current = 2;
  string desired = 3;
}

message PingRequest {}
message PingResponse {
  string hostname = 1;
  string agent_version = 2;
}
```

- Attributes are JSON bytes (not protobuf Struct) — stable schema as new resource types are added
- Server-streaming lets the server see real-time progress per resource

### Steps

1. Write `maddock.proto`, generate Go code with `protoc`
2. `internal/transport/server.go` — gRPC server impl (runs inside agent): deserialize → build Catalog → engine.Apply → stream reports
3. `internal/transport/client.go` — gRPC client (used by server binary): serialize Catalog → dial agent → read stream → return reports
4. Add `maddock-agent --serve --listen :9600` mode

**Test:** Start agent in serve mode, push a catalog with `grpcurl` or a test client.

---

## Phase 5 — Server Binary (Push Orchestration)

**Goal:** Working `maddock-server push` that targets multiple hosts in parallel.

### Server Config

```yaml
server:
  listen: ":9500"

targets:
  - hostname: web1.example.com
    address: 10.0.1.10:9600
    manifest: manifests/webserver.yaml
  - hostname: db1.example.com
    address: 10.0.1.20:9600
    manifest: manifests/dbserver.yaml
```

### Steps

1. Parse server config
2. For each target (bounded by `--parallel N` semaphore):
   - Parse manifest → Catalog
   - Dial agent, push catalog via gRPC
   - Stream and print per-resource progress
3. Print summary table across all hosts
4. Flags: `--config`, `--dry-run`, `--target <hostname>`, `--parallel N`, `--output json`

**Test:** Full push to multiple containers. Dry-run across fleet. Idempotency: second push = all OK.

---

## Phase 6 — Polish & Hardening

- `--output json` for machine-readable reports
- `--fail-fast` mode (stop on first resource failure, default is continue)
- Proper error categorization (`ResourceError` with type/name/phase)
- Structured logging with `log/slog`
- Integration tests in Docker (ubuntu + fedora images)
- End-to-end test: start agent in container, push from server, verify state

---

## Future Improvements

- **Pass `CheckResult` to `Apply()`**: Change the `Resource` interface to `Apply(ctx, *CheckResult)` so resources can skip unchanged attributes instead of unconditionally reapplying everything. The engine already has the `CheckResult` — it just doesn't forward it.

---

## Implementation Order (step-by-step)

| # | What | Milestone |
|---|------|-----------|
| 1 | `go mod init`, dirs, `resource.go`, `registry.go` | Compiles, core types exist |
| 2 | `util/exec.go` (Commander + mock) | Testable command execution |
| 3 | `resources/pkg/` (apt + dnf) | Package resource + unit tests |
| 4 | `resources/file/` (content + templates) | File resource + unit tests |
| 5 | `resources/service/` (systemd) | Service resource + unit tests |
| 6 | `catalog/parser.go` (YAML → Catalog) | Can parse manifest files |
| 7 | `engine/engine.go` + `report/report.go` | Engine runs catalogs, produces reports |
| 8 | `cmd/agent/main.go` (local apply) | **Milestone: working local agent** |
| 9 | `maddock.proto` + codegen | Protobuf types generated |
| 10 | `transport/server.go` (agent gRPC) | Agent accepts pushed catalogs |
| 11 | `transport/client.go` (server dialer) | Server can talk to agents |
| 12 | `cmd/server/main.go` (push mode) | **Milestone: working push server** |
| 13 | Polish: JSON output, fail-fast, exit codes | Production-ready CLI |

---

## External Dependencies (minimal)

| Package | Purpose |
|---------|---------|
| `gopkg.in/yaml.v3` | YAML parsing |
| `google.golang.org/grpc` | gRPC framework |
| `google.golang.org/protobuf` | Protobuf runtime |

Everything else is Go stdlib (`text/template`, `os/exec`, `log/slog`, `crypto/sha256`).

---

## Verification

After each phase, test idempotency: **run the same manifest twice — second run must show all resources as OK with zero changes.** This is the core property of a convergence-based IaC tool.

- **Phase 3 (local agent):** `sudo maddock-agent apply webserver.yaml` on an Ubuntu/Fedora VM or Docker container
- **Phase 5 (push server):** `maddock-server push --config server.yaml` targeting 2+ containers
- **Dry-run at any phase:** `--dry-run` shows what *would* change, exit code 3 if changes needed, 0 if converged
