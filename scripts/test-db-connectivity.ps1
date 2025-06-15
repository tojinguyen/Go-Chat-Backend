# Quick Database Connection Test Script for PowerShell
# This script tests database connectivity with enhanced error handling

param(
    [string]$RedisHost = "127.0.0.1",
    [int]$RedisPort = 6379,
    [string]$MySQLHost = "127.0.0.1",
    [int]$MySQLPort = 3306,
    [string]$MySQLUser = "root",
    [string]$MySQLPassword = "testpassword",
    [int]$MaxAttempts = 60
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

function Test-RedisConnection {
    param(
        [string]$Host,
        [int]$Port,
        [int]$MaxAttempts
    )
    
    Write-Status "Testing Redis connection to ${Host}:${Port}..."
    
    for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
        Write-Status "Attempt $attempt/$MaxAttempts..."
        
        # Test 1: Check if port is open
        try {
            $portTest = Test-NetConnection -ComputerName $Host -Port $Port -InformationLevel Quiet -WarningAction SilentlyContinue
            if ($portTest) {
                Write-Success "Port $Port is open"
            }
            else {
                Write-Warning "Port $Port is not open yet"
                Start-Sleep -Seconds 2
                continue
            }
        }
        catch {
            Write-Warning "Port check failed: $($_.Exception.Message)"
            Start-Sleep -Seconds 2
            continue
        }
        
        # Test 2: Try Redis ping
        try {
            $pingResult = redis-cli -h $Host -p $Port ping 2>$null
            if ($LASTEXITCODE -eq 0 -and $pingResult -eq "PONG") {
                Write-Success "Redis PING successful: $pingResult"
                
                # Test 3: Basic Redis operations
                Write-Status "Testing basic Redis operations..."
                redis-cli -h $Host -p $Port set test_key "test_value" 2>$null | Out-Null
                if ($LASTEXITCODE -eq 0) {
                    $getValue = redis-cli -h $Host -p $Port get test_key 2>$null
                    if ($getValue -eq "test_value") {
                        Write-Success "Redis SET/GET operations working"
                        redis-cli -h $Host -p $Port del test_key 2>$null | Out-Null
                        return $true
                    }
                    else {
                        Write-Error "Redis GET operation failed"
                    }
                }
                else {
                    Write-Error "Redis SET operation failed"
                }
            }
            else {
                Write-Warning "Redis PING failed: $pingResult (exit code: $LASTEXITCODE)"
            }
        }
        catch {
            Write-Warning "Redis ping failed: $($_.Exception.Message)"
        }
        
        Start-Sleep -Seconds 2
    }
    
    Write-Error "Redis connection test failed after $MaxAttempts attempts"
    return $false
}

function Test-MySQLConnection {
    param(
        [string]$Host,
        [int]$Port,
        [string]$User,
        [string]$Password,
        [int]$MaxAttempts
    )
    
    Write-Status "Testing MySQL connection to ${Host}:${Port}..."
    
    for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
        Write-Status "Attempt $attempt/$MaxAttempts..."
        
        # Test 1: Check if port is open
        try {
            $portTest = Test-NetConnection -ComputerName $Host -Port $Port -InformationLevel Quiet -WarningAction SilentlyContinue
            if ($portTest) {
                Write-Success "Port $Port is open"
            }
            else {
                Write-Warning "Port $Port is not open yet"
                Start-Sleep -Seconds 2
                continue
            }
        }
        catch {
            Write-Warning "Port check failed: $($_.Exception.Message)"
            Start-Sleep -Seconds 2
            continue
        }
        
        # Test 2: Try MySQL ping
        try {
            mysqladmin ping -h $Host -P $Port -u $User -p$Password --silent 2>$null | Out-Null
            if ($LASTEXITCODE -eq 0) {
                Write-Success "MySQL ping successful"
                
                # Test 3: Try basic MySQL query
                mysql -h $Host -P $Port -u $User -p$Password -e "SELECT 1;" 2>$null | Out-Null
                if ($LASTEXITCODE -eq 0) {
                    Write-Success "MySQL basic query working"
                    return $true
                }
                else {
                    Write-Warning "MySQL basic query failed"
                }
            }
            else {
                Write-Warning "MySQL ping failed (exit code: $LASTEXITCODE)"
            }
        }
        catch {
            Write-Warning "MySQL ping failed: $($_.Exception.Message)"
        }
        
        Start-Sleep -Seconds 2
    }
    
    Write-Error "MySQL connection test failed after $MaxAttempts attempts"
    return $false
}

function Main {
    Write-Status "Starting database connectivity tests..."
    
    # Test Redis
    if (Test-RedisConnection -Host $RedisHost -Port $RedisPort -MaxAttempts $MaxAttempts) {
        Write-Success "Redis connectivity test passed"
    }
    else {
        Write-Error "Redis connectivity test failed"
        exit 1
    }
    
    # Test MySQL
    if (Test-MySQLConnection -Host $MySQLHost -Port $MySQLPort -User $MySQLUser -Password $MySQLPassword -MaxAttempts $MaxAttempts) {
        Write-Success "MySQL connectivity test passed"
    }
    else {
        Write-Error "MySQL connectivity test failed"
        exit 1
    }
    
    Write-Success "All database connectivity tests passed!"
}

# Run main function
Main
