#!/bin/bash
# =============================================================================
# RealWorld Conduit - Development Server Script
# =============================================================================
# This script starts both backend and frontend development servers
# with proper process management and cleanup.
# =============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PIDFILE="${PROJECT_ROOT}/.dev.pid"

# Print colored message
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Cleanup function to kill background processes
cleanup() {
    log_info "Shutting down development servers..."

    # Kill processes from pidfile
    if [ -f "$PIDFILE" ]; then
        while read -r pid; do
            if kill -0 "$pid" 2>/dev/null; then
                kill "$pid" 2>/dev/null || true
                log_info "Stopped process $pid"
            fi
        done < "$PIDFILE"
        rm -f "$PIDFILE"
    fi

    # Kill any remaining processes on our ports
    lsof -ti:8080 | xargs kill -9 2>/dev/null || true
    lsof -ti:5173 | xargs kill -9 2>/dev/null || true

    log_success "Development servers stopped"
    exit 0
}

# Trap signals for cleanup
trap cleanup SIGINT SIGTERM EXIT

# Check if servers are already running
check_existing() {
    if [ -f "$PIDFILE" ]; then
        log_warn "Development servers may already be running"
        log_info "Cleaning up previous session..."
        cleanup
    fi
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check Go
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21+"
        exit 1
    fi

    # Check Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js is not installed. Please install Node.js 18+"
        exit 1
    fi

    # Check npm
    if ! command -v npm &> /dev/null; then
        log_error "npm is not installed. Please install npm"
        exit 1
    fi

    log_success "All prerequisites met"
}

# Start backend server
start_backend() {
    log_info "Starting backend server on port 8080..."
    cd "${PROJECT_ROOT}/backend"

    # Build and run
    go run ./cmd/server/main.go &
    BACKEND_PID=$!
    echo "$BACKEND_PID" >> "$PIDFILE"

    # Wait for backend to be ready
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            log_success "Backend server started (PID: $BACKEND_PID)"
            return 0
        fi
        sleep 1
    done

    log_error "Backend server failed to start"
    return 1
}

# Start frontend server
start_frontend() {
    log_info "Starting frontend server on port 5173..."
    cd "${PROJECT_ROOT}/frontend"

    # Install dependencies if node_modules doesn't exist
    if [ ! -d "node_modules" ]; then
        log_info "Installing frontend dependencies..."
        npm install
    fi

    # Start Vite dev server
    npm run dev &
    FRONTEND_PID=$!
    echo "$FRONTEND_PID" >> "$PIDFILE"

    # Wait for frontend to be ready
    for i in {1..30}; do
        if curl -s http://localhost:5173 > /dev/null 2>&1; then
            log_success "Frontend server started (PID: $FRONTEND_PID)"
            return 0
        fi
        sleep 1
    done

    log_error "Frontend server failed to start"
    return 1
}

# Main function
main() {
    echo ""
    echo "==========================================="
    echo "  RealWorld Conduit - Development Mode"
    echo "==========================================="
    echo ""

    check_existing
    check_prerequisites

    # Create empty pidfile
    > "$PIDFILE"

    # Start servers
    start_backend
    start_frontend

    echo ""
    log_success "Development servers are running!"
    echo ""
    echo "  Backend:  http://localhost:8080"
    echo "  Frontend: http://localhost:5173"
    echo "  Health:   http://localhost:8080/health"
    echo ""
    echo "Press Ctrl+C to stop all servers"
    echo ""

    # Wait for user interrupt
    wait
}

# Run main function
main "$@"
