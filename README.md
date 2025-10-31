# Certfix CLI

A cross-platform command-line interface tool for managing certificates, application configurations, and infrastructure operations. Built with Go and designed to work seamlessly on Linux, macOS, and Windows.

## Features

- 🔐 **Authentication**: Secure JWT-based authentication system
- ⚙️ **Configuration Management**: Flexible configuration with Viper
- 🖥️ **Instance Management**: Create, list, and delete instances
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

### 1. Login

Authenticate with your Certfix account:

```bash
certfix login --username your-email@example.com --password your-password
```

### 2. Configure (Optional)

Set custom configuration options:

```bash
certfix config set endpoint https://api.certfix.io
certfix config list
```

### 3. Manage Instances

Create and manage instances:

```bash
# Create an instance
certfix instance create my-instance --type standard --region us-east-1

# List all instances
certfix instance list

# Delete an instance
certfix instance delete <instance-id>
```

### 4. Manage Certificates

Handle SSL/TLS certificates:

```bash
# Create a certificate
certfix cert create example.com

# List all certificates
certfix cert list

# Renew a certificate
certfix cert renew <cert-id>

# Revoke a certificate
certfix cert revoke <cert-id>
```

## Usage

### Authentication Commands

```bash
# Login
certfix login --username <username> --password <password> [--endpoint <url>]

# Logout
certfix logout
```

### Configuration Commands

```bash
# Set a configuration value
certfix config set <key> <value>

# Get a configuration value
certfix config get <key>

# List all configurations
certfix config list
```

### Instance Commands

```bash
# Create an instance
certfix instance create <name> [--type <type>] [--region <region>]

# List instances
certfix instance list

# Delete an instance
certfix instance delete <id>
```

### Certificate Commands

```bash
# Create a certificate
certfix cert create <domain>

# List certificates
certfix cert list

# Renew a certificate
certfix cert renew <id>

# Revoke a certificate
certfix cert revoke <id>
```

## Global Flags

- `--config <path>`: Specify a custom config file (default: `~/.certfix/config.yaml`)
- `--verbose, -v`: Enable verbose output for debugging

## Configuration

Configuration is stored in `~/.certfix/config.yaml` by default.

### Configuration Options

| Key | Description | Default |
|-----|-------------|---------|
| `endpoint` | API endpoint URL | `https://api.certfix.io` |
| `timeout` | Request timeout (seconds) | `30` |
| `retry_attempts` | Number of retry attempts | `3` |

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

# Or
go run main.go
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