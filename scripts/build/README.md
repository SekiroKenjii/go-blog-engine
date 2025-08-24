# Build and Development Scripts

This directory contains scripts to help with building, deploying, and managing the Go Blog Engine development environment.

## Scripts Overview

### 🏗️ `build.sh`
**Comprehensive build script with multiple build types**

```bash
# Development build (default)
./scripts/build/build.sh

# Production build
./scripts/build/build.sh -t prod

# Docker build
./scripts/build/build.sh -t docker

# Skip tests and run verbose
./scripts/build/build.sh --skip-tests -v
```

**Features:**
- Multiple build types (dev, prod, docker)
- Automatic testing before build
- Verbose output option
- Error handling and cleanup
- Build optimization for production

### 🧹 `clean.sh`
**Intelligent cleanup script**

```bash
# Clean everything (default)
./scripts/build/clean.sh

# Clean only builds, skip Docker
./scripts/build/clean.sh --no-docker

# Force clean including dependencies
./scripts/build/clean.sh --clean-deps -f

# Clean with logs and verbose output
./scripts/build/clean.sh --clean-logs -v
```

**Features:**
- Selective cleanup options
- Docker resource cleanup
- Go cache and dependency cleanup
- Safety confirmations (unless forced)

### 🚀 `setup.sh`
**Complete development environment setup**

```bash
# Full setup
./scripts/build/setup.sh

# Setup without Docker
./scripts/build/setup.sh --skip-docker

# Verbose setup
./scripts/build/setup.sh -v
```

**Features:**
- Dependency installation
- Docker environment setup
- Database migration
- Code generation (SQLC, Swagger)
- Setup verification

### 🔧 `dev.sh`
**Development workflow helper**

```bash
# Start development environment
./scripts/build/dev.sh start

# Database operations
./scripts/build/dev.sh db migrate
./scripts/build/dev.sh db create add_user_table

# Testing
./scripts/build/dev.sh test --coverage

# Show environment status
./scripts/build/dev.sh status
```

**Features:**
- Start/stop development environment
- Database management
- Testing with various options
- Environment status monitoring
- Log viewing

## Quick Start

### First Time Setup
```bash
# 1. Run complete setup
./scripts/build/setup.sh

# 2. Start development
./scripts/build/dev.sh start
```

### Daily Development Workflow
```bash
# Start development environment
./scripts/build/dev.sh start

# Run tests
./scripts/build/dev.sh test --coverage

# Check database status
./scripts/build/dev.sh db status

# Clean environment when needed
./scripts/build/clean.sh
```

### Building for Production
```bash
# Clean build for production
./scripts/build/clean.sh
./scripts/build/build.sh -t prod
```

## Integration with Makefile

These scripts are designed to work alongside the project Makefile. Many operations can be performed either way:

**Via Scripts:**
```bash
./scripts/build/dev.sh test --coverage
./scripts/build/dev.sh db migrate
```

**Via Makefile:**
```bash
make test-coverage
make migrate-up
```

Choose based on your preference:
- **Scripts**: More interactive, better error messages, guided workflows
- **Makefile**: Faster for experienced users, better for CI/CD

## Environment Variables

Scripts respect these environment variables:

- `ENV`: Environment name (default: `develop`)
- `VERBOSE`: Enable verbose output (`true`/`false`)
- `SKIP_TESTS`: Skip tests during build (`true`/`false`)

Example:
```bash
ENV=production VERBOSE=true ./scripts/build/build.sh -t prod
```

## Prerequisites

Required tools for full functionality:
- **Go** (1.21+)
- **Docker** & Docker Compose
- **Make**

Optional tools (will show warnings if missing):
- **yq** - For config parsing
- **sqlc** - For code generation
- **swag** - For Swagger docs
- **goose** - For database migrations
- **golangci-lint** - For linting

Install optional tools:
```bash
# Go tools
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/pressly/goose/v3/cmd/goose@latest

# System tools
# yq: https://github.com/mikefarah/yq#install
# golangci-lint: https://golangci-lint.run/usage/install/
```

## Troubleshooting

### Common Issues

**Docker network issues:**
```bash
./scripts/build/clean.sh --force
./scripts/build/setup.sh
```

**Database connection issues:**
```bash
./scripts/build/dev.sh db status
./scripts/build/dev.sh restart
```

**Build failures:**
```bash
./scripts/build/clean.sh --clean-deps
./scripts/build/setup.sh --verbose
```

### Getting Help

Each script has built-in help:
```bash
./scripts/build/build.sh --help
./scripts/build/clean.sh --help
./scripts/build/setup.sh --help
./scripts/build/dev.sh --help
```

For specific commands:
```bash
./scripts/build/dev.sh db --help
./scripts/build/dev.sh test --help
```

## Contributing

When adding new scripts:

1. Follow the established patterns
2. Include comprehensive help text
3. Add error handling and validation
4. Update this README
5. Make scripts executable: `chmod +x script_name.sh`

### Script Structure Template

```bash
#!/bin/bash
set -euo pipefail

# Colors and logging functions
RED='\033[0;31m'
GREEN='\033[0;32m'
# ... etc

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
# ... etc

show_help() {
    cat << EOF
Script description and usage
EOF
}

main() {
    # Argument parsing
    # Validation
    # Main logic
}

main "$@"
```
