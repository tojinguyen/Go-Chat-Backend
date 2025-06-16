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

# Function to debug Docker containers
debug_containers() {
    print_status "Docker container status:"
    docker ps --filter "name=realtime_chat_app" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" || true
    
    print_status "Docker container logs (MySQL):"
    docker logs realtime_chat_app_mysql_test --tail 10 2>/dev/null || print_warning "MySQL container not found or not running"
    
    print_status "Docker container logs (Redis):"
    docker logs realtime_chat_app_redis_test --tail 10 2>/dev/null || print_warning "Redis container not found or not running"
}

# Function to wait for MySQL to be ready
wait_for_mysql() {
    local host=${1:-127.0.0.1}
    local port=${2:-3306}
    local user=${3:-root}
    local password=${4:-test_password}
    local max_attempts=60
    local attempt=1

    print_status "Waiting for MySQL to be ready at $host:$port with user '$user'..."
    
    while [ $attempt -le $max_attempts ]; do
        print_status "Attempt $attempt/$max_attempts: Checking MySQL..."
        
        # Check if port is open first
        if nc -z "$host" "$port" 2>/dev/null; then
            print_status "Port $port is open, trying MySQL ping..."
            
            # Try MySQL ping command with detailed error output
            local ping_result
            ping_result=$(mysqladmin ping -h"$host" -P"$port" -u"$user" -p"$password" --silent 2>&1)
            local ping_exit_code=$?
            
            if [ $ping_exit_code -eq 0 ]; then
                print_success "MySQL is ready!"
                return 0
            else
                print_warning "MySQL ping failed (exit code: $ping_exit_code): $ping_result"
                
                # Try different approaches if root fails
                if [ "$user" = "root" ] && [ $attempt -gt 10 ]; then
                    print_status "Trying alternative MySQL connection methods..."
                    
                    # Try without password for root
                    if mysqladmin ping -h"$host" -P"$port" -u"$user" --silent 2>/dev/null; then
                        print_warning "MySQL accessible without password, updating connection"
                        password=""
                        print_success "MySQL is ready (no password)!"
                        return 0
                    fi
                    
                    # Try with test_user if available from env
                    if [ -n "$MYSQL_USER" ] && [ "$MYSQL_USER" != "root" ]; then
                        print_status "Trying with configured user: $MYSQL_USER"
                        if mysqladmin ping -h"$host" -P"$port" -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" --silent 2>/dev/null; then
                            print_success "MySQL is ready with configured user!"
                            return 0
                        fi
                    fi
                fi
            fi
        else
            print_warning "Port $port is not open yet"
        fi
        
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_error "MySQL failed to start within expected time"
    print_error "Final connection attempt details:"
    print_error "  Host: $host"
    print_error "  Port: $port" 
    print_error "  User: $user"
    print_error "  Password length: ${#password}"
    
    # Final diagnostic attempt
    print_status "Running final MySQL diagnostic..."
    mysqladmin ping -h"$host" -P"$port" -u"$user" -p"$password" --verbose 2>&1 || true
    
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
    print_status "Stopping any existing test containers..."
    docker-compose -f docker-compose.test.yml down -v >/dev/null 2>&1 || true
    
    # Start test databases
    print_status "Starting test database containers..."
    if docker-compose -f docker-compose.test.yml up -d mysql-test redis-test; then
        print_success "Test database containers started"
        
        # Give containers time to initialize
        print_status "Waiting for containers to initialize..."
        sleep 5
        
        # Get MySQL credentials from docker-compose or environment
        local mysql_root_password="${MYSQL_PASSWORD:-test_password}"
        local mysql_user="${MYSQL_USER:-test_user}"
        local mysql_host="127.0.0.1"
        local mysql_port="3307"  # Port mapped in docker-compose.test.yml
        local redis_host="127.0.0.1"
        local redis_port="6380"  # Port mapped in docker-compose.test.yml
        
        print_status "Waiting for MySQL on $mysql_host:$mysql_port with user '$mysql_user' and password length: ${#mysql_root_password}"
        
        # Wait for databases to be ready using mapped ports
        wait_for_mysql "$mysql_host" "$mysql_port" "$mysql_user" "$mysql_root_password"
        wait_for_redis "$redis_host" "$redis_port"
        
        # Update environment for test databases (use mapped ports)
        export MYSQL_HOST="$mysql_host"
        export MYSQL_PORT="$mysql_port"
        export MYSQL_USER="$mysql_user"
        export MYSQL_PASSWORD="$mysql_root_password"
        export REDIS_HOST="$redis_host"
        export REDIS_PORT="$redis_port"
        
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
    
    # Debug containers before cleanup if there were issues
    if [ $? -ne 0 ]; then
        print_status "Script failed, showing debug information..."
        debug_containers
    fi
    
    stop_test_databases
}

# Main function
main() {
    local use_docker=true
    local skip_migrations=false
    local debug_mode=false
    
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
            --debug)
                debug_mode=true
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --no-docker        Use existing databases instead of starting Docker containers"
                echo "  --skip-migrations  Skip running database migrations"
                echo "  --debug           Enable debug mode with container logs"
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
        
        # Show debug info if requested
        if [ "$debug_mode" = true ]; then
            debug_containers
        fi
    else
        print_status "Using existing databases (--no-docker specified)"
        
        # When using existing databases, use the environment variables as-is
        # This supports both local development and CI environments
        wait_for_mysql "${MYSQL_HOST:-127.0.0.1}" "${MYSQL_PORT:-3306}" "${MYSQL_USER:-root}" "${MYSQL_PASSWORD:-test_password}"
        wait_for_redis "${REDIS_HOST:-127.0.0.1}" "${REDIS_PORT:-6379}"
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
