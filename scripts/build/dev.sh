#!/bin/bash

# =============================================================================
# Go Blog Engine Development Helper Script
# =============================================================================

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

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
Go Blog Engine Development Helper

Usage: $0 <command> [options]

Commands:
    start               Start development environment (docker + server)
    stop                Stop development environment
    restart             Restart development environment
    logs                Show application logs
    db                  Database operations
    test                Run tests with options
    lint                Run linting and formatting
    build               Quick build verification
    clean               Clean development environment
    status              Show environment status

Command-specific help:
    $0 <command> --help

Examples:
    $0 start            # Start everything
    $0 db migrate       # Run database migrations
    $0 test --coverage  # Run tests with coverage
    $0 logs --follow    # Follow application logs

EOF
}

show_db_help() {
    cat << EOF
Database operations:

Usage: $0 db <operation> [options]

Operations:
    migrate             Run migrations up
    rollback            Rollback last migration
    reset               Reset database (WARNING: destructive)
    status              Show migration status
    create <name>       Create new migration
    shell               Open database shell

Examples:
    $0 db migrate
    $0 db create add_user_table
    $0 db status

EOF
}

show_test_help() {
    cat << EOF
Test operations:

Usage: $0 test [options]

Options:
    --coverage          Run with coverage report
    --integration       Run integration tests only
    --unit              Run unit tests only
    --watch             Run tests in watch mode
    --verbose           Verbose output

Examples:
    $0 test --coverage
    $0 test --integration
    $0 test --watch

EOF
}

start_dev() {
    log_info "Starting development environment..."

    cd "$PROJECT_ROOT"

    # Check if services are already running
    if docker compose -f environments/docker/docker-compose.yml -p go-blog-engine ps | grep -q "Up"; then
        log_warning "Services already running. Use 'restart' to restart them."
    else
        make docker-up
        sleep 3
        make migrate-up
    fi

    log_info "Starting development server..."
    log_info "Press Ctrl+C to stop"

    # Start the server (this will block)
    make server
}

stop_dev() {
    log_info "Stopping development environment..."

    cd "$PROJECT_ROOT"

    # Stop docker services
    make docker-stop

    log_success "Development environment stopped"
}

restart_dev() {
    log_info "Restarting development environment..."

    stop_dev
    sleep 2

    cd "$PROJECT_ROOT"
    make docker-start
    sleep 3

    log_success "Development environment restarted"
}

show_logs() {
    local follow="false"

    # Parse options
    while [[ $# -gt 0 ]]; do
        case $1 in
            --follow|-f)
                follow="true"
                shift
                ;;
            --help)
                echo "Usage: $0 logs [--follow]"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    cd "$PROJECT_ROOT"

    if [[ "$follow" == "true" ]]; then
        make docker-logs
    else
        docker compose -f environments/docker/docker-compose.yml -p go-blog-engine logs
    fi
}

db_operations() {
    if [[ $# -eq 0 ]]; then
        show_db_help
        exit 1
    fi

    local operation="$1"
    shift

    cd "$PROJECT_ROOT"

    case "$operation" in
        migrate)
            log_info "Running database migrations..."
            make migrate-up
            ;;
        rollback)
            log_warning "Rolling back last migration..."
            make migrate-down
            ;;
        reset)
            log_error "This will reset the entire database!"
            read -p "Are you sure? Type 'yes' to confirm: " -r
            if [[ $REPLY == "yes" ]]; then
                make migrate-reset
            else
                log_info "Aborted."
            fi
            ;;
        status)
            make migrate-status
            ;;
        create)
            if [[ $# -eq 0 ]]; then
                log_error "Migration name required"
                exit 1
            fi
            local migration_name="$1"
            make migrate-create name="$migration_name"
            ;;
        shell)
            log_info "Opening database shell..."
            # This would need to be implemented based on your database setup
            log_warning "Database shell not implemented yet"
            ;;
        --help)
            show_db_help
            ;;
        *)
            log_error "Unknown database operation: $operation"
            show_db_help
            exit 1
            ;;
    esac
}

test_operations() {
    local coverage="false"
    local integration="false"
    local unit="false"
    local watch="false"
    local verbose="false"

    # Parse options
    while [[ $# -gt 0 ]]; do
        case $1 in
            --coverage)
                coverage="true"
                shift
                ;;
            --integration)
                integration="true"
                shift
                ;;
            --unit)
                unit="true"
                shift
                ;;
            --watch)
                watch="true"
                shift
                ;;
            --verbose)
                verbose="true"
                shift
                ;;
            --help)
                show_test_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    cd "$PROJECT_ROOT"

    if [[ "$coverage" == "true" ]]; then
        make test-coverage
    elif [[ "$integration" == "true" ]]; then
        make test-integration
    elif [[ "$watch" == "true" ]]; then
        log_info "Running tests in watch mode (install entr: apt-get install entr)"
        find . -name "*.go" | entr -c go test ./...
    else
        if [[ "$verbose" == "true" ]]; then
            go test -v ./...
        else
            make test
        fi
    fi
}

lint_operations() {
    cd "$PROJECT_ROOT"

    log_info "Running linting and formatting..."

    # Format code
    make fmt

    # Run linter if available
    if command -v golangci-lint &> /dev/null; then
        make lint
    else
        log_warning "golangci-lint not found, running go vet instead"
        make vet
    fi

    log_success "Linting complete"
}

quick_build() {
    cd "$PROJECT_ROOT"

    log_info "Running quick build verification..."

    # Generate code first
    make generate

    # Build
    make build

    log_success "Build verification complete"
}

clean_dev() {
    cd "$PROJECT_ROOT"

    log_info "Cleaning development environment..."

    # Use the clean script
    bash scripts/build/clean.sh --force

    log_success "Environment cleaned"
}

show_status() {
    cd "$PROJECT_ROOT"

    echo -e "${BLUE}=== Development Environment Status ===${NC}"
    echo

    # Docker status
    echo -e "${YELLOW}Docker Services:${NC}"
    if command -v docker &> /dev/null; then
        docker compose -f environments/docker/docker-compose.yml -p go-blog-engine ps
    else
        echo "Docker not available"
    fi
    echo

    # Database status
    echo -e "${YELLOW}Database Status:${NC}"
    make migrate-status 2>/dev/null || echo "Unable to check database status"
    echo

    # Git status
    echo -e "${YELLOW}Git Status:${NC}"
    git status --porcelain || echo "Not a git repository"
    echo

    # Go environment
    echo -e "${YELLOW}Go Environment:${NC}"
    go version
    echo "GOPATH: $(go env GOPATH)"
    echo "GOROOT: $(go env GOROOT)"
}

# =============================================================================
# MAIN
# =============================================================================

main() {
    if [[ $# -eq 0 ]]; then
        show_help
        exit 1
    fi

    local command="$1"
    shift

    # Check if we're in the right directory
    if [[ ! -f "${PROJECT_ROOT}/go.mod" ]]; then
        log_error "go.mod not found. Please run this script from the project root or scripts/build directory."
        exit 1
    fi

    case "$command" in
        start)
            start_dev "$@"
            ;;
        stop)
            stop_dev "$@"
            ;;
        restart)
            restart_dev "$@"
            ;;
        logs)
            show_logs "$@"
            ;;
        db)
            db_operations "$@"
            ;;
        test)
            test_operations "$@"
            ;;
        lint)
            lint_operations "$@"
            ;;
        build)
            quick_build "$@"
            ;;
        clean)
            clean_dev "$@"
            ;;
        status)
            show_status "$@"
            ;;
        --help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
