# Certfix CLI

A cross-platform command-line interface tool for managing services, policies, events, and service configurations. Built with Go and designed to work seamlessly on Linux, macOS, and Windows.

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
Download certfix-windows-amd64.exe and add to PATH
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

Set up your Certfix CLI with the API endpoint:

```bash
certfix configure
```

This will interactively prompt you for:

- API URL (e.g., https://api.certfix.io)
- Timeout in seconds (default: 30)
- Retry attempts (default: 3)

Or configure non-interactively:

```bash
certfix configure --api-url https://api.certfix.io
```

### 2. Login

Authenticate with your email and personal access token:

```bash
certfix login
```

Or provide credentials directly:

```bash
certfix login --email your-email@example.com --token your-personal-access-token
```

### 3. Start Managing Resources

Now you can manage your infrastructure:

```bash
# List all services
certfix services list

# List all policies
certfix policy list

# List all events
certfix events list

# Create a new service
certfix services create --name "my-api-service"

# Apply configuration from YAML
certfix apply config.yml
```

## Documentation

For complete command reference, usage examples, and detailed documentation, see:

📚 **[Complete CLI Reference](DOCS/CLI_REFERENCE.md)**

The CLI Reference includes:

- All available commands and subcommands
- Detailed flag descriptions
- Usage examples for every command
- Configuration file format
- YAML configuration structure for `apply` command
- Troubleshooting guide

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
certfix configure --show

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
