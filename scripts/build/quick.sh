#!/bin/bash

# =============================================================================
# Go Blog Engine - Quick Development Commands
# =============================================================================

# This is a simple wrapper script for the most common development tasks
# Place this in your PATH or create an alias for quick access

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

show_help() {
    cat << EOF
${GREEN}Go Blog Engine - Quick Commands${NC}

Usage: $0 <command>

Quick Commands:
    ${YELLOW}start${NC}      Start development environment
    ${YELLOW}stop${NC}       Stop development environment
    ${YELLOW}restart${NC}    Restart development environment
    ${YELLOW}test${NC}       Run tests
    ${YELLOW}build${NC}      Quick build
    ${YELLOW}clean${NC}      Clean environment
    ${YELLOW}setup${NC}      Initial setup
    ${YELLOW}status${NC}     Show status
    ${YELLOW}logs${NC}       Show logs
    ${YELLOW}migrate${NC}    Run migrations
    ${YELLOW}help${NC}       Show detailed help

For more options, use the full scripts in scripts/build/

EOF
}

main() {
    cd "$PROJECT_ROOT"

    case "${1:-help}" in
        start)
            echo -e "${BLUE}Starting development environment...${NC}"
            scripts/build/dev.sh start
            ;;
        stop)
            echo -e "${BLUE}Stopping development environment...${NC}"
            scripts/build/dev.sh stop
            ;;
        restart)
            echo -e "${BLUE}Restarting development environment...${NC}"
            scripts/build/dev.sh restart
            ;;
        test)
            echo -e "${BLUE}Running tests...${NC}"
            scripts/build/dev.sh test
            ;;
        build)
            echo -e "${BLUE}Running quick build...${NC}"
            scripts/build/dev.sh build
            ;;
        clean)
            echo -e "${BLUE}Cleaning environment...${NC}"
            scripts/build/clean.sh
            ;;
        setup)
            echo -e "${BLUE}Running setup...${NC}"
            scripts/build/setup.sh
            ;;
        status)
            scripts/build/dev.sh status
            ;;
        logs)
            scripts/build/dev.sh logs --follow
            ;;
        migrate)
            echo -e "${BLUE}Running migrations...${NC}"
            scripts/build/dev.sh db migrate
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            echo -e "${YELLOW}Unknown command: $1${NC}"
            echo
            show_help
            exit 1
            ;;
    esac
}

main "$@"
