# PowerShell script to run test migrations with proper environment variables
# Usage: .\scripts\run-test-migration.ps1

Write-Host "Setting up test environment variables..." -ForegroundColor Cyan

# Set test environment variables
$env:RUN_MODE = "test"
$env:PORT = "8080"
$env:CORS_ALLOW_ORIGIN = "http://localhost:3000"

# MySQL Test Config (assuming Docker containers are running)
$env:MYSQL_HOST = "localhost"
$env:MYSQL_PORT = "3307"
$env:MYSQL_USER = "testuser"
$env:MYSQL_PASSWORD = "testpassword"
$env:MYSQL_DATABASE = "gochat_test"
$env:MYSQL_SSL_MODE = "disable"
$env:MYSQL_MIGRATE_MODE = "auto"

# Required environment variables (dummy values for migration)
$env:ACCESS_TOKEN_SECRET_KEY = "test_access_secret_key_for_migration"
$env:ACCESS_TOKEN_EXPIRE_MINUTES = "60"
$env:REFRESH_TOKEN_SECRET_KEY = "test_refresh_secret_key_for_migration"
$env:REFRESH_TOKEN_EXPIRE_MINUTES = "1440"

$env:FRONTEND_URI = "http://localhost:3000"
$env:FRONTEND_PORT = "3000"

$env:EMAIL_HOST = "smtp.test.com"
$env:EMAIL_PORT = "587"
$env:EMAIL_USER = "test@test.com"
$env:EMAIL_PASS = "testpass"
$env:EMAIL_FROM = "test@test.com"
$env:EMAIL_NAME = "Test Chat App"

$env:VERIFICATION_CODE_LENGTH = "6"
$env:VERIFICATION_CODE_EXPIRE_MINUTES = "5"

$env:CLOUDINARY_CLOUD_NAME = "test_cloud"
$env:CLOUDINARY_API_KEY = "test_api_key"
$env:CLOUDINARY_API_SECRET = "test_api_secret"

$env:REDIS_HOST = "localhost"
$env:REDIS_PORT = "6380"
$env:REDIS_PASSWORD = ""
$env:REDIS_DB = "0"

$env:KAFKA_BROKERS = "localhost:9092"
$env:KAFKA_CHAT_TOPIC = "test_chat_topic"
$env:KAFKA_CONSUMER_GROUP = "test_chat_group"
$env:KAFKA_ENABLED = "false"

Write-Host "Environment variables set" -ForegroundColor Green

# Check if test database containers are running
Write-Host "Checking if test database containers are running..." -ForegroundColor Yellow

$mysqlContainer = docker ps --filter "name=realtime_chat_app_mysql_test" --filter "status=running" --quiet
$redisContainer = docker ps --filter "name=realtime_chat_app_redis_test" --filter "status=running" --quiet

if (-not $mysqlContainer -or -not $redisContainer) {
    Write-Host "Starting test database containers..." -ForegroundColor Yellow
    docker-compose -f docker-compose.test.yml up -d mysql-test redis-test --wait
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: Failed to start test containers" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "Waiting for containers to be ready..." -ForegroundColor Yellow
    Start-Sleep -Seconds 10
}

Write-Host "Test containers are running" -ForegroundColor Green

# Run the migration using Goose
Write-Host "Running test migration using Goose..." -ForegroundColor Cyan
$env:GOOSE_DRIVER = "mysql"
$env:GOOSE_DBSTRING = "testuser:testpassword@tcp(localhost:3307)/gochat_test?parseTime=true"

goose -dir migrations/mysql up

if ($LASTEXITCODE -eq 0) {
    Write-Host "SUCCESS: Test migration completed successfully!" -ForegroundColor Green
} else {
    Write-Host "ERROR: Test migration failed!" -ForegroundColor Red
    exit 1
}
