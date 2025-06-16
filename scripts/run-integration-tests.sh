#!/bin/bash

# Integration Test Runner Script
# This script sets up and runs integration tests with real databases

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to wait for MySQL to be ready
wait_for_mysql() {
    local host=${1:-127.0.0.1}
    local port=${2:-3306}
    local user=${3:-root}
    local password=${4:-test_password}
    local max_attempts=60
    local attempt=1

    print_status "Waiting for MySQL to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        # Check if port is open first
        if nc -z "$host" "$port" 2>/dev/null; then
            # If port is open, try MySQL ping command
            if mysqladmin ping -h"$host" -P"$port" -u"$user" -p"$password" --silent 2>/dev/null; then
                print_success "MySQL is ready!"
                return 0
            fi
        fi
        
        echo "Attempt $attempt/$max_attempts: MySQL not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_error "MySQL failed to start within expected time"
    return 1
}

# Function to wait for Redis to be ready
wait_for_redis() {
    local host=${1:-127.0.0.1}
    local port=${2:-6379}
    local max_attempts=60
    local attempt=1

    print_status "Waiting for Redis to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        # Check if port is open first
        if nc -z "$host" "$port" 2>/dev/null; then
            # If port is open, try Redis ping command
            if redis-cli -h "$host" -p "$port" ping >/dev/null 2>&1; then
                print_success "Redis is ready!"
                return 0
            fi
        fi
        
        echo "Attempt $attempt/$max_attempts: Redis not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_error "Redis failed to start within expected time"
    return 1
}

# Function to setup test environment
setup_test_env() {
    print_status "Setting up test environment..."
    
    # Load test environment variables
    if [ -f ".env.test" ]; then
        export $(grep -v '^#' .env.test | xargs)
        print_success "Loaded test environment variables from .env.test"
    else
        print_warning ".env.test file not found, using default environment"
    fi
}

# Function to run database migrations
run_migrations() {
    print_status "Running database migrations..."
    
    if go run ./cmd/migrate/main.go; then
        print_success "Database migrations completed"
    else
        print_error "Database migrations failed"
        return 1
    fi
}

# Function to start test databases with Docker
start_test_databases() {
    print_status "Starting test databases with Docker..."
    
    if ! command_exists docker-compose; then
        print_error "docker-compose is not installed"
        return 1
    fi
    
    # Stop any existing test containers
    docker-compose -f docker-compose.test.yml down >/dev/null 2>&1 || true
    
    # Start test databases
    if docker-compose -f docker-compose.test.yml up -d; then
        print_success "Test databases started"
        
        # Wait for databases to be ready
        wait_for_mysql "127.0.0.1" "3307" "root" "test_password"
        wait_for_redis "127.0.0.1" "6380"
        
        # Update environment for test databases
        export MYSQL_PORT=3307
        export REDIS_PORT=6380
        
        return 0
    else
        print_error "Failed to start test databases"
        return 1
    fi
}

# Function to stop test databases
stop_test_databases() {
    print_status "Stopping test databases..."
    
    if docker-compose -f docker-compose.test.yml down -v >/dev/null 2>&1; then
        print_success "Test databases stopped"
    else
        print_warning "Failed to stop test databases (they may not have been running)"
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_status "Running integration tests..."
    
    # Check if Go is installed
    if ! command_exists go; then
        print_error "Go is not installed"
        return 1
    fi
    
    # Download dependencies
    print_status "Downloading Go dependencies..."
    go mod download
    
    # Run integration tests
    if go test -v -tags=integration -coverprofile=integration-coverage.out ./tests/integration/...; then
        print_success "Integration tests passed!"
        
        # Generate coverage report if go tool cover is available
        if command_exists go && go tool cover >/dev/null 2>&1; then
            print_status "Generating coverage report..."
            go tool cover -html=integration-coverage.out -o integration-coverage.html
            print_success "Coverage report generated: integration-coverage.html"
        fi
        
        return 0
    else
        print_error "Integration tests failed!"
        return 1
    fi
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up..."
    stop_test_databases
}

# Main function
main() {
    local use_docker=true
    local skip_migrations=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --no-docker)
                use_docker=false
                shift
                ;;
            --skip-migrations)
                skip_migrations=true
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --no-docker        Use existing databases instead of starting Docker containers"
                echo "  --skip-migrations  Skip running database migrations"
                echo "  --help            Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
    
    print_status "Starting integration test runner..."
    
    # Setup trap for cleanup
    trap cleanup EXIT
    
    # Setup test environment
    setup_test_env
    
    # Start databases if using Docker
    if [ "$use_docker" = true ]; then
        start_test_databases
    else
        print_status "Using existing databases (--no-docker specified)"
        
        # Wait for existing databases
        wait_for_mysql "$MYSQL_HOST" "$MYSQL_PORT" "$MYSQL_USER" "$MYSQL_PASSWORD"
        wait_for_redis "$REDIS_HOST" "$REDIS_PORT"
    fi
    
    # Run migrations unless skipped
    if [ "$skip_migrations" = false ]; then
        run_migrations
    else
        print_status "Skipping database migrations (--skip-migrations specified)"
    fi
    
    # Run integration tests
    run_integration_tests
    
    print_success "Integration test runner completed successfully!"
}

# Run main function with all arguments
main "$@"
