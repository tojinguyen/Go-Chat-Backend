# Integration Tests with Real Database

This directory contains integration tests that run against real MySQL and Redis instances, providing comprehensive testing of the entire application stack.

## Overview

Unlike unit tests that use mocks, these integration tests:
- ✅ Use real MySQL database connections
- ✅ Use real Redis connections  
- ✅ Test actual SQL queries and Redis operations
- ✅ Verify data persistence and retrieval
- ✅ Test transaction handling
- ✅ Validate database constraints and relationships
- ✅ Run automatically on GitHub Actions with CI/CD

## Test Structure

```
tests/integration/
├── setup_test.go                    # Test setup and teardown
├── auth_integration_test.go         # Auth use case integration tests
├── repository_integration_test.go   # Repository layer integration tests
└── redis_integration_test.go        # Redis operations integration tests
```

## Running Integration Tests

### Prerequisites

1. **Go 1.24.2** or later
2. **Docker and Docker Compose** (for local testing)
3. **MySQL client tools** (for database operations)
4. **Redis CLI tools** (for Redis operations)

### Method 1: Using Scripts (Recommended)

#### On Windows (PowerShell):
```powershell
# Run integration tests with Docker
.\scripts\run-integration-tests.ps1

# Run with existing databases
.\scripts\run-integration-tests.ps1 -NoDocker

# Skip migrations (if already run)
.\scripts\run-integration-tests.ps1 -SkipMigrations

# Show help
.\scripts\run-integration-tests.ps1 -Help
```

#### On Linux/macOS (Bash):
```bash
# Make script executable
chmod +x scripts/run-integration-tests.sh

# Run integration tests with Docker
./scripts/run-integration-tests.sh

# Run with existing databases
./scripts/run-integration-tests.sh --no-docker

# Skip migrations (if already run)
./scripts/run-integration-tests.sh --skip-migrations

# Show help
./scripts/run-integration-tests.sh --help
```

### Method 2: Using Makefile

```bash
# Setup test databases and run integration tests
make test-integration-docker

# Run integration tests (requires manual database setup)
make test-integration

# Setup test database only
make setup-test-db

# Clean test database
make clean-test-db

# Run all tests (unit + integration)
make test-all
```

### Method 3: Manual Setup

1. **Start test databases:**
   ```bash
   docker-compose -f docker-compose.test.yml up -d
   ```

2. **Load test environment:**
   ```bash
   export $(cat .env.test | grep -v '^#' | xargs)
   ```

3. **Run migrations:**
   ```bash
   go run ./cmd/migrate/main.go
   ```

4. **Run integration tests:**
   ```bash
   go test -v -tags=integration ./tests/integration/...
   ```

5. **Stop test databases:**
   ```bash
   docker-compose -f docker-compose.test.yml down -v
   ```

## GitHub Actions Integration

Integration tests run automatically on GitHub Actions for:
- ✅ Push to `main`, `master`, or `develop` branches
- ✅ Pull requests to `main`, `master`, or `develop` branches

The CI pipeline:
1. Sets up MySQL 8.0 and Redis services
2. Waits for databases to be ready
3. Runs database migrations
4. Executes all integration tests
5. Generates and uploads coverage reports
6. Cleans up test data

### GitHub Actions Workflow

See `.github/workflows/integration-tests.yml` for the complete CI configuration.

## Test Configuration

### Environment Variables

Integration tests use the following test-specific environment variables:

```env
# Database Configuration
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=testpassword
MYSQL_DATABASE=gochat_test

# Redis Configuration  
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration (Test keys only!)
ACCESS_TOKEN_SECRET_KEY=test-access-secret-key-for-testing-only-123456789
REFRESH_TOKEN_SECRET_KEY=test-refresh-secret-key-for-testing-only-987654321
```

See `.env.test` for the complete test configuration.

### Database Setup

Integration tests use a separate test database:
- **Database Name:** `gochat_test`
- **Port:** 3307 (Docker) or 3306 (GitHub Actions)
- **Data:** Automatically cleaned between test runs

## Test Coverage

### Auth Integration Tests (`auth_integration_test.go`)

Tests the complete auth flow with real database:

- ✅ **Login Integration**
  - Successful login with real user data
  - User not found scenarios
  - Invalid password validation
  - Refresh token storage in Redis

- ✅ **Registration Integration**
  - Complete registration flow
  - Email existence validation
  - Verification code creation in database

- ✅ **Verification Integration**
  - Email verification with real database
  - User account creation after verification
  - Expired/invalid code handling

### Repository Integration Tests (`repository_integration_test.go`)

Tests database operations directly:

- ✅ **Account Repository**
  - Create and retrieve user accounts
  - Email existence checks
  - Database constraint validation
  - Find by ID and email operations

- ✅ **Verification Repository**
  - Verification code CRUD operations
  - Status updates and deletions
  - Email-based lookups

### Redis Integration Tests (`redis_integration_test.go`)

Tests Redis operations:

- ✅ **Basic Operations**
  - Set and get string values
  - Key expiration handling
  - Key deletion operations

- ✅ **Real-world Scenarios**
  - Refresh token storage and retrieval
  - Concurrent operations
  - Connection health checks

## Test Data Management

### Automatic Cleanup

Integration tests automatically clean up test data:

```go
func cleanupTestData() {
    // Cleans all test tables
    tables := []string{
        "verification_register_code",
        "friend_requests", 
        "friend_ships",
        "message",
        "chat_rooms",
        "account",
    }
    
    // Clean Redis test data
    TestRedis.FlushDB(ctx)
}
```

### Isolation

Each test runs in isolation:
- Database is cleaned before each test run
- Redis is flushed between tests
- No cross-test data contamination

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```
   Solution: Ensure Docker containers are running and healthy
   Check: docker-compose -f docker-compose.test.yml ps
   ```

2. **Migration Errors**
   ```
   Solution: Verify migration files in migrations/mysql/
   Check: Database user has necessary permissions
   ```

3. **Redis Connection Failed**
   ```
   Solution: Ensure Redis container is running
   Check: redis-cli -h 127.0.0.1 -p 6379 ping
   ```

4. **Port Conflicts**
   ```
   Solution: Stop services using ports 3307 and 6380
   Or modify docker-compose.test.yml port mappings
   ```

### Debug Mode

Run tests with verbose output:
```bash
go test -v -tags=integration ./tests/integration/... -args -test.v
```

### Database Inspection

Connect to test database during development:
```bash
mysql -h 127.0.0.1 -P 3307 -u root -ptestpassword gochat_test
```

## Performance Considerations

- Integration tests are slower than unit tests
- Use parallel execution when possible: `go test -parallel=4`
- Database connections are pooled and reused
- Consider running integration tests separately from unit tests

## Best Practices

1. **Test Isolation**: Each test should be independent
2. **Data Cleanup**: Always clean test data between runs
3. **Real Scenarios**: Test actual user workflows
4. **Error Cases**: Test both success and failure scenarios
5. **Performance**: Monitor test execution time
6. **CI/CD**: Ensure tests pass consistently in CI environment

## Extending Integration Tests

To add new integration tests:

1. Create new test file with `// +build integration` tag
2. Use `TestMain` for setup if needed
3. Follow existing patterns for database/Redis operations
4. Add test to appropriate GitHub Actions workflow
5. Update this README with new test coverage

## Security Notes

- Test environment uses **mock credentials only**
- Never use production secrets in test configuration
- Test database is isolated and temporary
- All test data is automatically cleaned up

---

For more information about the testing strategy, see:
- [Unit Tests Documentation](../internal/usecase/auth/README_TESTS.md)
- [GitHub Actions Configuration](../.github/workflows/integration-tests.yml)
- [Docker Compose Test Setup](../docker-compose.test.yml)
