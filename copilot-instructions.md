# Certfix CLI - Copilot Instructions

## Project Overview

Certfix CLI is a cross-platform command-line interface tool for managing certificates, application configurations, and infrastructure operations. It's built with Go and follows patterns similar to AWS CLI and Azure CLI, requiring authentication before performing operations.

## Architecture

### High-Level Architecture

```
certfix-cli/
├── main.go                 # Application entry point
├── cmd/certfix/           # Command definitions (Cobra)
│   ├── root.go            # Root command and global flags
│   ├── login.go           # Authentication command
│   ├── logout.go          # Logout command
│   ├── config.go          # Configuration management
│   ├── instance.go        # Instance management
│   └── cert.go            # Certificate management
├── internal/              # Private application code
│   ├── auth/              # Authentication logic (JWT)
│   ├── config/            # Configuration management (Viper)
│   └── api/               # API client logic
├── pkg/                   # Public libraries
│   ├── client/            # HTTP client
│   ├── logger/            # Logging (Logrus)
│   └── models/            # Data models
└── Makefile               # Build automation
```

### Technology Stack

- **Language**: Go (Golang)
- **CLI Framework**: Cobra - Command structure and argument parsing
- **Configuration**: Viper - Configuration file and environment variable management
- **Authentication**: JWT (golang-jwt/jwt) - Token-based authentication
- **Logging**: Logrus - Structured logging
- **HTTP Client**: Native net/http - API communication

## Key Components

### 1. Command Structure (cmd/certfix/)

All CLI commands are organized under the `cmd/certfix/` directory:

- **root.go**: Base command with global flags (--config, --verbose)
- **login.go**: Authenticate with username/password, stores JWT token
- **logout.go**: Removes stored authentication token
- **config.go**: Manage configuration (set, get, list)
- **instance.go**: Manage instances (create, list, delete)
- **cert.go**: Manage certificates (create, list, renew, revoke)

### 2. Authentication Flow (internal/auth/)

1. User runs `certfix login --username <user> --password <pass>`
2. CLI sends credentials to API endpoint
3. API returns JWT token
4. Token is stored in `~/.certfix/token.json` with expiration
5. Subsequent commands use stored token for authentication
6. Token is validated for expiration before each API call

### 3. Configuration Management (internal/config/)

- Uses Viper for flexible configuration management
- Configuration file: `~/.certfix/config.yaml`
- Supports environment variables with CERTFIX_ prefix
- Default values for endpoint, timeout, retry attempts
- Can be overridden with --config flag

### 4. API Client (internal/api/ and pkg/client/)

- **pkg/client/**: Generic HTTP client with retry logic
- **internal/api/**: Business logic for specific API operations
- Automatically adds authentication headers
- Handles response parsing and error handling

### 5. Logging (pkg/logger/)

- Uses Logrus for structured logging
- Configurable log levels (info, debug)
- Enabled with --verbose flag
- Logs to stdout with timestamps

## Development Guidelines

### Adding a New Command

1. Create a new file in `cmd/certfix/` (e.g., `newcommand.go`)
2. Define the command using Cobra:
   ```go
   var newCmd = &cobra.Command{
       Use:   "new",
       Short: "Short description",
       Long:  "Long description",
       RunE: func(cmd *cobra.Command, args []string) error {
           // Implementation
           return nil
       },
   }
   ```
3. Register command in `init()` function:
   ```go
   func init() {
       rootCmd.AddCommand(newCmd)
   }
   ```
4. Check authentication if needed:
   ```go
   if !auth.IsAuthenticated() {
       return fmt.Errorf("not authenticated")
   }
   ```

### Adding a New API Endpoint

1. Add method to `internal/api/client.go`:
   ```go
   func (c *Client) NewOperation() (*models.Result, error) {
       token, err := auth.GetToken()
       if err != nil {
           return nil, err
       }
       
       response, err := c.httpClient.PostWithAuth("/endpoint", payload, token)
       if err != nil {
           return nil, err
       }
       
       // Parse and return result
       return result, nil
   }
   ```

2. Add corresponding model in `pkg/models/models.go` if needed

### Configuration Keys

Common configuration keys:
- `endpoint`: API base URL (default: https://api.certfix.io)
- `timeout`: Request timeout in seconds (default: 30)
- `retry_attempts`: Number of retry attempts (default: 3)

### Error Handling

- Use wrapped errors with fmt.Errorf and %w
- Log errors with appropriate levels
- Return user-friendly error messages
- Check authentication before API calls

## Building and Distribution

### Local Development Build

```bash
make build
# Output: bin/certfix
```

### Cross-Platform Builds

```bash
make build-all
# Outputs all platform binaries to dist/
```

### Supported Platforms

- Linux: amd64, arm64
- macOS (Darwin): amd64, arm64
- Windows: amd64, arm64

### Binary Distribution

Binaries are compiled and distributed without source code. Users only receive:
- Compiled binary for their platform
- README with usage instructions
- No access to source code

## Testing

### Running Tests

```bash
make test
```

### Testing Authentication Flow

```bash
# Build
make build

# Test login (will fail without real API)
./bin/certfix login --username test --password test123

# Test config
./bin/certfix config set endpoint https://api.example.com
./bin/certfix config get endpoint
./bin/certfix config list

# Test help
./bin/certfix --help
./bin/certfix cert --help
```

## Security Considerations

1. **Token Storage**: JWT tokens stored in `~/.certfix/token.json` with 0600 permissions
2. **Config Directory**: `~/.certfix/` created with 0700 permissions
3. **Password Handling**: Passwords never logged or stored
4. **HTTPS**: All API communications use HTTPS
5. **Token Expiration**: Tokens checked for expiration before use

## Common Workflows

### First-Time Setup

```bash
# Login
certfix login --username user@example.com --password mypassword

# Configure (optional)
certfix config set endpoint https://api.certfix.io

# Create instance
certfix instance create my-instance --type standard --region us-east-1

# Create certificate
certfix cert create example.com
```

### Daily Operations

```bash
# List instances
certfix instance list

# List certificates
certfix cert list

# Renew certificate
certfix cert renew <cert-id>

# Delete instance
certfix instance delete <instance-id>
```

### Troubleshooting

```bash
# Enable verbose logging
certfix --verbose cert list

# Check configuration
certfix config list

# Re-authenticate
certfix logout
certfix login --username user@example.com --password mypassword
```

## Code Style

- Follow standard Go conventions
- Use gofmt for formatting
- Run `make lint` before committing
- Add comments for exported functions
- Keep functions focused and small
- Use meaningful variable names

## Dependencies Management

```bash
# Add new dependency
go get github.com/package/name

# Update dependencies
make deps

# Or manually
go mod download
go mod tidy
```

## Future Enhancements

Potential areas for expansion:
- Shell completion (bash, zsh, fish)
- Output formatting options (JSON, YAML, table)
- Interactive mode for sensitive inputs
- Bulk operations support
- Webhook management
- Monitoring and alerting
- Backup and restore operations
- Multi-profile support
- Plugin system

## Maintenance

### Updating Dependencies

```bash
go get -u ./...
go mod tidy
make test
```

### Creating a Release

1. Update version in Makefile
2. Build all platforms: `make build-all`
3. Test binaries on each platform
4. Create release notes
5. Package and distribute binaries

## Support and Documentation

For contributors and maintainers:
- Keep this file updated with architectural changes
- Document new commands in README.md
- Add inline code comments for complex logic
- Update tests when adding features
- Follow semantic versioning for releases
