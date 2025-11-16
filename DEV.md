# 1. Build for your current platform (macOS)

make build

# 2. Test the configure command

./bin/certfix configure

# Or test with flags

./bin/certfix configure --api-url https://localhost:8080

# Check the configuration was saved

./bin/certfix config list

# Test other commands

./bin/certfix version

# Dev MODE

# Run directly with go run

go run main.go configure

# With flags

go run main.go configure --api-url https://api.example.com --timeout 60

# Check config

go run main.go config list

# Install LOCAL and TEST

# Install to your GOBIN

make install

# Now you can use it directly

certfix configure
certfix version
certfix config list

# Remove the configurations

cat ~/.certfix/config.yaml
cat ~/.certfix/token.json
