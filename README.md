# certfix

Command-line interface for managing the full lifecycle of services, certificates, policies, and events in [certfix-core](https://github.com/certfix/certfix-core). Built with Go + Cobra + Viper. Runs on Linux, macOS, and Windows.

---

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Building from Source](#building-from-source)
- [Configuration](#configuration)
- [Authentication](#authentication)
- [Commands](#commands)
  - [Auth](#auth)
  - [Services](#services)
  - [Service Groups](#service-groups)
  - [Policies](#policies)
  - [Certificates](#certificates)
  - [Service Keys](#service-keys)
  - [Events](#events)
  - [Service Matrix](#service-matrix)
  - [Apply](#apply)
- [YAML Config Format](#yaml-config-format)
- [Development](#development)
- [Project Structure](#project-structure)

---

## Installation

Download the binary for your platform from the [releases page](https://github.com/certfix/certfix-cli/releases/latest):

```bash
# Linux (amd64)
wget https://github.com/certfix/certfix-cli/releases/latest/download/certfix-linux-amd64
chmod +x certfix-linux-amd64
sudo mv certfix-linux-amd64 /usr/local/bin/certfix

# Linux (arm64)
wget https://github.com/certfix/certfix-cli/releases/latest/download/certfix-linux-arm64
chmod +x certfix-linux-arm64
sudo mv certfix-linux-arm64 /usr/local/bin/certfix

# macOS (Intel)
wget https://github.com/certfix/certfix-cli/releases/latest/download/certfix-darwin-amd64
chmod +x certfix-darwin-amd64
sudo mv certfix-darwin-amd64 /usr/local/bin/certfix

# macOS (Apple Silicon)
wget https://github.com/certfix/certfix-cli/releases/latest/download/certfix-darwin-arm64
chmod +x certfix-darwin-arm64
sudo mv certfix-darwin-arm64 /usr/local/bin/certfix

# Windows — download certfix-windows-amd64.exe and add to PATH
```

Verify:

```bash
certfix version
```

---

## Quick Start

```bash
# 1. Point the CLI at your certfix-core instance
certfix configure --api-url https://certfix.example.com

# 2. Log in with your email and personal access token
certfix login --email you@example.com --token <personal-access-token>

# 3. Start managing resources
certfix services list
certfix policy list
certfix certs list <service-hash>
```

---

## Building from Source

**Prerequisites:** Go 1.24+, Make

```bash
git clone https://github.com/certfix/certfix-cli.git
cd certfix-cli

make deps        # Download dependencies
make build       # Build for current platform → bin/certfix
make build-all   # Cross-compile for all platforms → dist/
make test        # Run tests
make lint        # Run linter
make clean       # Remove bin/ and dist/
```

**All Makefile targets:**

| Target | Description |
|--------|-------------|
| `build` | Build for current platform → `bin/certfix` |
| `build-all` | Cross-compile for Linux, macOS, Windows (amd64 + arm64) |
| `build-linux` | Linux amd64 only |
| `build-darwin` | macOS amd64 + arm64 |
| `build-windows` | Windows amd64 only |
| `deps` | `go mod download` + `go mod tidy` |
| `test` | `go test -v ./...` |
| `lint` | golangci-lint (falls back to `go fmt` + `go vet`) |
| `install` | Install to `$GOBIN` |
| `dev` | `go run main.go` |
| `run` | Build + run |
| `clean` | Remove `bin/` and `dist/` |
| `version` | Print current version |

The version string is injected at build time via `-X main.Version=$(VERSION)` from the first line of the Makefile. To release a new version, update `VERSION` in the Makefile, build, and tag.

---

## Configuration

Configuration is stored at `~/.certfix/config.yaml` and managed by the `configure` command.

```bash
certfix configure                          # Interactive wizard
certfix configure --api-url <url>          # Set endpoint non-interactively
certfix configure --show                   # Print current configuration
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--api-url` | `-a` | `https://certfix.io` | Base URL of certfix-core (without `/api`) |
| `--timeout` | `-t` | 30 | HTTP request timeout in seconds |
| `--retry-attempts` | `-r` | 3 | Number of retries on failure |
| `--show` | `-s` | — | Print current configuration and exit |

> The CLI appends `/api/v0.1.0` to the configured URL automatically. Set `--api-url http://localhost:3001` for local development.

---

## Authentication

```bash
certfix login                                                  # Interactive
certfix login --email you@example.com --token <pat>           # Non-interactive
certfix whoami                                                  # Confirm identity
certfix logout
```

**How it works:**

1. `login` calls `POST /auth/cli` with your email and personal access token
2. The returned JWT is saved to `~/.certfix/token.json` (mode `0600`)
3. All subsequent commands attach it as `Authorization: Bearer <token>`
4. On expiry the CLI prints: *"token expired: please run 'certfix login'"*

---

## Commands

All commands accept `--verbose` / `-v` for debug output and `--output` / `-o table|json` where applicable.

---

### Auth

```bash
certfix login [--email <email>] [--token <pat>]
certfix logout
certfix whoami [--output table|json]
certfix version
```

---

### Services

```bash
# List
certfix services list [--active] [--group <group-id>] [--output table|json]

# Get
certfix services get <service-hash> [--output table|json]

# Create
certfix services create \
  --name <name> \
  [--hash <custom-hash>] \
  [--webhook <url>] \
  [--group <group-id>] \
  [--policy <policy-id>] \
  [--dns api.example.com,svc.internal] \
  [--active] \
  [--output table|json]

# Update
certfix services update <service-hash> \
  [--name <name>] \
  [--webhook <url>] [--clear-webhook] \
  [--group <group-id>] [--clear-group] \
  [--policy <policy-id>] [--clear-policy] \
  [--dns <names>] [--clear-dns] \
  [--active] \
  [--output table|json]

# Lifecycle
certfix services activate <service-hash>
certfix services deactivate <service-hash>
certfix services delete <service-hash> [--force]

# Certificate operations
certfix services rotate <hash>[,<hash>,...]         # Trigger rotation
certfix services generate-hash <service-name>        # Preview hash for a name
```

**Aliases:** `service`, `svc`

---

### Service Groups

```bash
certfix service-groups list [--enabled] [--output table|json]
certfix service-groups get <group-id> [--output table|json]

certfix service-groups create \
  --name <name> \
  [--description <text>] \
  [--enabled]

certfix service-groups update <group-id> \
  [--name <name>] \
  [--description <text>] \
  [--enabled]

certfix service-groups enable <group-id>
certfix service-groups disable <group-id>
certfix service-groups delete <group-id> [--force]
```

**Aliases:** `service-group`, `svc-group`, `svc-groups`

---

### Policies

```bash
certfix policy list [--strategy <strategy>] [--enabled] [--output table|json]
certfix policy get <policy-id> [--output table|json]

certfix policy create \
  --name <name> \
  --strategy "Gradual|Maintenance Window|Events" \
  [--enabled] \
  # Cron (Gradual / Maintenance Window):
  [--cron-minute <0-59|*>] \
  [--cron-hour <0-23|*>] \
  [--cron-day <1-31|*>] \
  [--cron-month <1-12|*>] \
  [--cron-weekday <0-7|*>] \
  # Event-based:
  [--event-id <event-id>] \
  [--event-total <count>]

certfix policy update <policy-id> [same flags as create, all optional]

certfix policy enable <policy-id>
certfix policy disable <policy-id>
certfix policy delete <policy-id> [--force]
```

**Aliases:** `policies`, `politica`, `politicas`

**Strategies:**
- `Gradual` — zero-downtime; new cert issued, old revoked after all agents confirm
- `Maintenance Window` — cron-scheduled; brief downtime during swap
- `Events` — rotation triggered after N occurrences of a named event

---

### Certificates

```bash
certfix certs list <service-hash> [--output table|json]
certfix certs get <unique-id> [--output table|json]
certfix certs revoke <unique-id> \
  [--reason cessationOfOperation|superseded|keyCompromise] \
  [--force] \
  [--output table|json]
```

**Aliases:** `cert`, `certificate`, `certificates`

---

### Service Keys

API keys are scoped to a service and used by agents to authenticate.

```bash
certfix keys list <service-hash> [--output table|json]
certfix keys get <service-hash> [--output table|json]

certfix keys add <service-hash> \
  --name <name> \
  --expiration <days>           # e.g. 365

certfix keys enable <service-hash> <key-id>
certfix keys disable <service-hash> <key-id>
certfix keys toggle <service-hash> <key-id>
certfix keys delete <service-hash> <key-id> [--force]
```

**Alias:** `key`

---

### Events

Events are counters that can trigger a policy rotation when a threshold is reached.

```bash
certfix events list [--severity low|medium|high|critical] [--enabled] [--output table|json]
certfix events get <event-id> [--output table|json]

certfix events create \
  --name <name> \
  --severity low|medium|high|critical \
  [--enabled] \
  [--reset-unit minutes|hours|days] \
  [--reset-value <count>]        # 0 = never reset

certfix events update <event-id> [same flags as create, all optional]

certfix events enable <event-id>
certfix events disable <event-id>
certfix events delete <event-id> [--force]
```

**Aliases:** `event`, `eventos`, `evento`

---

### Service Matrix

The matrix defines which services communicate with each other (used for mTLS client certificate generation).

```bash
certfix matrix list <service-hash> [--output table|json]
certfix matrix get <service-hash> [--output table|json]

certfix matrix add <source-hash> <related-hash>

certfix matrix enable <service-hash> <relation-id>
certfix matrix disable <service-hash> <relation-id>
certfix matrix toggle <service-hash> <relation-id>
certfix matrix delete <service-hash> <relation-id> [--force]
```

**Alias:** `matriz`

---

### Apply

Declaratively create all resources from a YAML file. Resources are created in dependency order: events → policies → service groups → services → keys → relations.

```bash
certfix apply config.yml              # Apply and exit on first error (with rollback)
certfix apply config.yml --dry-run    # Preview changes without creating anything
certfix apply config.yml --skip-existing  # Ignore already-existing resources
```

On error, all resources created in the current run are automatically deleted in reverse order.

---

## YAML Config Format

```yaml
events:
  - name: "high-error-rate"
    severity: "high"        # low | medium | high | critical
    enabled: true
    reset_unit: "hours"     # minutes | hours | days
    reset_value: 24         # 0 = never reset

policies:
  - name: "gradual-nightly"
    strategy: "gradual"     # gradual | maintenance_window | events
    enabled: true
    cron_config:            # required for gradual and maintenance_window
      minute: "0"
      hour: "2"
      day: "*"
      month: "*"
      weekday: "*"

  - name: "event-driven"
    strategy: "events"
    enabled: true
    event_config:
      event_id: "high-error-rate"   # must match an event name above
      total_events: 10

service_groups:
  - name: "backend"
    description: "Core backend services"
    enabled: true

services:
  - name: "payments-api"
    hash: "payments-api"            # optional; generated from name if omitted
    active: true
    webhook_url: "https://notify.example.com/hook"
    group_name: "backend"           # must match a service_groups[].name above
    policy_name: "gradual-nightly"  # must match a policies[].name above
    dns_names:
      - "payments.internal"
      - "payments.example.com"
    keys:
      - name: "prod-agent-key"
        enabled: true
        expiration_days: 365
    relations:
      - target_hash: "auth-api"     # must be an existing service hash
```

Full working example: [`yml-certfix-config.yml`](yml-certfix-config.yml)

---

## Development

```bash
# Run without building
make dev -- services list

# Or directly with go
go run main.go configure
go run main.go services list

# Test against local certfix-core
certfix configure --api-url http://localhost:3001
certfix login --email admin@example.com --token <pat>

# Enable verbose output for any command
certfix --verbose services list
```

**Token and config files:**

| File | Purpose | Permissions |
|------|---------|-------------|
| `~/.certfix/config.yaml` | API endpoint, timeout, retry settings | `0700` directory |
| `~/.certfix/token.json` | Stored JWT and expiry | `0600` |

---

## Project Structure

```
certfix-cli/
├── main.go                     # Entry point — cobra root initialization
├── cmd/certfix/                # One file per command group (~22 files)
│   ├── root.go                 # Root command, --verbose flag, config init
│   ├── login.go / logout.go / whoami.go
│   ├── configure.go / version.go
│   ├── services.go             # services list/get/create/update/delete/rotate
│   ├── serviceGroups.go
│   ├── policy.go
│   ├── certificates.go
│   ├── keys.go
│   ├── events.go
│   ├── matrix.go
│   ├── apply.go                # YAML declarative apply with rollback
│   └── ...
├── internal/
│   ├── auth/auth.go            # Login, token storage/retrieval, logout
│   └── config/config.go        # Viper wrapper: read/write ~/.certfix/config.yaml
├── pkg/
│   ├── client/client.go        # HTTP client: GET/POST/PUT/PATCH/DELETE + auth headers
│   ├── logger/logger.go        # Logrus init (verbose → DEBUG, default → WARN)
│   └── models/models.go        # Structs for YAML apply format + rollback tracking
├── DOCS/
│   └── CLI_REFERENCE.md        # Full command reference with examples
├── yml-certfix-config.yml      # Example YAML for `certfix apply`
├── Makefile
├── go.mod                      # Module: github.com/certfix/certfix-cli, Go 1.24
└── go.sum
```

**Key dependencies:**

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI command structure and flag parsing |
| `github.com/spf13/viper` | Configuration file management |
| `github.com/golang-jwt/jwt/v5` | JWT parsing for token expiry |
| `github.com/sirupsen/logrus` | Structured logging |
| `gopkg.in/yaml.v3` | YAML parsing for `apply` command |
| `golang.org/x/term` | Secure password/token input |
