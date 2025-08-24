#!/bin/bash

# =============================================================================
# Go Blog Engine Development Setup Script
# =============================================================================

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Default values
SKIP_DEPS="false"
SKIP_DOCKER="false"
SKIP_DB="false"
VERBOSE="false"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# =============================================================================
# FUNCTIONS
# =============================================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_help() {
    cat << EOF
Go Blog Engine Development Setup Script

Usage: $0 [OPTIONS]

Options:
    --skip-deps            Skip dependency installation
    --skip-docker          Skip Docker setup
    --skip-db              Skip database setup
    -v, --verbose          Enable verbose output
    -h, --help             Show this help message

This script will:
    1. Install Go dependencies
    2. Set up Docker environment
    3. Run database migrations
    4. Generate code (SQLC, Swagger)
    5. Verify the setup

Examples:
    $0                      # Full setup
    $0 --skip-docker        # Setup without Docker
    $0 --verbose            # Setup with verbose output

EOF
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    local missing_tools=()

    # Check Go
    if ! command -v go &> /dev/null; then
        missing_tools+=("go")
    fi

    # Check Docker (if not skipping)
    if [[ "$SKIP_DOCKER" != "true" ]] && ! command -v docker &> /dev/null; then
        missing_tools+=("docker")
    fi

    # Check make
    if ! command -v make &> /dev/null; then
        missing_tools+=("make")
    fi

    # Check yq (for database config parsing)
    if ! command -v yq &> /dev/null; then
        log_warning "yq not found. Database operations may not work properly."
        log_info "Install yq: https://github.com/mikefarah/yq#install"
    fi

    # Check sqlc
    if ! command -v sqlc &> /dev/null; then
        log_warning "sqlc not found. Code generation may not work."
        log_info "Install sqlc: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"
    fi

    # Check swag
    if ! command -v swag &> /dev/null; then
        log_warning "swag not found. Swagger documentation generation may not work."
        log_info "Install swag: go install github.com/swaggo/swag/cmd/swag@latest"
    fi

    # Check goose
    if ! command -v goose &> /dev/null; then
        log_warning "goose not found. Database migrations may not work."
        log_info "Install goose: go install github.com/pressly/goose/v3/cmd/goose@latest"
    fi

    if [[ ${#missing_tools[@]} -ne 0 ]]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        exit 1
    fi

    log_success "Prerequisites check passed"
}

setup_environment() {
    log_info "Setting up project environment..."

    cd "$PROJECT_ROOT"

    # Create necessary directories
    mkdir -p bin
    mkdir -p logs
    mkdir -p docs

    # Ensure config directory exists
    if [[ ! -d "config/env" ]]; then
        log_error "Config directory not found. Please ensure config/env/ exists with environment files."
        exit 1
    fi

    # Check if development config exists
    if [[ ! -f "config/env/develop.yml" ]]; then
        if [[ -f "config/env/example.yml" ]]; then
            log_warning "develop.yml not found. Consider copying from example.yml"
            log_info "cp config/env/example.yml config/env/develop.yml"
        else
            log_error "No configuration files found in config/env/"
            exit 1
        fi
    fi

    log_success "Environment setup complete"
}

install_dependencies() {
    if [[ "$SKIP_DEPS" == "true" ]]; then
        log_warning "Skipping dependency installation"
        return 0
    fi

    log_info "Installing Go dependencies..."

    cd "$PROJECT_ROOT"

    if [[ "$VERBOSE" == "true" ]]; then
        go mod download
        go mod tidy
    else
        go mod download > /dev/null 2>&1
        go mod tidy > /dev/null 2>&1
    fi

    log_success "Dependencies installed"
}

setup_docker() {
    if [[ "$SKIP_DOCKER" == "true" ]]; then
        log_warning "Skipping Docker setup"
        return 0
    fi

    log_info "Setting up Docker environment..."

    cd "$PROJECT_ROOT"

    # Create docker network
    if ! docker network ls | grep -q "blog_engine_network"; then
        log_info "Creating Docker network..."
        docker network create --driver bridge blog_engine_network
    fi

    # Start Docker services
    if [[ "$VERBOSE" == "true" ]]; then
        make docker-up
    else
        make docker-up > /dev/null 2>&1
    fi

    # Wait for services to be ready
    log_info "Waiting for services to be ready..."
    sleep 5

    log_success "Docker environment ready"
}

setup_database() {
    if [[ "$SKIP_DB" == "true" ]]; then
        log_warning "Skipping database setup"
        return 0
    fi

    log_info "Setting up database..."

    cd "$PROJECT_ROOT"

    # Run migrations
    if [[ "$VERBOSE" == "true" ]]; then
        make migrate-up
    else
        make migrate-up > /dev/null 2>&1
    fi

    log_success "Database setup complete"
}

generate_code() {
    log_info "Generating code..."

    cd "$PROJECT_ROOT"

    # Generate SQLC code
    if command -v sqlc &> /dev/null; then
        if [[ "$VERBOSE" == "true" ]]; then
            make sqlc-gen
        else
            make sqlc-gen > /dev/null 2>&1
        fi
        log_success "SQLC code generated"
    else
        log_warning "Skipping SQLC generation (sqlc not found)"
    fi

    # Generate Swagger docs
    if command -v swag &> /dev/null; then
        if [[ "$VERBOSE" == "true" ]]; then
            make swag-init
        else
            make swag-init > /dev/null 2>&1 || true
        fi
        log_success "Swagger documentation generated"
    else
        log_warning "Skipping Swagger generation (swag not found)"
    fi
}

verify_setup() {
    log_info "Verifying setup..."

    cd "$PROJECT_ROOT"

    # Check if binary can be built
    if go build -o /tmp/blog-engine-test ./cmd/server > /dev/null 2>&1; then
        rm -f /tmp/blog-engine-test
        log_success "Build verification passed"
    else
        log_error "Build verification failed"
        return 1
    fi

    # Check if tests pass
    if [[ "$VERBOSE" == "true" ]]; then
        go test ./...
    else
        if go test ./... > /dev/null 2>&1; then
            log_success "Test verification passed"
        else
            log_warning "Some tests failed"
        fi
    fi
}

show_next_steps() {
    cat << EOF

${GREEN}✅ Setup completed successfully!${NC}

${BLUE}Next steps:${NC}
    1. Start the development server:
       ${YELLOW}make dev${NC} or ${YELLOW}make server${NC}

    2. View API documentation:
       ${YELLOW}http://localhost:8080/swagger/index.html${NC}

    3. Common development commands:
       ${YELLOW}make help${NC}                 # Show all available commands
       ${YELLOW}make test${NC}                 # Run tests
       ${YELLOW}make lint${NC}                 # Run linter
       ${YELLOW}make migrate-status${NC}       # Check migration status

${BLUE}Useful files:${NC}
    - Configuration: ${YELLOW}config/env/develop.yml${NC}
    - Docker services: ${YELLOW}environments/docker/docker-compose.yml${NC}
    - Makefile: ${YELLOW}./Makefile${NC}

EOF
}

# =============================================================================
# MAIN
# =============================================================================

main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-deps)
                SKIP_DEPS="true"
                shift
                ;;
            --skip-docker)
                SKIP_DOCKER="true"
                shift
                ;;
            --skip-db)
                SKIP_DB="true"
                shift
                ;;
            -v|--verbose)
                VERBOSE="true"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    log_info "Starting development environment setup..."

    # Execute setup steps
    check_prerequisites
    setup_environment
    install_dependencies
    setup_docker
    setup_database
    generate_code
    verify_setup

    show_next_steps
}

# Run main function with all arguments
main "$@"
