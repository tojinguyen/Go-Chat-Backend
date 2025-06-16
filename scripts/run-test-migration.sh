#!/bin/bash

# Bash script to run test migrations with proper environment variables from .env.test
# Usage: ./scripts/run-test-migration.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

print_status "Setting up test environment variables from .env.test..."

# Load test environment variables from .env.test
if [ -f ".env.test" ]; then
    export $(grep -v '^#' .env.test | xargs)
    print_success "Environment variables loaded from .env.test"
else
    print_error ".env.test file not found!"
    exit 1
fi

# Check if test database containers are running
print_status "Checking if test database containers are running..."

mysql_container=$(docker ps --filter "name=realtime_chat_app_mysql_test" --filter "status=running" --quiet)
redis_container=$(docker ps --filter "name=realtime_chat_app_redis_test" --filter "status=running" --quiet)

if [[ -z "$mysql_container" || -z "$redis_container" ]]; then
    print_status "Starting test database containers..."
    docker-compose -f docker-compose.test.yml up -d mysql-test redis-test --wait
    
    if [[ $? -ne 0 ]]; then
        print_error "Failed to start test containers"
        exit 1
    fi
    
    print_status "Waiting for containers to be ready..."
    sleep 10
fi

print_success "Test containers are running"

# Run the migration using Go migrate command
print_status "Running test migration..."

# Use the migrate command from Go
if go run ./cmd/migrate/main.go; then
    print_success "Test migration completed successfully!"
else
    print_error "Test migration failed!"
    exit 1
fi
