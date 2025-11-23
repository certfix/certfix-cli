# Certfix CLI

A cross-platform command-line interface tool for managing SSL/TLS certificates and Certificate Authority operations. Built with Go and designed to work seamlessly on Linux, macOS, and Windows.

## Features

- 🔐 **Authentication**: Secure JWT-based authentication with support for personal access tokens
- ⚙️ **Configuration Management**: Flexible configuration with Viper
- 📜 **Certificate Operations**: Create, list, and revoke SSL/TLS certificates (server and client)
- 🔄 **Synchronization**: Sync certificates with the Certificate Authority
- 💾 **Backup**: Create CA backups
- 📝 **Structured Logging**: Comprehensive logging with Logrus
- 🌍 **Cross-Platform**: Compiled binaries for Linux, macOS, and Windows

## Installation

### From Release Binary

Download the appropriate binary for your platform from the releases page:

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

# Windows
# Download certfix-windows-amd64.exe and add to PATH
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/certfix/certfix-cli.git
cd certfix-cli

# Install dependencies
make deps

# Build for your current platform
make build

# Or build for all platforms
make build-all
```

## Quick Start

### 1. Configure API Endpoint

Set up your Certfix CLI with the API endpoint and API token:

**Interactive mode:**

```bash
certfix configure
```

This will prompt you for:

- API URL (default: https://api.certfix.io)
- API Token (for accessing the backoffice API)
- Timeout in seconds (default: 30)
- Retry attempts (default: 3)

**Non-interactive mode:**

```bash
certfix configure set endpoint https://api.certfix.io
certfix configure set api_token your-secure-api-token-here
certfix configure set timeout 60
certfix configure set retry_attempts 3
```

**View current configuration:**

```bash
certfix configure
```

### 2. Login

Authenticate with your Certfix account using either username/password or personal access tokens.

> **💡 Recommended**: Use [personal access tokens](PERSONAL_TOKEN_AUTH.md) for better security and easier token management.

**Interactive mode (recommended):**

```bash
certfix login
```

This will prompt you to choose:

1. Username and Password (traditional)
2. Personal Access Token (recommended)

**Personal access token (command-line):**

```bash
certfix login --email your-email@example.com --token pat_abc123...
```

**Traditional username/password:**

```bash
certfix login --username your-email@example.com --password your-password
```

For detailed information about personal access tokens, see [PERSONAL_TOKEN_AUTH.md](PERSONAL_TOKEN_AUTH.md).

### 3. Create Certificates

Create server or client certificates:

```bash
# Create a server certificate
certfix cert create webapp-prod --type server --days 365

# Create a client certificate
certfix cert create client-app --type client --days 90
```

### 4. Manage Certificates

List and manage your certificates:

```bash
# List valid certificates
certfix cert list valid

# List revoked certificates
certfix cert list revoked

# List expiring certificates (next 30 days)
certfix cert list expiring 30
```

### 5. Synchronize & Backup

```bash
# Sync certificates with CA
certfix sync

# Create a backup
certfix backup
```

## Usage

### Configuration Commands

#### `certfix configure`

Configure the CLI with API endpoint and settings.

**Interactive mode:**

```bash
certfix configure
```

**Non-interactive mode:**

```bash
certfix configure [flags]
```

**Flags:**

- `--api-url, -a <url>` - API endpoint URL (e.g., https://api.certfix.io)
- `--timeout, -t <seconds>` - Request timeout in seconds
- `--retry-attempts, -r <count>` - Number of retry attempts for failed requests

**Examples:**

```bash
# Configure with all options
certfix configure --api-url https://api.certfix.io --timeout 60 --retry-attempts 5

# Configure only API URL
certfix configure -a https://staging.certfix.io

# View current configuration (run without flags)
certfix configure
```

---

### Authentication Commands

#### `certfix login`

Authenticate with the Certfix API using username/password or personal access tokens.

> **💡 Recommended**: Use personal access tokens for CLI authentication. See [PERSONAL_TOKEN_AUTH.md](PERSONAL_TOKEN_AUTH.md) for setup instructions.

**Interactive mode (recommended):**

```bash
certfix login
```

This prompts you to choose:

1. Username and Password (traditional)
2. Personal Access Token (recommended for CLI)

**Non-interactive mode with personal token:**

```bash
certfix login --email admin@example.com --token pat_abc123xyz...
```

**Non-interactive mode with username/password:**

```bash
certfix login --username admin@example.com --password mypassword
```

**Flags:**

- `--email, -e <email>` - Email for personal token authentication
- `--token, -t <token>` - Personal access token for authentication
- `--username, -u <username>` - Username for password authentication
- `--password, -p <password>` - Password for authentication

**Examples:**

```bash
# Interactive login (choose auth method)
certfix login

# Login with personal access token
certfix login -e admin@example.com -t pat_1234567890abcdef

# Traditional login with username/password
certfix login -u admin@example.com -p mypassword
```

**Note:** Requires API endpoint and API token to be configured first via `certfix configure`.

---

#### `certfix logout`

Remove stored authentication token.

```bash
certfix logout
```

---

### Certificate Commands

#### `certfix cert create`

Create a new SSL/TLS certificate (server or client).

```bash
certfix cert create <common-name> [flags]
```

**Required Flags:**

- `--type, -t <type>` - Certificate type: `server` or `client`

**Optional Flags:**

- `--description, -d <text>` - Certificate description
- `--days <number>` - Validity period in days
- `--key-size, -k <bits>` - RSA key size in bits (e.g., 2048, 4096)
- `--san, -s <names>` - Subject Alternative Names (format: `DNS:example.com,IP:192.168.1.1`)

**Examples:**

```bash
# Create server certificate (minimal)
certfix cert create webapp-prod --type server

# Create server certificate with all options
certfix cert create webapp-prod \
  --type server \
  --description "Production Web Server" \
  --days 365 \
  --key-size 2048 \
  --san "DNS:webapp.local,DNS:webapp.example.com,IP:192.168.1.100"

# Create client certificate
certfix cert create mobile-app-client \
  --type client \
  --description "Mobile app client certificate" \
  --days 90
```

**Output:**

```
✓ Certificate created successfully
Unique ID:     20251116-225759-c28b
Serial Number: 1001
App Name:      webapp-prod
```

---

#### `certfix cert list`

List certificates by status.

```bash
certfix cert list <type> [days]
```

**Types:**

- `valid` - List all valid certificates
- `revoked` - List all revoked certificates
- `expiring <days>` - List certificates expiring in specified days

**Examples:**

```bash
# List valid certificates
certfix cert list valid

# List revoked certificates
certfix cert list revoked

# List certificates expiring in next 30 days
certfix cert list expiring 30

# List certificates expiring in next 7 days
certfix cert list expiring 7
```

**Output (JSON):**

```json
[
  {
    "app_name": "webapp-prod",
    "unique_id": "20251116-225759-c28b",
    "client_id": "N/A",
    "certificate_type": "server",
    "expiration_date": "2026-11-16 22:57:59",
    "status": "valid",
    "revocation_date": null
  }
]
```

---

#### `certfix cert revoke`

Revoke a certificate or all certificates.

```bash
certfix cert revoke <unique-id|all> [flags]
```

**Flags:**

- `--cascade, -c` - Cascade revocation (default: true)
- `--reason, -r <reason>` - Revocation reason (default: superseded)

**Examples:**

```bash
# Revoke specific certificate (default options)
certfix cert revoke 20251116-225759-c28b

# Revoke with custom options
certfix cert revoke 20251116-225759-c28b --cascade=false --reason="keyCompromise"

# Revoke all certificates
certfix cert revoke all

# Revoke all with custom reason
certfix cert revoke all --reason="unspecified"
```

**Output:**

```
✓ Certificate '20251116-225759-c28b' revoked successfully
```

---

### Utility Commands

#### `certfix sync`

Synchronize certificates with the Certificate Authority.

```bash
certfix sync
```

**Output:**

```
✓ Synchronization successful
Synced: 3 certificates
```

---

#### `certfix backup`

Create a backup of the Certificate Authority.

```bash
certfix backup
```

**Output:**

```
Backup status: success
```

---

#### `certfix version`

Display the CLI version.

```bash
certfix version
```

**Output:**

```
Certfix CLI v1.0.0
```

## Global Flags

Available for all commands:

- `--config <path>` - Specify a custom config file (default: `~/.certfix/config.yaml`)
- `--verbose, -v` - Enable verbose output for debugging

**Example:**

```bash
certfix --verbose cert list valid
certfix --config /custom/path/config.yaml login
```

---

## Configuration Files

### Config File Location

Default: `~/.certfix/config.yaml`

### Configuration Options

| Key              | Description               | Default                  |
| ---------------- | ------------------------- | ------------------------ |
| `endpoint`       | API endpoint URL          | `https://api.certfix.io` |
| `api_token`      | API token for requests    | (none)                   |
| `timeout`        | Request timeout (seconds) | `30`                     |
| `retry_attempts` | Number of retry attempts  | `3`                      |

### Environment Variables

You can also use environment variables with the `CERTFIX_` prefix:

```bash
export CERTFIX_ENDPOINT=https://api.certfix.io
export CERTFIX_API_TOKEN=your-secure-api-token-here
export CERTFIX_TIMEOUT=60
```

### Token Storage

Authentication tokens are stored in `~/.certfix/token.json` with restricted permissions (0600).

**Example token file:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-11-17T10:30:00Z"
}
```

---

## Command Reference

### Complete Command Tree

```
certfix
├── configure              # Configure API settings
│   ├── --api-url, -a     # API endpoint URL
│   ├── --timeout, -t     # Request timeout
│   └── --retry-attempts, -r  # Retry attempts
├── login                  # Authenticate
│   ├── --email, -e       # Email for token auth
│   ├── --token, -t       # Personal access token
│   ├── --username, -u    # Username for password auth
│   └── --password, -p    # Password
├── logout                 # Remove auth token
├── backup                 # Create CA backup
├── sync                   # Sync certificates
├── cert                   # Certificate management
│   ├── create <name>     # Create certificate
│   │   ├── --type, -t    # server or client (required)
│   │   ├── --description, -d  # Description
│   │   ├── --days        # Validity days
│   │   ├── --key-size, -k  # RSA key size
│   │   └── --san, -s     # Subject Alternative Names
│   ├── list <type>       # List certificates
│   │   ├── valid         # List valid certs
│   │   ├── revoked       # List revoked certs
│   │   └── expiring <days>  # List expiring certs
│   └── revoke <id|all>   # Revoke certificate
│       ├── --cascade, -c # Cascade revocation
│       └── --reason, -r  # Revocation reason
└── version                # Show version
```

---

## Architecture

Certfix CLI follows clean architecture principles:

```
certfix-cli/
├── cmd/certfix/           # Command definitions (Cobra)
├── internal/              # Private application code
│   ├── auth/              # Authentication (JWT)
│   ├── config/            # Configuration (Viper)
│   └── api/               # API client
├── pkg/                   # Public libraries
│   ├── client/            # HTTP client
│   ├── logger/            # Logging (Logrus)
│   └── models/            # Data models
└── main.go                # Entry point
```

## Technology Stack

- **Go**: Primary programming language
- **Cobra**: CLI command structure and parsing
- **Viper**: Configuration management
- **JWT**: Token-based authentication
- **Logrus**: Structured logging

## Development

### Prerequisites

- Go 1.21 or higher
- Make (optional, for using Makefile)

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run linter
make lint

# Clean build artifacts
make clean
```

### Running in Development

```bash
# Run without building
make dev

# Or run directly with go
go run main.go configure
go run main.go version

# Test the binary
./bin/certfix configure
./bin/certfix version
```

### Testing Locally

```bash
# Build and test
make build
./bin/certfix configure --api-url https://localhost:8080

# Or use go run for quick testing
go run main.go configure

# Check configuration
cat ~/.certfix/config.yaml
```

### Testing in Docker

```bash
# Build for Linux
make build-linux

# Test in Ubuntu container
docker run -it --rm \
  -v $(pwd)/dist/certfix-linux-amd64:/usr/local/bin/certfix \
  ubuntu:latest certfix configure --api-url https://api.example.com

# Interactive testing
docker run -it --rm \
  -v $(pwd)/dist/certfix-linux-amd64:/usr/local/bin/certfix \
  ubuntu:latest /bin/bash
```

---

## Security

- Authentication tokens are stored securely in `~/.certfix/token.json` with restricted permissions (0600)
- Configuration directory `~/.certfix/` is created with restricted permissions (0700)
- All API communications use HTTPS
- Passwords are never logged or stored locally

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
certfix configure

# Reset configuration
rm ~/.certfix/config.yaml
certfix configure --api-url https://api.certfix.io
```

### API Endpoint Not Configured

If you see this error when trying to login:

```
⚠ No API endpoint configured.
Please run 'certfix configure' first to set up your API endpoint.
```

**Solution:**

```bash
certfix configure --api-url https://api.certfix.io
```

---

## License

This project is proprietary software. All rights reserved.

## Support

For support, please contact:

- Email: development@certfix.io
- Documentation: https://docs.certfix.io
- Issues: https://github.com/certfix/certfix-cli/issues
