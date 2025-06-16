#!/bin/bash

# Quick Redis Connection Test Script
# This script tests Redis connectivity with enhanced error handling

set -e

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

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to test Redis with enhanced debugging
test_redis_connection() {
    local host=${1:-127.0.0.1}
    local port=${2:-6379}
    local max_attempts=${3:-60}
    local attempt=1

    print_status "Testing Redis connection to $host:$port..."
    
    while [ $attempt -le $max_attempts ]; do
        print_status "Attempt $attempt/$max_attempts..."
        
        # Test 1: Check if port is open
        if command -v nc >/dev/null 2>&1; then
            if nc -z "$host" "$port" 2>/dev/null; then
                print_success "Port $port is open"
            else
                print_warning "Port $port is not open yet"
                sleep 2
                attempt=$((attempt + 1))
                continue
            fi
        else
            print_warning "netcat (nc) not available, skipping port check"
        fi
        
        # Test 2: Try Redis ping
        if command -v redis-cli >/dev/null 2>&1; then
            local ping_result
            ping_result=$(redis-cli -h "$host" -p "$port" ping 2>&1)
            local ping_exit_code=$?
            
            if [ $ping_exit_code -eq 0 ] && [ "$ping_result" = "PONG" ]; then
                print_success "Redis PING successful: $ping_result"
                
                # Test 3: Basic Redis operations
                print_status "Testing basic Redis operations..."
                if redis-cli -h "$host" -p "$port" set test_key "test_value" >/dev/null 2>&1; then
                    local get_result
                    get_result=$(redis-cli -h "$host" -p "$port" get test_key 2>/dev/null)
                    if [ "$get_result" = "test_value" ]; then
                        print_success "Redis SET/GET operations working"
                        redis-cli -h "$host" -p "$port" del test_key >/dev/null 2>&1
                        return 0
                    else
                        print_error "Redis GET operation failed"
                    fi
                else
                    print_error "Redis SET operation failed"
                fi
            else
                print_warning "Redis PING failed: $ping_result (exit code: $ping_exit_code)"
            fi
        else
            print_error "redis-cli not available"
            return 1
        fi
        
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_error "Redis connection test failed after $max_attempts attempts"
    return 1
}

# Function to test MySQL connection
test_mysql_connection() {
    local host=${1:-127.0.0.1}
    local port=${2:-3306}
    local user=${3:-root}
    local password=${4:-testpassword}
    local max_attempts=${5:-60}
    local attempt=1

    print_status "Testing MySQL connection to $host:$port..."
    
    while [ $attempt -le $max_attempts ]; do
        print_status "Attempt $attempt/$max_attempts..."
        
        # Test 1: Check if port is open
        if command -v nc >/dev/null 2>&1; then
            if nc -z "$host" "$port" 2>/dev/null; then
                print_success "Port $port is open"
            else
                print_warning "Port $port is not open yet"
                sleep 2
                attempt=$((attempt + 1))
                continue
            fi
        fi
        
        # Test 2: Try MySQL ping
        if command -v mysqladmin >/dev/null 2>&1; then
            if mysqladmin ping -h"$host" -P"$port" -u"$user" -p"$password" --silent 2>/dev/null; then
                print_success "MySQL ping successful"
                
                # Test 3: Try basic MySQL query
                if mysql -h"$host" -P"$port" -u"$user" -p"$password" -e "SELECT 1;" >/dev/null 2>&1; then
                    print_success "MySQL basic query working"
                    return 0
                else
                    print_warning "MySQL basic query failed"
                fi
            else
                print_warning "MySQL ping failed"
            fi
        else
            print_error "mysqladmin not available"
            return 1
        fi
        
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_error "MySQL connection test failed after $max_attempts attempts"
    return 1
}

# Main function
main() {
    local redis_host=${REDIS_HOST:-127.0.0.1}
    local redis_port=${REDIS_PORT:-6379}
    local mysql_host=${MYSQL_HOST:-127.0.0.1}
    local mysql_port=${MYSQL_PORT:-3306}
    local mysql_user=${MYSQL_USER:-root}
    local mysql_password=${MYSQL_PASSWORD:-test_password}
    
    print_status "Starting database connectivity tests..."
    
    # Test Redis
    if test_redis_connection "$redis_host" "$redis_port"; then
        print_success "Redis connectivity test passed"
    else
        print_error "Redis connectivity test failed"
        exit 1
    fi
    
    # Test MySQL
    if test_mysql_connection "$mysql_host" "$mysql_port" "$mysql_user" "$mysql_password"; then
        print_success "MySQL connectivity test passed"
    else
        print_error "MySQL connectivity test failed"
        exit 1
    fi
    
    print_success "All database connectivity tests passed!"
}

# Run main function if script is executed directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi
