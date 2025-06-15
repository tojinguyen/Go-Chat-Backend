# Integration Test Runner Script for Windows PowerShell
# This script sets up and runs integration tests with real databases

param(
    [switch]$NoDocker,
    [switch]$SkipMigrations,
    [switch]$Help
)

# Colors for output
$Green = "Green"
$Red = "Red"
$Yellow = "Yellow"
$Blue = "Blue"

function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Red
}

function Test-Command {
    param([string]$Command)
    try {
        if (Get-Command $Command -ErrorAction SilentlyContinue) {
            return $true
        }
    }
    catch {
        return $false
    }
    return $false
}

function Wait-ForMySQL {
    param(
        [string]$Host = "127.0.0.1",
        [int]$Port = 3306,
        [string]$User = "root",
        [string]$Password = "testpassword",
        [int]$MaxAttempts = 60
    )
    
    Write-Status "Waiting for MySQL to be ready..."
    
    for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
        try {
            # Check if port is open first
            $portTest = Test-NetConnection -ComputerName $Host -Port $Port -InformationLevel Quiet -WarningAction SilentlyContinue
            if ($portTest) {
                # If port is open, try MySQL ping command
                $result = mysqladmin ping -h $Host -P $Port -u $User -p$Password --silent 2>$null
                if ($LASTEXITCODE -eq 0) {
                    Write-Success "MySQL is ready!"
                    return $true
                }
            }
        }
        catch {
            # Continue trying
        }
        
        Write-Host "Attempt $attempt/$MaxAttempts`: MySQL not ready yet..."
        Start-Sleep -Seconds 2
    }
    
    Write-Error "MySQL failed to start within expected time"
    return $false
}

function Wait-ForRedis {
    param(
        [string]$Host = "127.0.0.1",
        [int]$Port = 6379,
        [int]$MaxAttempts = 60
    )
    
    Write-Status "Waiting for Redis to be ready..."
    
    for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
        try {
            # Use Test-NetConnection to check if port is open first
            $portTest = Test-NetConnection -ComputerName $Host -Port $Port -InformationLevel Quiet -WarningAction SilentlyContinue
            if ($portTest) {
                # If port is open, try Redis ping command
                $result = redis-cli -h $Host -p $Port ping 2>$null
                if ($LASTEXITCODE -eq 0 -and $result -eq "PONG") {
                    Write-Success "Redis is ready!"
                    return $true
                }
            }
        }
        catch {
            # Continue trying
        }
        
        Write-Host "Attempt $attempt/$MaxAttempts`: Redis not ready yet..."
        Start-Sleep -Seconds 2
    }
    
    Write-Error "Redis failed to start within expected time"
    return $false
}

function Setup-TestEnvironment {
    Write-Status "Setting up test environment..."
    
    if (Test-Path ".env.test") {
        $envContent = Get-Content ".env.test" | Where-Object { $_ -notmatch "^#" -and $_.Trim() -ne "" }
        foreach ($line in $envContent) {
            if ($line -match "^([^=]+)=(.*)$") {
                $key = $matches[1]
                $value = $matches[2]
                [System.Environment]::SetEnvironmentVariable($key, $value, "Process")
            }
        }
        Write-Success "Loaded test environment variables"
    }
    else {
        Write-Warning ".env.test file not found, using default environment"
    }
}

function Invoke-Migrations {
    Write-Status "Running database migrations..."
    
    try {
        go run ./cmd/migrate/main.go
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Database migrations completed"
            return $true
        }
        else {
            Write-Error "Database migrations failed"
            return $false
        }
    }
    catch {
        Write-Error "Database migrations failed: $($_.Exception.Message)"
        return $false
    }
}

function Start-TestDatabases {
    Write-Status "Starting test databases with Docker..."
    
    if (-not (Test-Command "docker-compose")) {
        Write-Error "docker-compose is not installed"
        return $false
    }
    
    # Stop any existing test containers
    try {
        docker-compose -f docker-compose.test.yml down 2>$null | Out-Null
    }
    catch {
        # Ignore errors when stopping non-existent containers
    }
    
    # Start test databases
    try {
        docker-compose -f docker-compose.test.yml up -d
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Test databases started"
            
            # Wait for databases to be ready
            if ((Wait-ForMySQL -Host "127.0.0.1" -Port 3307 -User "root" -Password "testpassword") -and
                (Wait-ForRedis -Host "127.0.0.1" -Port 6380)) {
                
                # Update environment for test databases
                [System.Environment]::SetEnvironmentVariable("MYSQL_PORT", "3307", "Process")
                [System.Environment]::SetEnvironmentVariable("REDIS_PORT", "6380", "Process")
                
                return $true
            }
            else {
                return $false
            }
        }
        else {
            Write-Error "Failed to start test databases"
            return $false
        }
    }
    catch {
        Write-Error "Failed to start test databases: $($_.Exception.Message)"
        return $false
    }
}

function Stop-TestDatabases {
    Write-Status "Stopping test databases..."
    
    try {
        docker-compose -f docker-compose.test.yml down -v 2>$null | Out-Null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Test databases stopped"
        }
        else {
            Write-Warning "Failed to stop test databases (they may not have been running)"
        }
    }
    catch {
        Write-Warning "Failed to stop test databases (they may not have been running)"
    }
}

function Invoke-IntegrationTests {
    Write-Status "Running integration tests..."
    
    # Check if Go is installed
    if (-not (Test-Command "go")) {
        Write-Error "Go is not installed"
        return $false
    }
    
    # Download dependencies
    Write-Status "Downloading Go dependencies..."
    go mod download
    
    # Run integration tests
    try {
        go test -v -tags=integration -coverprofile=integration-coverage.out ./tests/integration/...
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Integration tests passed!"
            
            # Generate coverage report if possible
            try {
                go tool cover -html=integration-coverage.out -o integration-coverage.html
                Write-Success "Coverage report generated: integration-coverage.html"
            }
            catch {
                Write-Warning "Could not generate coverage report"
            }
            
            return $true
        }
        else {
            Write-Error "Integration tests failed!"
            return $false
        }
    }
    catch {
        Write-Error "Integration tests failed: $($_.Exception.Message)"
        return $false
    }
}

function Show-Help {
    Write-Host "Usage: .\run-integration-tests.ps1 [OPTIONS]"
    Write-Host "Options:"
    Write-Host "  -NoDocker        Use existing databases instead of starting Docker containers"
    Write-Host "  -SkipMigrations  Skip running database migrations"
    Write-Host "  -Help            Show this help message"
}

function Main {
    if ($Help) {
        Show-Help
        exit 0
    }
    
    Write-Status "Starting integration test runner..."
    
    # Setup test environment
    Setup-TestEnvironment
    
    # Start databases if using Docker
    if (-not $NoDocker) {
        if (-not (Start-TestDatabases)) {
            exit 1
        }
    }
    else {
        Write-Status "Using existing databases (-NoDocker specified)"
        
        # Wait for existing databases
        $mysqlHost = [System.Environment]::GetEnvironmentVariable("MYSQL_HOST") ?? "127.0.0.1"
        $mysqlPort = [int]([System.Environment]::GetEnvironmentVariable("MYSQL_PORT") ?? "3306")
        $mysqlUser = [System.Environment]::GetEnvironmentVariable("MYSQL_USER") ?? "root"
        $mysqlPassword = [System.Environment]::GetEnvironmentVariable("MYSQL_PASSWORD") ?? "testpassword"
        $redisHost = [System.Environment]::GetEnvironmentVariable("REDIS_HOST") ?? "127.0.0.1"
        $redisPort = [int]([System.Environment]::GetEnvironmentVariable("REDIS_PORT") ?? "6379")
        
        if (-not (Wait-ForMySQL -Host $mysqlHost -Port $mysqlPort -User $mysqlUser -Password $mysqlPassword)) {
            exit 1
        }
        if (-not (Wait-ForRedis -Host $redisHost -Port $redisPort)) {
            exit 1
        }
    }
    
    # Run migrations unless skipped
    if (-not $SkipMigrations) {
        if (-not (Invoke-Migrations)) {
            if (-not $NoDocker) {
                Stop-TestDatabases
            }
            exit 1
        }
    }
    else {
        Write-Status "Skipping database migrations (-SkipMigrations specified)"
    }
    
    # Run integration tests
    $testResult = Invoke-IntegrationTests
    
    # Cleanup
    if (-not $NoDocker) {
        Stop-TestDatabases
    }
    
    if ($testResult) {
        Write-Success "Integration test runner completed successfully!"
        exit 0
    }
    else {
        exit 1
    }
}

# Run main function
Main
