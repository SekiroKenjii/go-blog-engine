#!/bin/bash

# =============================================================================
# Go Blog Engine Clean Script
# =============================================================================

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Default values
CLEAN_DOCKER="true"
CLEAN_BUILDS="true"
CLEAN_DEPS="false"
CLEAN_LOGS="false"
FORCE="false"
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
Go Blog Engine Clean Script

Usage: $0 [OPTIONS]

Options:
    --no-docker            Skip Docker cleanup
    --no-builds            Skip build artifacts cleanup
    --clean-deps           Clean Go module cache
    --clean-logs           Clean log files
    -f, --force            Force cleanup without confirmation
    -v, --verbose          Enable verbose output
    -h, --help             Show this help message

What gets cleaned by default:
    - Docker containers and networks
    - Build artifacts (bin/ directory)
    - Go build cache
    - Test coverage files

With additional flags:
    --clean-deps           Go module cache (~/.go/pkg/mod)
    --clean-logs           Log files

Examples:
    $0                      # Clean Docker and builds
    $0 --no-docker          # Clean only builds
    $0 --clean-deps -f      # Clean everything including deps, no confirmation
    $0 --clean-logs -v      # Clean with logs, verbose output

EOF
}

confirm_action() {
    local message="$1"

    if [[ "$FORCE" == "true" ]]; then
        return 0
    fi

    echo -e "${YELLOW}$message${NC}"
    read -p "Are you sure? [y/N] " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        return 0
    else
        return 1
    fi
}

clean_docker() {
    if [[ "$CLEAN_DOCKER" != "true" ]]; then
        log_info "Skipping Docker cleanup"
        return 0
    fi

    log_info "Cleaning Docker resources..."

    cd "$PROJECT_ROOT"

    # Check if Docker is available
    if ! command -v docker &> /dev/null; then
        log_warning "Docker not found, skipping Docker cleanup"
        return 0
    fi

    # Stop and remove containers
    if [[ "$VERBOSE" == "true" ]]; then
        make docker-down
    else
        make docker-down > /dev/null 2>&1 || true
    fi

    # Remove the Docker network
    if docker network ls | grep -q "blog_engine_network"; then
        log_info "Removing Docker network..."
        docker network rm blog_engine_network 2>/dev/null || true
    fi

    # Clean up Docker system (with confirmation unless forced)
    if confirm_action "This will remove unused Docker images, containers, and networks."; then
        log_info "Cleaning Docker system..."
        if [[ "$VERBOSE" == "true" ]]; then
            docker system prune -f
        else
            docker system prune -f > /dev/null 2>&1
        fi
    fi

    log_success "Docker cleanup complete"
}

clean_builds() {
    if [[ "$CLEAN_BUILDS" != "true" ]]; then
        log_info "Skipping build cleanup"
        return 0
    fi

    log_info "Cleaning build artifacts..."

    cd "$PROJECT_ROOT"

    # Remove build directory
    if [[ -d "bin" ]]; then
        log_info "Removing bin/ directory..."
        rm -rf bin/
    fi

    # Clean Go build cache
    log_info "Cleaning Go build cache..."
    if [[ "$VERBOSE" == "true" ]]; then
        go clean -cache
        go clean -testcache
    else
        go clean -cache > /dev/null 2>&1
        go clean -testcache > /dev/null 2>&1
    fi

    # Remove coverage files
    if [[ -f "coverage.out" ]]; then
        log_info "Removing coverage files..."
        rm -f coverage.out coverage.html
    fi

    # Remove any generated documentation
    if [[ -d "docs" ]] && [[ -f "docs/docs.go" ]]; then
        log_info "Cleaning generated documentation..."
        rm -f docs/docs.go docs/swagger.json docs/swagger.yaml
    fi

    log_success "Build cleanup complete"
}

clean_dependencies() {
    if [[ "$CLEAN_DEPS" != "true" ]]; then
        return 0
    fi

    if confirm_action "This will clean the Go module cache. You'll need to re-download dependencies."; then
        log_info "Cleaning Go module cache..."
        if [[ "$VERBOSE" == "true" ]]; then
            go clean -modcache
        else
            go clean -modcache > /dev/null 2>&1
        fi
        log_success "Dependencies cleanup complete"
    fi
}

clean_logs() {
    if [[ "$CLEAN_LOGS" != "true" ]]; then
        return 0
    fi

    log_info "Cleaning log files..."

    cd "$PROJECT_ROOT"

    # Remove log files
    if [[ -d "logs" ]]; then
        if confirm_action "This will remove all log files in the logs/ directory."; then
            rm -rf logs/*
            log_success "Log files cleaned"
        fi
    fi

    # Remove any .log files in the project root
    find . -name "*.log" -type f -not -path "./vendor/*" -delete 2>/dev/null || true
}

# =============================================================================
# MAIN
# =============================================================================

main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --no-docker)
                CLEAN_DOCKER="false"
                shift
                ;;
            --no-builds)
                CLEAN_BUILDS="false"
                shift
                ;;
            --clean-deps)
                CLEAN_DEPS="true"
                shift
                ;;
            --clean-logs)
                CLEAN_LOGS="true"
                shift
                ;;
            -f|--force)
                FORCE="true"
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

    # Check if we're in the right directory
    if [[ ! -f "${PROJECT_ROOT}/go.mod" ]]; then
        log_error "go.mod not found. Please run this script from the project root or scripts/build directory."
        exit 1
    fi

    log_info "Starting cleanup process..."

    # Execute cleanup steps
    clean_docker
    clean_builds
    clean_dependencies
    clean_logs

    log_success "Cleanup completed successfully!"
}

# Run main function with all arguments
main "$@"
