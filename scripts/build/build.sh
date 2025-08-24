#!/bin/bash

# =============================================================================
# Go Blog Engine Build Script
# =============================================================================

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Default values
BUILD_TYPE="dev"
SKIP_TESTS="false"
VERBOSE="false"
PLATFORM="linux/amd64"

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
Go Blog Engine Build Script

Usage: $0 [OPTIONS]

Options:
    -t, --type TYPE         Build type: dev, prod, docker (default: dev)
    -p, --platform PLATFORM Target platform for docker build (default: linux/amd64)
    --skip-tests           Skip running tests
    -v, --verbose          Enable verbose output
    -h, --help             Show this help message

Build Types:
    dev     - Development build with debugging info
    prod    - Production build optimized for deployment
    docker  - Docker containerized build

Examples:
    $0                      # Development build
    $0 -t prod              # Production build
    $0 -t docker            # Docker build
    $0 --skip-tests -v      # Skip tests with verbose output

EOF
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check if we're in the right directory
    if [[ ! -f "${PROJECT_ROOT}/go.mod" ]]; then
        log_error "go.mod not found. Please run this script from the project root or scripts/build directory."
        exit 1
    fi

    # Check Go installation
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi

    # Check Docker installation for docker builds
    if [[ "$BUILD_TYPE" == "docker" ]] && ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi

    log_success "Prerequisites check passed"
}

setup_environment() {
    log_info "Setting up build environment..."

    cd "$PROJECT_ROOT"

    # Create necessary directories
    mkdir -p bin
    mkdir -p logs

    # Ensure docker network exists for docker builds
    if [[ "$BUILD_TYPE" == "docker" ]]; then
        if ! docker network ls | grep -q "blog_engine_network"; then
            log_info "Creating Docker network..."
            docker network create --driver bridge blog_engine_network || true
        fi
    fi

    log_success "Environment setup complete"
}

run_tests() {
    if [[ "$SKIP_TESTS" == "true" ]]; then
        log_warning "Skipping tests as requested"
        return 0
    fi

    log_info "Running tests..."

    if [[ "$VERBOSE" == "true" ]]; then
        go test -v ./...
    else
        go test ./...
    fi

    log_success "Tests passed"
}

build_dev() {
    log_info "Building for development..."

    local binary_name="server"
    local build_flags=""

    if [[ "$VERBOSE" == "true" ]]; then
        build_flags="-v"
    fi

    # Build with debug information
    go build $build_flags \
        -ldflags "-X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.GitCommit=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')" \
        -o "bin/${binary_name}" \
        ./cmd/server

    log_success "Development build complete: bin/${binary_name}"
}

build_prod() {
    log_info "Building for production..."

    local binary_name="server"
    local build_flags="-trimpath"

    if [[ "$VERBOSE" == "true" ]]; then
        build_flags="$build_flags -v"
    fi

    # Build optimized for production
    CGO_ENABLED=0 go build $build_flags \
        -ldflags "-s -w -X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev') -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.GitCommit=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')" \
        -o "bin/${binary_name}" \
        ./cmd/server

    log_success "Production build complete: bin/${binary_name}"
}

build_docker() {
    log_info "Building Docker containers..."

    # Build and start containers
    if [[ "$VERBOSE" == "true" ]]; then
        make docker-build
    else
        make docker-build > /dev/null 2>&1
    fi

    log_success "Docker build complete"
}

cleanup_on_error() {
    log_error "Build failed. Cleaning up..."
    # Add any cleanup logic here
    exit 1
}

# =============================================================================
# MAIN
# =============================================================================

main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--type)
                BUILD_TYPE="$2"
                shift 2
                ;;
            -p|--platform)
                PLATFORM="$2"
                shift 2
                ;;
            --skip-tests)
                SKIP_TESTS="true"
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

    # Validate build type
    if [[ ! "$BUILD_TYPE" =~ ^(dev|prod|docker)$ ]]; then
        log_error "Invalid build type: $BUILD_TYPE"
        show_help
        exit 1
    fi

    # Set up error handling
    trap cleanup_on_error ERR

    log_info "Starting build process (type: $BUILD_TYPE)"

    # Execute build steps
    check_prerequisites
    setup_environment

    # Only run tests for non-docker builds (docker will run its own tests)
    if [[ "$BUILD_TYPE" != "docker" ]]; then
        run_tests
    fi

    # Execute the appropriate build
    case "$BUILD_TYPE" in
        "dev")
            build_dev
            ;;
        "prod")
            build_prod
            ;;
        "docker")
            build_docker
            ;;
    esac

    log_success "Build completed successfully!"
}

# Run main function with all arguments
main "$@"
