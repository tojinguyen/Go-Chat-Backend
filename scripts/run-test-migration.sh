#!/bin/bash

# Bash script to run test migrations with proper environment variables
# Usage: ./scripts/run-test-migration.sh

echo "🔧 Setting up test environment variables..."

# Set test environment variables
export RUN_MODE="test"
export PORT="8080"
export CORS_ALLOW_ORIGIN="http://localhost:3000"

# MySQL Test Config (assuming Docker containers are running)
export MYSQL_HOST="localhost"
export MYSQL_PORT="3307"
export MYSQL_USER="testuser"
export MYSQL_PASSWORD="testpassword"
export MYSQL_DATABASE="gochat_test"
export MYSQL_SSL_MODE="disable"
export MYSQL_MIGRATE_MODE="auto"

# Required environment variables (dummy values for migration)
export ACCESS_TOKEN_SECRET_KEY="test_access_secret_key_for_migration"
export ACCESS_TOKEN_EXPIRE_MINUTES="60"
export REFRESH_TOKEN_SECRET_KEY="test_refresh_secret_key_for_migration"
export REFRESH_TOKEN_EXPIRE_MINUTES="1440"

export FRONTEND_URI="http://localhost:3000"
export FRONTEND_PORT="3000"

export EMAIL_HOST="smtp.test.com"
export EMAIL_PORT="587"
export EMAIL_USER="test@test.com"
export EMAIL_PASS="testpass"
export EMAIL_FROM="test@test.com"
export EMAIL_NAME="Test Chat App"

export VERIFICATION_CODE_LENGTH="6"
export VERIFICATION_CODE_EXPIRE_MINUTES="5"

export CLOUDINARY_CLOUD_NAME="test_cloud"
export CLOUDINARY_API_KEY="test_api_key"
export CLOUDINARY_API_SECRET="test_api_secret"

export REDIS_HOST="localhost"
export REDIS_PORT="6380"
export REDIS_PASSWORD=""
export REDIS_DB="0"

export KAFKA_BROKERS="localhost:9092"
export KAFKA_CHAT_TOPIC="test_chat_topic"
export KAFKA_CONSUMER_GROUP="test_chat_group"
export KAFKA_ENABLED="false"

echo "✅ Environment variables set"

# Check if test database containers are running
echo "🔍 Checking if test database containers are running..."

mysql_container=$(docker ps --filter "name=realtime_chat_app_mysql_test" --filter "status=running" --quiet)
redis_container=$(docker ps --filter "name=realtime_chat_app_redis_test" --filter "status=running" --quiet)

if [[ -z "$mysql_container" || -z "$redis_container" ]]; then
    echo "🚀 Starting test database containers..."
    docker-compose -f docker-compose.test.yml up -d mysql-test redis-test --wait
    
    if [[ $? -ne 0 ]]; then
        echo "❌ Failed to start test containers"
        exit 1
    fi
    
    echo "⏳ Waiting for containers to be ready..."
    sleep 10
fi

echo "✅ Test containers are running"

# Run the migration using Goose
echo "🗄️  Running test migration using Goose..."
export GOOSE_DRIVER="mysql"
export GOOSE_DBSTRING="testuser:testpassword@tcp(localhost:3307)/gochat_test?parseTime=true"

goose -dir migrations/mysql up

if [[ $? -eq 0 ]]; then
    echo "✅ Test migration completed successfully!"
else
    echo "❌ Test migration failed!"
    exit 1
fi
