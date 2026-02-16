# Certfix CLI - Complete Command Reference

This document provides comprehensive documentation for all available Certfix CLI commands, their parameters, and usage examples.

## Table of Contents

- [Installation](#installation)
- [Global Options](#global-options)
- [Configuration Commands](#configuration-commands)
- [Authentication Commands](#authentication-commands)
- [Apply Command](#apply-command)
- [Service Commands](#service-commands)
- [Service Group Commands](#service-group-commands)
- [Policy Commands](#policy-commands)
- [Event Commands](#event-commands)
- [API Key Commands](#api-key-commands)
- [Service Matrix Commands](#service-matrix-commands)

---

## Installation

### From Binary

Download the appropriate binary for your platform:

```bash
# Linux
wget https://github.com/certfix/certfix-cli/releases/download/v1.0.0/certfix-linux-amd64
chmod +x certfix-linux-amd64
sudo mv certfix-linux-amd64 /usr/local/bin/certfix

# macOS (Intel)
wget https://github.com/certfix/certfix-cli/releases/download/v1.0.0/certfix-darwin-amd64
chmod +x certfix-darwin-amd64
sudo mv certfix-darwin-amd64 /usr/local/bin/certfix

# macOS (Apple Silicon)
wget https://github.com/certfix/certfix-cli/releases/download/v1.0.0/certfix-darwin-arm64
chmod +x certfix-darwin-arm64
sudo mv certfix-darwin-arm64 /usr/local/bin/certfix
```

### Building from Source

```bash
git clone https://github.com/certfix/certfix-cli.git
cd certfix-cli
make build
```

---

## Global Options

The following flags are available for all commands:

| Flag        | Short | Description                         |
| ----------- | ----- | ----------------------------------- |
| `--verbose` | `-v`  | Enable verbose output for debugging |

### Example

```bash
certfix --verbose services list
```

---

## Configuration Commands

### `certfix configure`

Configure Certfix CLI settings including API endpoint, timeout, and retry attempts.

#### Usage

```bash
# Interactive mode (recommended for first-time setup)
certfix configure

# Non-interactive mode with flags
certfix configure [flags]

# Show current configuration
certfix configure --show
```

#### Flags

| Flag               | Short | Type   | Default | Description                                     |
| ------------------ | ----- | ------ | ------- | ----------------------------------------------- |
| `--show`           | `-s`  | bool   | false   | Show current configuration                      |
| `--api-url`        | `-a`  | string | -       | API endpoint URL (e.g., https://api.certfix.io) |
| `--timeout`        | `-t`  | int    | 30      | Request timeout in seconds                      |
| `--retry-attempts` | `-r`  | int    | 3       | Number of retry attempts for failed requests    |

#### Examples

```bash
# Interactive configuration (prompts for all settings)
certfix configure

# Configure API URL only
certfix configure --api-url https://api.certfix.io

# Configure all settings at once
certfix configure \
  --api-url https://api.certfix.io \
  --timeout 60 \
  --retry-attempts 5

# View current configuration
certfix configure --show
```

#### Configuration Storage

Configuration is stored in: `~/.certfix/config.yaml`

Example configuration file:

```yaml
endpoint: https://api.certfix.io
timeout: 30
retry_attempts: 3
```

---

## Authentication Commands

### `certfix login`

Authenticate with Certfix services using email and personal access token.

#### Usage

```bash
# Interactive mode (prompts for credentials)
certfix login

# Non-interactive mode with flags
certfix login --email <email> --token <token>
```

#### Flags

| Flag      | Short | Type   | Description                      |
| --------- | ----- | ------ | -------------------------------- |
| `--email` | `-e`  | string | Email address for authentication |
| `--token` | `-t`  | string | Personal access token            |

#### Examples

```bash
# Interactive login (recommended)
certfix login

# Login with flags
certfix login --email admin@example.com --token pat_abc123xyz...

# Login with short flags
certfix login -e admin@example.com -t pat_abc123xyz...
```

#### Prerequisites

- API endpoint must be configured first via `certfix configure`
- You need a valid personal access token from your Certfix account

#### Token Storage

Authentication tokens are stored securely in: `~/.certfix/token.json` with permissions `0600`

---

### `certfix logout`

Remove stored authentication token and log out from Certfix services.

#### Usage

```bash
certfix logout
```

#### Examples

```bash
# Logout from Certfix
certfix logout
```

---

### `certfix version`

Display the current version of Certfix CLI.

#### Usage

```bash
certfix version
```

#### Output

```
Certfix CLI v1.0.0
```

---

## Apply Command

### `certfix apply`

Apply a complete CertFix configuration from a YAML file. This command allows you to define and deploy entire infrastructure configurations declaratively.

#### Usage

```bash
certfix apply <config-file.yml> [flags]
```

#### Arguments

| Argument          | Type   | Required | Description                     |
| ----------------- | ------ | -------- | ------------------------------- |
| `config-file.yml` | string | Yes      | Path to YAML configuration file |

#### Flags

| Flag              | Type | Default | Description                                          |
| ----------------- | ---- | ------- | ---------------------------------------------------- |
| `--dry-run`       | bool | false   | Show what would be created without making changes    |
| `--skip-existing` | bool | false   | Skip resources that already exist instead of failing |

#### YAML Configuration Structure

The configuration file can contain:

- Events
- Policies
- Service Groups
- Services (with API keys and relations)

Example configuration file:

```yaml
events:
  - name: "Service Down"
    severity: "critical"
    enabled: true
  - name: "High CPU Usage"
    severity: "high"
    enabled: true

policies:
  - name: "Gradual Deployment"
    strategy: "gradual"
    enabled: true
    cron_config:
      minute: "0"
      hour: "2"
      day: "*"
      month: "*"
      weekday: "*"
  - name: "Event-Driven Policy"
    strategy: "events"
    enabled: true
    event_config:
      event_id: "1"
      total: 5

service_groups:
  - name: "Production Services"
    description: "All production-grade services"
    enabled: true
  - name: "Development Services"
    description: "Development and testing services"
    enabled: false

services:
  - name: "api-gateway"
    hash: "custom-hash-123"
    group_name: "Production Services"
    policy_name: "Gradual Deployment"
    webhook_url: "https://hooks.slack.com/services/xxx"
    keys:
      - name: "production-key"
        expiration_days: 365
      - name: "backup-key"
        expiration_days: 180
    relations:
      - target_hash: "backend-service-hash"
        type: "depends_on"
```

#### Examples

```bash
# Preview changes without applying (dry-run)
certfix apply config.yml --dry-run

# Apply configuration
certfix apply config.yml

# Apply and skip existing resources
certfix apply config.yml --skip-existing

# Apply with verbose output
certfix --verbose apply config.yml
```

#### Rollback Behavior

If an error occurs during apply, all created resources will be automatically rolled back to prevent partial configurations.

---

## Service Commands

### `certfix services`

Manage services including listing, creating, updating, activating/deactivating, and deleting services.

**Aliases:** `service`, `svc`

---

#### `certfix services list`

List all services with optional filtering.

**Aliases:** `ls`

##### Usage

```bash
certfix services list [flags]
```

##### Flags

| Flag       | Short | Type   | Default | Description                 |
| ---------- | ----- | ------ | ------- | --------------------------- |
| `--active` | `-a`  | bool   | false   | Show only active services   |
| `--group`  | `-g`  | string | -       | Filter by service group ID  |
| `--output` | `-o`  | string | table   | Output format (table, json) |

##### Examples

```bash
# List all services
certfix services list

# List only active services
certfix services list --active

# List services in a specific group
certfix services list --group 5

# List services in JSON format
certfix services list --output json

# Combine filters
certfix services list --active --group 5 --output json
```

##### Output (Table Format)

```
HASH           NAME                          GROUP                POLICY              STATUS    CREATED AT
----           ----                          -----                ------              ------    ----------
abc123...      api-gateway                   Production          Gradual Deploy      Active    2026-01-15 10:30
def456...      backend-service               Production          N/A                 Active    2026-01-15 10:31
```

---

#### `certfix services get`

Get detailed information about a specific service.

##### Usage

```bash
certfix services get <service-hash> [flags]
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |

##### Flags

| Flag       | Short | Type   | Default | Description                 |
| ---------- | ----- | ------ | ------- | --------------------------- |
| `--output` | `-o`  | string | table   | Output format (table, json) |

##### Examples

```bash
# Get service details
certfix services get abc123def456

# Get service details in JSON format
certfix services get abc123def456 --output json
```

---

#### `certfix services create`

Create a new service.

##### Usage

```bash
certfix services create [flags]
```

##### Flags

| Flag        | Short | Type   | Required | Default | Description                          |
| ----------- | ----- | ------ | -------- | ------- | ------------------------------------ |
| `--name`    | `-n`  | string | **Yes**  | -       | Name of the service                  |
| `--hash`    | -     | string | No       | (auto)  | Custom service hash (must be unique) |
| `--webhook` | `-w`  | string | No       | -       | Webhook URL for the service          |
| `--group`   | `-g`  | string | No       | -       | Service group ID                     |
| `--policy`  | `-p`  | string | No       | -       | Policy ID                            |
| `--active`  | `-a`  | bool   | No       | true    | Activate the service immediately     |

##### Examples

```bash
# Create a basic service (minimum required)
certfix services create --name "api-gateway"

# Create a service with all options
certfix services create \
  --name "api-gateway" \
  --hash "custom-hash-123" \
  --webhook "https://hooks.slack.com/services/xxx" \
  --group 5 \
  --policy 3 \
  --active true

# Create an inactive service
certfix services create --name "test-service" --active false
```

---

#### `certfix services update`

Update an existing service.

##### Usage

```bash
certfix services update <service-hash> [flags]
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |

##### Flags

| Flag              | Short | Type   | Description                        |
| ----------------- | ----- | ------ | ---------------------------------- |
| `--name`          | `-n`  | string | New name for the service           |
| `--webhook`       | `-w`  | string | New webhook URL                    |
| `--group`         | `-g`  | string | New service group ID               |
| `--policy`        | `-p`  | string | New policy ID                      |
| `--active`        | `-a`  | bool   | Activate or deactivate the service |
| `--clear-webhook` | -     | bool   | Clear the webhook URL              |
| `--clear-group`   | -     | bool   | Clear the service group            |
| `--clear-policy`  | -     | bool   | Clear the policy                   |

##### Examples

```bash
# Update service name
certfix services update abc123 --name "new-api-gateway"

# Update webhook and policy
certfix services update abc123 \
  --webhook "https://new-webhook.com" \
  --policy 5

# Clear webhook
certfix services update abc123 --clear-webhook

# Deactivate service
certfix services update abc123 --active false

# Update multiple properties
certfix services update abc123 \
  --name "updated-service" \
  --group 3 \
  --active true
```

---

#### `certfix services activate`

Activate a service.

##### Usage

```bash
certfix services activate <service-hash>
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |

##### Examples

```bash
# Activate a service
certfix services activate abc123def456
```

---

#### `certfix services deactivate`

Deactivate a service.

##### Usage

```bash
certfix services deactivate <service-hash>
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |

##### Examples

```bash
# Deactivate a service
certfix services deactivate abc123def456
```

---

#### `certfix services delete`

Delete a service.

**Aliases:** `rm`, `remove`

##### Usage

```bash
certfix services delete <service-hash> [flags]
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |

##### Flags

| Flag      | Short | Type | Default | Description                         |
| --------- | ----- | ---- | ------- | ----------------------------------- |
| `--force` | `-f`  | bool | false   | Force deletion without confirmation |

##### Examples

```bash
# Delete service (with confirmation prompt)
certfix services delete abc123def456

# Force delete without confirmation
certfix services delete abc123def456 --force
```

---

#### `certfix services generate-hash`

Generate a hash for a service name.

##### Usage

```bash
certfix services generate-hash <service-name> [flags]
```

##### Arguments

| Argument       | Type   | Required | Description         |
| -------------- | ------ | -------- | ------------------- |
| `service-name` | string | Yes      | Name of the service |

##### Flags

| Flag       | Short | Type   | Default | Description                 |
| ---------- | ----- | ------ | ------- | --------------------------- |
| `--output` | `-o`  | string | table   | Output format (table, json) |

##### Examples

```bash
# Generate hash for a service name
certfix services generate-hash "my-api-service"

# Generate hash in JSON format
certfix services generate-hash "my-api-service" --output json
```

##### Output

```
Service Name: my-api-service
Service Hash: abc123def456789
```

---

## Service Group Commands

### `certfix service-groups`

Manage service groups including listing, creating, updating, enabling/disabling, and deleting service groups.

**Aliases:** `service-group`, `svc-groups`, `svc-group`

---

#### `certfix service-groups list`

List all service groups with optional filtering.

**Aliases:** `ls`

##### Usage

```bash
certfix service-groups list [flags]
```

##### Flags

| Flag        | Short | Type   | Default | Description                      |
| ----------- | ----- | ------ | ------- | -------------------------------- |
| `--enabled` | `-e`  | bool   | false   | Show only enabled service groups |
| `--output`  | `-o`  | string | table   | Output format (table, json)      |

##### Examples

```bash
# List all service groups
certfix service-groups list

# List only enabled service groups
certfix service-groups list --enabled

# List in JSON format
certfix service-groups list --output json
```

---

#### `certfix service-groups get`

Get detailed information about a specific service group.

##### Usage

```bash
certfix service-groups get <service-group-id> [flags]
```

##### Arguments

| Argument           | Type   | Required | Description      |
| ------------------ | ------ | -------- | ---------------- |
| `service-group-id` | string | Yes      | Service group ID |

##### Flags

| Flag       | Short | Type   | Default | Description                 |
| ---------- | ----- | ------ | ------- | --------------------------- |
| `--output` | `-o`  | string | table   | Output format (table, json) |

##### Examples

```bash
# Get service group details
certfix service-groups get 5

# Get in JSON format
certfix service-groups get 5 --output json
```

---

#### `certfix service-groups create`

Create a new service group.

##### Usage

```bash
certfix service-groups create [flags]
```

##### Flags

| Flag            | Short | Type   | Required | Default | Description                          |
| --------------- | ----- | ------ | -------- | ------- | ------------------------------------ |
| `--name`        | `-n`  | string | **Yes**  | -       | Name of the service group            |
| `--description` | `-d`  | string | No       | -       | Description of the service group     |
| `--enabled`     | `-e`  | bool   | No       | true    | Enable the service group immediately |

##### Examples

```bash
# Create a basic service group
certfix service-groups create --name "Production Services"

# Create with description
certfix service-groups create \
  --name "Production Services" \
  --description "All production-grade services"

# Create disabled service group
certfix service-groups create \
  --name "Test Services" \
  --enabled false
```

---

#### `certfix service-groups update`

Update an existing service group.

##### Usage

```bash
certfix service-groups update <service-group-id> [flags]
```

##### Arguments

| Argument           | Type   | Required | Description      |
| ------------------ | ------ | -------- | ---------------- |
| `service-group-id` | string | Yes      | Service group ID |

##### Flags

| Flag            | Short | Type   | Description                         |
| --------------- | ----- | ------ | ----------------------------------- |
| `--name`        | `-n`  | string | New name for the service group      |
| `--description` | `-d`  | string | New description                     |
| `--enabled`     | `-e`  | bool   | Enable or disable the service group |

##### Examples

```bash
# Update name
certfix service-groups update 5 --name "New Name"

# Update description
certfix service-groups update 5 --description "Updated description"

# Disable service group
certfix service-groups update 5 --enabled false
```

---

#### `certfix service-groups enable`

Enable a service group.

##### Usage

```bash
certfix service-groups enable <service-group-id>
```

##### Examples

```bash
certfix service-groups enable 5
```

---

#### `certfix service-groups disable`

Disable a service group.

##### Usage

```bash
certfix service-groups disable <service-group-id>
```

##### Examples

```bash
certfix service-groups disable 5
```

---

#### `certfix service-groups delete`

Delete a service group.

**Aliases:** `rm`, `remove`

##### Usage

```bash
certfix service-groups delete <service-group-id> [flags]
```

##### Arguments

| Argument           | Type   | Required | Description      |
| ------------------ | ------ | -------- | ---------------- |
| `service-group-id` | string | Yes      | Service group ID |

##### Flags

| Flag      | Short | Type | Default | Description                         |
| --------- | ----- | ---- | ------- | ----------------------------------- |
| `--force` | `-f`  | bool | false   | Force deletion without confirmation |

##### Examples

```bash
# Delete with confirmation
certfix service-groups delete 5

# Force delete
certfix service-groups delete 5 --force
```

---

## Policy Commands

### `certfix policy`

Manage policies including listing, creating, updating, enabling/disabling, and deleting policies.

**Aliases:** `policies`, `politica`, `politicas`

---

#### `certfix policy list`

List all policies with optional filtering.

**Aliases:** `ls`

##### Usage

```bash
certfix policy list [flags]
```

##### Flags

| Flag         | Short | Type   | Default | Description                                              |
| ------------ | ----- | ------ | ------- | -------------------------------------------------------- |
| `--strategy` | `-s`  | string | -       | Filter by strategy (Gradual, Maintenance Window, Events) |
| `--enabled`  | `-e`  | bool   | false   | Show only enabled policies                               |
| `--output`   | `-o`  | string | table   | Output format (table, json)                              |

##### Examples

```bash
# List all policies
certfix policy list

# List only enabled policies
certfix policy list --enabled

# List policies by strategy
certfix policy list --strategy "Gradual"
certfix policy list --strategy "Events"

# List in JSON format
certfix policy list --output json
```

---

#### `certfix policy get`

Get detailed information about a specific policy.

##### Usage

```bash
certfix policy get <policy-id> [flags]
```

##### Arguments

| Argument    | Type   | Required | Description |
| ----------- | ------ | -------- | ----------- |
| `policy-id` | string | Yes      | Policy ID   |

##### Flags

| Flag       | Short | Type   | Default | Description                 |
| ---------- | ----- | ------ | ------- | --------------------------- |
| `--output` | `-o`  | string | table   | Output format (table, json) |

##### Examples

```bash
# Get policy details
certfix policy get 3

# Get in JSON format
certfix policy get 3 --output json
```

---

#### `certfix policy create`

Create a new policy.

##### Usage

```bash
certfix policy create [flags]
```

##### Flags

| Flag         | Type    | Required | Default | Description                                      |
| ------------ | ------- | -------- | ------- | ------------------------------------------------ |
| `--name`     | **Yes** | -        | -       | Name of the policy                               |
| `--strategy` | **Yes** | -        | -       | Strategy: Gradual, Maintenance Window, or Events |
| `--enabled`  | No      | bool     | true    | Enable the policy immediately                    |

##### Cron Configuration Flags (for Gradual and Maintenance Window strategies)

| Flag             | Type   | Default | Description              |
| ---------------- | ------ | ------- | ------------------------ |
| `--cron-minute`  | string | \*      | Cron minute (0-59 or \*) |
| `--cron-hour`    | string | \*      | Cron hour (0-23 or \*)   |
| `--cron-day`     | string | \*      | Cron day (1-31 or \*)    |
| `--cron-month`   | string | \*      | Cron month (1-12 or \*)  |
| `--cron-weekday` | string | \*      | Cron weekday (0-7 or \*) |

##### Event Configuration Flags (for Events strategy)

| Flag            | Type   | Default | Description                      |
| --------------- | ------ | ------- | -------------------------------- |
| `--event-id`    | string | -       | Event ID for Events strategy     |
| `--event-total` | int    | 1       | Total events for Events strategy |

##### Examples

```bash
# Create a Gradual policy with cron schedule
certfix policy create \
  --name "Nightly Deployment" \
  --strategy "Gradual" \
  --cron-minute "0" \
  --cron-hour "2" \
  --cron-day "*" \
  --cron-month "*" \
  --cron-weekday "*"

# Create a Maintenance Window policy
certfix policy create \
  --name "Weekend Maintenance" \
  --strategy "Maintenance Window" \
  --cron-weekday "6" \
  --cron-hour "22"

# Create an Events-based policy
certfix policy create \
  --name "Emergency Response" \
  --strategy "Events" \
  --event-id "5" \
  --event-total 3

# Create disabled policy
certfix policy create \
  --name "Test Policy" \
  --strategy "Gradual" \
  --enabled false
```

---

#### `certfix policy update`

Update an existing policy.

##### Usage

```bash
certfix policy update <policy-id> [flags]
```

##### Arguments

| Argument    | Type   | Required | Description |
| ----------- | ------ | -------- | ----------- |
| `policy-id` | string | Yes      | Policy ID   |

##### Flags

Same flags as `policy create`, but all are optional.

##### Examples

```bash
# Update policy name
certfix policy update 3 --name "New Policy Name"

# Update strategy
certfix policy update 3 --strategy "Events"

# Update cron schedule
certfix policy update 3 \
  --cron-hour "3" \
  --cron-minute "30"

# Update event configuration
certfix policy update 3 \
  --event-id "7" \
  --event-total 5

# Disable policy
certfix policy update 3 --enabled false
```

---

#### `certfix policy enable`

Enable a policy.

##### Usage

```bash
certfix policy enable <policy-id>
```

##### Examples

```bash
certfix policy enable 3
```

---

#### `certfix policy disable`

Disable a policy.

##### Usage

```bash
certfix policy disable <policy-id>
```

##### Examples

```bash
certfix policy disable 3
```

---

#### `certfix policy delete`

Delete a policy.

**Aliases:** `rm`, `remove`

##### Usage

```bash
certfix policy delete <policy-id> [flags]
```

##### Arguments

| Argument    | Type   | Required | Description |
| ----------- | ------ | -------- | ----------- |
| `policy-id` | string | Yes      | Policy ID   |

##### Flags

| Flag      | Short | Type | Default | Description                         |
| --------- | ----- | ---- | ------- | ----------------------------------- |
| `--force` | `-f`  | bool | false   | Force deletion without confirmation |

##### Examples

```bash
# Delete with confirmation
certfix policy delete 3

# Force delete
certfix policy delete 3 --force
```

---

## Event Commands

### `certfix events`

Manage events including listing, creating, updating, enabling/disabling, and deleting events.

**Aliases:** `event`, `eventos`, `evento`

---

#### `certfix events list`

List all events with optional filtering.

**Aliases:** `ls`

##### Usage

```bash
certfix events list [flags]
```

##### Flags

| Flag         | Short | Type   | Default | Description                                      |
| ------------ | ----- | ------ | ------- | ------------------------------------------------ |
| `--severity` | `-s`  | string | -       | Filter by severity (low, medium, high, critical) |
| `--enabled`  | `-e`  | bool   | false   | Show only enabled events                         |
| `--output`   | `-o`  | string | table   | Output format (table, json)                      |

##### Examples

```bash
# List all events
certfix events list

# List only enabled events
certfix events list --enabled

# List events by severity
certfix events list --severity critical
certfix events list --severity high

# List in JSON format
certfix events list --output json

# Combine filters
certfix events list --severity high --enabled
```

---

#### `certfix events get`

Get detailed information about a specific event.

##### Usage

```bash
certfix events get <event-id> [flags]
```

##### Arguments

| Argument   | Type   | Required | Description |
| ---------- | ------ | -------- | ----------- |
| `event-id` | string | Yes      | Event ID    |

##### Flags

| Flag       | Short | Type   | Default | Description                 |
| ---------- | ----- | ------ | ------- | --------------------------- |
| `--output` | `-o`  | string | table   | Output format (table, json) |

##### Examples

```bash
# Get event details
certfix events get 5

# Get in JSON format
certfix events get 5 --output json
```

---

#### `certfix events create`

Create a new event.

##### Usage

```bash
certfix events create [flags]
```

##### Flags

| Flag         | Short | Type   | Required | Default | Description                                 |
| ------------ | ----- | ------ | -------- | ------- | ------------------------------------------- |
| `--name`     | `-n`  | string | **Yes**  | -       | Name of the event                           |
| `--severity` | `-s`  | string | **Yes**  | -       | Severity level: low, medium, high, critical |
| `--enabled`  | `-e`  | bool   | No       | true    | Enable the event immediately                |

##### Examples

```bash
# Create a critical event
certfix events create \
  --name "Service Down" \
  --severity critical

# Create a high severity event
certfix events create \
  --name "High CPU Usage" \
  --severity high \
  --enabled true

# Create a disabled event
certfix events create \
  --name "Test Event" \
  --severity low \
  --enabled false
```

---

#### `certfix events update`

Update an existing event.

##### Usage

```bash
certfix events update <event-id> [flags]
```

##### Arguments

| Argument   | Type   | Required | Description |
| ---------- | ------ | -------- | ----------- |
| `event-id` | string | Yes      | Event ID    |

##### Flags

| Flag         | Short | Type   | Description                                     |
| ------------ | ----- | ------ | ----------------------------------------------- |
| `--name`     | `-n`  | string | New name for the event                          |
| `--severity` | `-s`  | string | New severity level: low, medium, high, critical |
| `--enabled`  | `-e`  | bool   | Enable or disable the event                     |

##### Examples

```bash
# Update event name
certfix events update 5 --name "Updated Event Name"

# Update severity
certfix events update 5 --severity critical

# Disable event
certfix events update 5 --enabled false

# Update multiple properties
certfix events update 5 \
  --name "New Name" \
  --severity high \
  --enabled true
```

---

#### `certfix events enable`

Enable an event.

##### Usage

```bash
certfix events enable <event-id>
```

##### Examples

```bash
certfix events enable 5
```

---

#### `certfix events disable`

Disable an event.

##### Usage

```bash
certfix events disable <event-id>
```

##### Examples

```bash
certfix events disable 5
```

---

#### `certfix events delete`

Delete an event.

**Aliases:** `rm`, `remove`

##### Usage

```bash
certfix events delete <event-id> [flags]
```

##### Arguments

| Argument   | Type   | Required | Description |
| ---------- | ------ | -------- | ----------- |
| `event-id` | string | Yes      | Event ID    |

##### Flags

| Flag      | Short | Type | Default | Description                         |
| --------- | ----- | ---- | ------- | ----------------------------------- |
| `--force` | `-f`  | bool | false   | Force deletion without confirmation |

##### Examples

```bash
# Delete with confirmation
certfix events delete 5

# Force delete
certfix events delete 5 --force
```

---

## API Key Commands

### `certfix keys`

Manage service API keys including listing, creating, enabling/disabling, and deleting keys.

**Aliases:** `key`

---

#### `certfix keys list`

List all API keys for a service.

**Aliases:** `ls`

##### Usage

```bash
certfix keys list <service-hash> [flags]
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |

##### Flags

| Flag       | Short | Type   | Default | Description                 |
| ---------- | ----- | ------ | ------- | --------------------------- |
| `--output` | `-o`  | string | table   | Output format (table, json) |

##### Examples

```bash
# List all API keys for a service
certfix keys list abc123def456

# List in JSON format
certfix keys list abc123def456 --output json
```

##### Output (Table Format)

```
KEY ID         KEY NAME            API KEY              STATUS    EXPIRATION  CREATED AT
------         --------            -------              ------    ----------  ----------
key123...      production-key      sk_prod_abc123...   Enabled   2027-01-15  2026-01-15 10:30
key456...      backup-key          sk_prod_def456...   Disabled  2026-06-15  2026-01-15 10:31
```

---

#### `certfix keys get`

Get complete API keys data for a service.

##### Usage

```bash
certfix keys get <service-hash> [flags]
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |

##### Flags

| Flag       | Short | Type   | Default | Description                 |
| ---------- | ----- | ------ | ------- | --------------------------- |
| `--output` | `-o`  | string | table   | Output format (table, json) |

##### Examples

```bash
# Get API keys data
certfix keys get abc123def456

# Get in JSON format
certfix keys get abc123def456 --output json
```

---

#### `certfix keys add`

Add a new API key to a service.

##### Usage

```bash
certfix keys add <service-hash> [flags]
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |

##### Flags

| Flag           | Short | Type   | Required | Default | Description               |
| -------------- | ----- | ------ | -------- | ------- | ------------------------- |
| `--name`       | `-n`  | string | **Yes**  | -       | Name of the API key       |
| `--expiration` | `-e`  | int    | **Yes**  | 365     | Expiration period in days |

##### Examples

```bash
# Add a new API key (1 year expiration)
certfix keys add abc123def456 \
  --name "production-key" \
  --expiration 365

# Add a short-lived API key (90 days)
certfix keys add abc123def456 \
  --name "temporary-key" \
  --expiration 90

# Add a long-lived API key (2 years)
certfix keys add abc123def456 \
  --name "backup-key" \
  --expiration 730
```

---

#### `certfix keys toggle`

Toggle an API key's enable/disable status.

##### Usage

```bash
certfix keys toggle <service-hash> <key-id>
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |
| `key-id`       | string | Yes      | API key ID              |

##### Examples

```bash
# Toggle API key status
certfix keys toggle abc123def456 key789
```

---

#### `certfix keys enable`

Enable an API key.

##### Usage

```bash
certfix keys enable <service-hash> <key-id>
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |
| `key-id`       | string | Yes      | API key ID              |

##### Examples

```bash
# Enable an API key
certfix keys enable abc123def456 key789
```

---

#### `certfix keys disable`

Disable an API key.

##### Usage

```bash
certfix keys disable <service-hash> <key-id>
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |
| `key-id`       | string | Yes      | API key ID              |

##### Examples

```bash
# Disable an API key
certfix keys disable abc123def456 key789
```

---

#### `certfix keys delete`

Delete an API key.

##### Usage

```bash
certfix keys delete <service-hash> <key-id> [flags]
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |
| `key-id`       | string | Yes      | API key ID              |

##### Flags

| Flag      | Short | Type | Default | Description                         |
| --------- | ----- | ---- | ------- | ----------------------------------- |
| `--force` | `-f`  | bool | false   | Force deletion without confirmation |

##### Examples

```bash
# Delete with confirmation
certfix keys delete abc123def456 key789

# Force delete
certfix keys delete abc123def456 key789 --force
```

---

## Service Matrix Commands

### `certfix matrix`

Manage service matrix (service relations) including listing, creating, enabling/disabling, and deleting service relations.

**Aliases:** `matriz`

---

#### `certfix matrix list`

List all relations for a service.

**Aliases:** `ls`

##### Usage

```bash
certfix matrix list <service-hash> [flags]
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |

##### Flags

| Flag       | Short | Type   | Default | Description                 |
| ---------- | ----- | ------ | ------- | --------------------------- |
| `--output` | `-o`  | string | table   | Output format (table, json) |

##### Examples

```bash
# List all service relations
certfix matrix list abc123def456

# List in JSON format
certfix matrix list abc123def456 --output json
```

##### Output (Table Format)

```
RELATION ID    SOURCE SERVICE       RELATED SERVICE      STATUS    CREATED AT
-----------    --------------       ---------------      ------    ----------
rel123...      api-gateway          backend-service      Enabled   2026-01-15 10:30
rel456...      api-gateway          database-service     Enabled   2026-01-15 10:31
```

---

#### `certfix matrix get`

Get complete matrix data for a service including all available services.

##### Usage

```bash
certfix matrix get <service-hash> [flags]
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |

##### Flags

| Flag       | Short | Type   | Default | Description                 |
| ---------- | ----- | ------ | ------- | --------------------------- |
| `--output` | `-o`  | string | table   | Output format (table, json) |

##### Examples

```bash
# Get matrix data
certfix matrix get abc123def456

# Get in JSON format
certfix matrix get abc123def456 --output json
```

---

#### `certfix matrix add`

Add a service relation.

##### Usage

```bash
certfix matrix add <service-hash> [flags]
```

##### Arguments

| Argument       | Type   | Required | Description                    |
| -------------- | ------ | -------- | ------------------------------ |
| `service-hash` | string | Yes      | Source service hash identifier |

##### Flags

Specific flags depend on implementation (check code for details).

##### Examples

```bash
# Add a service relation
certfix matrix add abc123def456 --target xyz789 --type depends_on
```

---

#### `certfix matrix toggle`

Toggle a service relation's enable/disable status.

##### Usage

```bash
certfix matrix toggle <service-hash> <relation-id>
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |
| `relation-id`  | string | Yes      | Relation ID             |

##### Examples

```bash
# Toggle relation status
certfix matrix toggle abc123def456 rel789
```

---

#### `certfix matrix enable`

Enable a service relation.

##### Usage

```bash
certfix matrix enable <service-hash> <relation-id>
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |
| `relation-id`  | string | Yes      | Relation ID             |

##### Examples

```bash
# Enable a service relation
certfix matrix enable abc123def456 rel789
```

---

#### `certfix matrix disable`

Disable a service relation.

##### Usage

```bash
certfix matrix disable <service-hash> <relation-id>
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |
| `relation-id`  | string | Yes      | Relation ID             |

##### Examples

```bash
# Disable a service relation
certfix matrix disable abc123def456 rel789
```

---

#### `certfix matrix delete`

Delete a service relation.

##### Usage

```bash
certfix matrix delete <service-hash> <relation-id> [flags]
```

##### Arguments

| Argument       | Type   | Required | Description             |
| -------------- | ------ | -------- | ----------------------- |
| `service-hash` | string | Yes      | Service hash identifier |
| `relation-id`  | string | Yes      | Relation ID             |

##### Flags

| Flag      | Short | Type | Default | Description                         |
| --------- | ----- | ---- | ------- | ----------------------------------- |
| `--force` | `-f`  | bool | false   | Force deletion without confirmation |

##### Examples

```bash
# Delete with confirmation
certfix matrix delete abc123def456 rel789

# Force delete
certfix matrix delete abc123def456 rel789 --force
```

---

## Configuration Files

### Config File Location

Default: `~/.certfix/config.yaml`

### Configuration Options

| Key              | Description               | Default                  |
| ---------------- | ------------------------- | ------------------------ |
| `endpoint`       | API endpoint URL          | `https://api.certfix.io` |
| `timeout`        | Request timeout (seconds) | `30`                     |
| `retry_attempts` | Number of retry attempts  | `3`                      |

### Environment Variables

You can use environment variables with the `CERTFIX_` prefix:

```bash
export CERTFIX_ENDPOINT=https://api.certfix.io
export CERTFIX_TIMEOUT=60
export CERTFIX_RETRY_ATTEMPTS=5
```

### Token Storage

Authentication tokens are stored in `~/.certfix/token.json` with restricted permissions (0600).

Example token file:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-02-17T10:30:00Z"
}
```

---

## Architecture

Certfix CLI follows clean architecture principles:

```
certfix-cli/
├── cmd/certfix/           # Command definitions (Cobra)
│   ├── root.go           # Root command
│   ├── configure.go      # Configuration command
│   ├── login.go          # Authentication commands
│   ├── services.go       # Service management
│   ├── service_groups.go # Service group management
│   ├── policy.go         # Policy management
│   ├── eventos.go        # Event management
│   ├── keys.go           # API key management
│   ├── matrix.go         # Service matrix management
│   └── apply.go          # Configuration apply
├── internal/              # Private application code
│   ├── auth/             # Authentication (JWT)
│   ├── config/           # Configuration (Viper)
│   └── api/              # API client
├── pkg/                   # Public libraries
│   ├── client/           # HTTP client
│   ├── logger/           # Logging (Logrus)
│   └── models/           # Data models
└── main.go               # Entry point
```

---

## Technology Stack

- **Go**: Primary programming language (1.21+)
- **Cobra**: CLI command structure and parsing
- **Viper**: Configuration management
- **JWT**: Token-based authentication
- **Logrus**: Structured logging
- **YAML**: Configuration file format

---

## Security

- Authentication tokens are stored securely with restricted permissions (0600)
- Configuration directory `~/.certfix/` is created with restricted permissions (0700)
- All API communications use HTTPS
- Passwords are never logged or stored locally
- Personal access tokens are recommended for authentication

---

## Troubleshooting

### Authentication Issues

```bash
# Enable verbose logging
certfix --verbose login

# Re-authenticate
certfix logout
certfix login
```

### Configuration Issues

```bash
# View current configuration
certfix configure --show

# Reset configuration
rm ~/.certfix/config.yaml
certfix configure
```

### API Endpoint Not Configured

If you see: `⚠ No API endpoint configured`

**Solution:**

```bash
certfix configure --api-url https://api.certfix.io
```

### Connection Errors

```bash
# Check configuration
certfix configure --show

# Test with verbose output
certfix --verbose services list

# Verify API endpoint is accessible
curl https://api.certfix.io/health
```

---

## License

This project is proprietary software. All rights reserved.

## Support

For support, please contact:

- Email: development@certfix.io
- Documentation: https://docs.certfix.io
- Issues: https://github.com/certfix/certfix-cli/issues
