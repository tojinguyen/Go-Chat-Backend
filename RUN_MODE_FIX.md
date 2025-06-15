# Quick Fix: RUN_MODE Environment Variable Error

If you see this error when running `go run ./cmd/migrate/main.go`:
```
Failed to load environment: value for this field is required [RUN_MODE]
```

## 🚀 Quick Solutions:

### Windows (PowerShell):
```powershell
.\scripts\run-test-migration.ps1
```

### Unix/Linux/macOS:
```bash
./scripts/run-test-migration.sh
```

### Any Platform (using Make):
```bash
make migrate-test
```

### Manual (using Goose directly):
```bash
# Start test containers first
docker-compose -f docker-compose.test.yml up -d mysql-test redis-test --wait

# Set environment and run migration
export GOOSE_DRIVER=mysql
export GOOSE_DBSTRING="testuser:testpassword@tcp(localhost:3307)/gochat_test?parseTime=true"
goose -dir migrations/mysql up
```

**Windows PowerShell manual:**
```powershell
# Start test containers first
docker-compose -f docker-compose.test.yml up -d mysql-test redis-test --wait

# Set environment and run migration
$env:GOOSE_DRIVER="mysql"
$env:GOOSE_DBSTRING="testuser:testpassword@tcp(localhost:3307)/gochat_test?parseTime=true"
goose -dir migrations/mysql up
```

## ✅ What This Fixes:
- Sets required environment variables (RUN_MODE, database configs, etc.)
- Uses proper Goose migration tool instead of custom runner
- Automatically starts test database containers
- Works with the existing Goose migration files

## 📖 Full Documentation:
See `docs/TEST_MIGRATION_GUIDE.md` for complete details and troubleshooting.
