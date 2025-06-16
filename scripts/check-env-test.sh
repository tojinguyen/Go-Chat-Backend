#!/bin/bash

# Script to verify .env.test file exists and has all required variables
# Usage: ./scripts/check-env-test.sh

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

# Required environment variables for the application
REQUIRED_VARS=(
    "RUN_MODE"
    "PORT"
    "CORS_ALLOW_ORIGIN"
    "MYSQL_HOST"
    "MYSQL_PORT"
    "MYSQL_USER"
    "MYSQL_PASSWORD"
    "MYSQL_DATABASE"
    "ACCESS_TOKEN_SECRET_KEY"
    "ACCESS_TOKEN_EXPIRE_MINUTES"
    "REFRESH_TOKEN_SECRET_KEY"
    "REFRESH_TOKEN_EXPIRE_MINUTES"
    "FRONTEND_URI"
    "FRONTEND_PORT"
    "EMAIL_HOST"
    "EMAIL_PORT"
    "EMAIL_USER"
    "EMAIL_PASS"
    "EMAIL_FROM"
    "EMAIL_NAME"
    "CLOUDINARY_CLOUD_NAME"
    "CLOUDINARY_API_KEY"
    "CLOUDINARY_API_SECRET"
    "REDIS_HOST"
    "REDIS_PORT"
    "REDIS_DB"
    "KAFKA_BROKERS"
    "KAFKA_CHAT_TOPIC"
    "KAFKA_CONSUMER_GROUP"
)

check_env_test_file() {
    print_status "Checking .env.test file..."
    
    if [ ! -f ".env.test" ]; then
        print_error ".env.test file not found!"
        print_status "Please create .env.test file with test environment variables"
        return 1
    fi
    
    print_success ".env.test file exists"
    
    # Load variables from .env.test
    export $(grep -v '^#' .env.test | xargs)
    
    print_status "Checking required environment variables..."
    
    local missing_vars=()
    
    for var in "${REQUIRED_VARS[@]}"; do
        if [ -z "${!var}" ]; then
            missing_vars+=("$var")
        fi
    done
    
    if [ ${#missing_vars[@]} -gt 0 ]; then
        print_error "Missing required environment variables:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        return 1
    fi
    
    print_success "All required environment variables are present"
    
    # Validate some critical values
    if [ "$RUN_MODE" != "test" ]; then
        print_warning "RUN_MODE is not set to 'test' (current: $RUN_MODE)"
    fi
    
    if [ "$MYSQL_DATABASE" = "chat_app_db" ]; then
        print_warning "MySQL database name is same as production (should be test database)"
    fi
    
    print_success ".env.test validation completed"
    return 0
}

# Show environment variables (for debugging)
show_env_vars() {
    print_status "Current environment variables from .env.test:"
    echo "  RUN_MODE: $RUN_MODE"
    echo "  MYSQL_HOST: $MYSQL_HOST"
    echo "  MYSQL_PORT: $MYSQL_PORT"
    echo "  MYSQL_USER: $MYSQL_USER"
    echo "  MYSQL_DATABASE: $MYSQL_DATABASE"
    echo "  REDIS_HOST: $REDIS_HOST"
    echo "  REDIS_PORT: $REDIS_PORT"
    echo "  FRONTEND_URI: $FRONTEND_URI"
}

main() {
    print_status "Environment test file checker starting..."
    
    if check_env_test_file; then
        if [ "$1" = "--show-vars" ]; then
            show_env_vars
        fi
        print_success "Environment validation passed!"
        return 0
    else
        print_error "Environment validation failed!"
        return 1
    fi
}

# Run main function with all arguments
main "$@"
