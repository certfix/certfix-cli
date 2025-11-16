# Certfix CLI

A cross-platform command-line interface tool for managing certificates, application configurations, and infrastructure operations. Built with Go and designed to work seamlessly on Linux, macOS, and Windows.

## Features

- 🔐 **Authentication**: Secure JWT-based authentication system
- ⚙️ **Configuration Management**: Flexible configuration with Viper
- 📜 **Certificate Operations**: Create, renew, revoke, and manage SSL/TLS certificates
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

Set up your Certfix CLI with the API endpoint and preferences:

**Interactive mode:**

```bash
certfix configure
```

This will prompt you for:

- API URL (default: https://api.certfix.io)
- Timeout in seconds (default: 30)
- Retry attempts (default: 3)

**Non-interactive mode:**

```bash
certfix configure --api-url https://api.certfix.io --timeout 60 --retry-attempts 3
```

**View current configuration:**

```bash
certfix configure
```

### 2. Login

Authenticate with your Certfix account:

**Interactive mode (recommended):**

```bash
certfix login
```

This will securely prompt you for:

- Username
- Password (hidden input)

**Non-interactive mode:**

```bash
certfix login --username your-email@example.com --password your-password
```

### 3. Manage Certificates

Handle SSL/TLS certificates:

```bash
# Create a certificate
certfix cert create example.com

# List all certificates
certfix cert list

# Revoke a certificate
certfix cert revoke <cert-id>
```

## Usage

### Configuration Commands

```bash
# Interactive configuration (recommended)
certfix configure

# Configure with flags
certfix configure --api-url <url> [--timeout <seconds>] [--retry-attempts <count>]

# Examples
certfix configure --api-url https://api.certfix.io
certfix configure --api-url https://staging.certfix.io --timeout 90 --retry-attempts 5
```

### Authentication Commands

```bash
# Login (interactive - recommended)
certfix login

# Login (non-interactive)
certfix login --username <username> --password <password>

# Logout
certfix logout
```

### Certificate Commands

```bash
# Create a certificate
certfix cert create <domain>

# List certificates
certfix cert list

# Revoke a certificate
certfix cert revoke <id>
```

## Global Flags

- `--config <path>`: Specify a custom config file (default: `~/.certfix/config.yaml`)
- `--verbose, -v`: Enable verbose output for debugging

## Configuration

Configuration is stored in `~/.certfix/config.yaml` by default.

### Configuration Options

| Key              | Description               | Default                  |
| ---------------- | ------------------------- | ------------------------ |
| `endpoint`       | API endpoint URL          | `https://api.certfix.io` |
| `timeout`        | Request timeout (seconds) | `30`                     |
| `retry_attempts` | Number of retry attempts  | `3`                      |

### Environment Variables

You can also use environment variables with the `CERTFIX_` prefix:

```bash
export CERTFIX_ENDPOINT=https://api.certfix.io
export CERTFIX_TIMEOUT=60
```

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

## Testing

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Security

- Authentication tokens are stored securely in `~/.certfix/token.json` with restricted permissions (0600)
- Configuration directory `~/.certfix/` is created with restricted permissions (0700)
- All API communications use HTTPS
- Passwords are never logged or stored locally

## Troubleshooting

### Authentication Issues

```bash
# Enable verbose logging
certfix --verbose login --username user --password pass

# Re-authenticate
certfix logout
certfix login --username user --password pass
```

### Configuration Issues

```bash
# Check current configuration
certfix config list

# Reset configuration
rm ~/.certfix/config.yaml
certfix config set endpoint https://api.certfix.io
```

## License

This project is proprietary software. All rights reserved.

## Support

For support, please contact:

- Email: support@certfix.io
- Documentation: https://docs.certfix.io
- Issues: https://github.com/certfix/certfix-cli/issues
