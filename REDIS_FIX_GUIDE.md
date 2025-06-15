# Database Connection Troubleshooting Guide

## Vấn đề đã được khắc phục

### 1. Lỗi Redis Service bị đứng trong GitHub Actions

**Nguyên nhân:**
- Redis image `redis:8.0-rc1` không ổn định trong môi trường GitHub Actions
- Timeout quá thấp (30 attempts = 60 giây) 
- Health check interval quá dài
- Thiếu kiểm tra port connectivity trước khi test Redis commands

**Giải pháp đã áp dụng:**

#### A. Thay đổi Redis Image
- Đổi từ `redis:8.0-rc1` sang `redis:7.2.4-alpine` (version ổn định hơn)
- Thêm Redis configuration tối ưu: `--appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru`

#### B. Cải thiện Health Check
- Giảm interval từ 10s xuống 5s
- Giảm timeout từ 5s xuống 3s
- Tăng retries từ 5 lên 10
- Thêm `start_period` để Redis có thời gian khởi động

#### C. Tăng Timeout cho Wait Scripts
- Tăng max attempts từ 30 lên 60 (từ 60s lên 120s)
- Thêm port connectivity check trước khi test Redis commands
- Cải thiện error handling và debugging output

#### D. Thêm Debugging Tools
- Cài đặt mysql-client và redis-tools trong GitHub Actions
- Thêm connectivity test scripts với chi tiết debug
- Thêm service status checking

### 2. Các file đã được cập nhật

#### GitHub Actions Workflow (`.github/workflows/integration-tests.yml`)
```yaml
# Cải thiện services health check
services:
  redis:
    image: redis:7.2.4-alpine  # Stable version
    options: >-
      --health-cmd="redis-cli ping"
      --health-interval=5s      # Faster check
      --health-timeout=3s       # Shorter timeout
      --health-retries=10       # More retries
      --health-start-period=10s # Grace period

# Thêm enhanced wait scripts với timeout
- name: Wait for Redis to be ready
  run: |
    max_attempts=60  # 2 minutes instead of 1
    # ... enhanced wait logic
```

#### Docker Compose Test (docker-compose.test.yml)
```yaml
redis-test:
  image: redis:7.2.4-alpine  # Stable version
  command: redis-server --appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru
  healthcheck:
    test: ["CMD", "redis-cli", "ping"]
    interval: 5s      # Faster checks
    timeout: 3s       # Shorter timeout
    retries: 10       # More retries
    start_period: 10s # Grace period
```

#### PowerShell Script (scripts/run-integration-tests.ps1)
```powershell
function Wait-ForRedis {
    param([int]$MaxAttempts = 60)  # Increased from 30
    
    # Test port connectivity first
    $portTest = Test-NetConnection -ComputerName $Host -Port $Port -InformationLevel Quiet
    
    # Then test Redis ping
    if ($portTest) {
        $result = redis-cli -h $Host -p $Port ping
        if ($LASTEXITCODE -eq 0 -and $result -eq "PONG") {
            # Success
        }
    }
}
```

#### Bash Script (scripts/run-integration-tests.sh)
```bash
wait_for_redis() {
    local max_attempts=60  # Increased from 30
    
    # Check port connectivity first
    if nc -z "$host" "$port" 2>/dev/null; then
        # Then test Redis ping
        if redis-cli -h "$host" -p "$port" ping >/dev/null 2>&1; then
            # Success
        fi
    fi
}
```

### 3. Debugging Tools mới

#### Test Database Connectivity Scripts
- `scripts/test-db-connectivity.sh` - Enhanced bash testing script
- `scripts/test-db-connectivity.ps1` - Enhanced PowerShell testing script

Các script này thực hiện:
1. Port connectivity check
2. Database ping test
3. Basic operations test (SET/GET cho Redis, SELECT 1 cho MySQL)
4. Detailed error reporting

#### Fallback Configuration
- `docker-compose.test.fallback.yml` - Backup configuration với Redis stable version

### 4. Cách sử dụng

#### Local Testing
```bash
# Test trên Linux/Mac
./scripts/run-integration-tests.sh

# Test trên Windows
.\scripts\run-integration-tests.ps1

# Test connectivity riêng biệt
./scripts/test-db-connectivity.sh
.\scripts\test-db-connectivity.ps1
```

#### Troubleshooting trong GitHub Actions
Nếu vẫn gặp vấn đề:

1. **Kiểm tra logs**:
   - Service startup logs
   - Health check status
   - Connectivity test results

2. **Thử fallback configuration**:
   ```yaml
   # Trong workflow, thay đổi docker-compose file
   docker-compose -f docker-compose.test.fallback.yml up -d
   ```

3. **Tăng timeout thêm**:
   ```bash
   # Trong wait scripts, tăng max_attempts
   max_attempts=90  # 3 minutes
   ```

### 5. Monitoring và Alerts

Script bây giờ sẽ output detailed logs:
- Port connectivity status
- Service health check results
- Timing information
- Error details with exit codes

Tất cả thay đổi này sẽ giúp:
1. ✅ Giảm tỷ lệ fail do Redis timeout
2. ✅ Cải thiện debugging khi có lỗi
3. ✅ Tăng reliability của integration tests
4. ✅ Faster feedback khi có vấn đề

## Kết luận

Các thay đổi trên sẽ khắc phục vấn đề Redis service bị đứng trong GitHub Actions bằng cách:
- Sử dụng Redis image ổn định hơn
- Tối ưu health check timing
- Tăng timeout và retries
- Thêm better error handling và debugging
- Port connectivity checking trước database commands
