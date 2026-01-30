# Development Script - Start App Vault

Write-Host "🚀 Starting App Vault Development Environment..." -ForegroundColor Cyan
Write-Host ""

# Function to check if a port is in use
function Test-Port {
    param([int]$Port)
    $connection = Test-NetConnection -ComputerName localhost -Port $Port -InformationLevel Quiet -WarningAction SilentlyContinue
    return $connection
}

# Function to start a process in a new window
function Start-ProcessInNewWindow {
    param(
        [string]$Title,
        [string]$Command,
        [string]$WorkingDirectory
    )
    
    $ps = Start-Process powershell -ArgumentList "-NoExit", "-Command", "& {Write-Host '$Title' -ForegroundColor Green; cd '$WorkingDirectory'; $Command}" -PassThru
    return $ps
}

# Check prerequisites
Write-Host "📋 Checking prerequisites..." -ForegroundColor Yellow

# Check Go
if (!(Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Go is not installed. Please install Go 1.23+ from https://go.dev" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Go: $(go version)" -ForegroundColor Green

# Check Node.js
if (!(Get-Command node -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Node.js is not installed. Please install Node.js 18+ from https://nodejs.org" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Node.js: v$(node --version)" -ForegroundColor Green

# Check PostgreSQL connection
Write-Host ""
Write-Host "🔌 Checking PostgreSQL..." -ForegroundColor Yellow
$dbRunning = Test-Port -Port 5432
if (-not $dbRunning) {
    Write-Host "⚠️  PostgreSQL is not running on port 5432" -ForegroundColor Yellow
    Write-Host "   Starting PostgreSQL with Docker..." -ForegroundColor Yellow
    
    if (Get-Command docker -ErrorAction SilentlyContinue) {
        docker run --name appvault-postgres `
            -e POSTGRES_PASSWORD=postgres `
            -e POSTGRES_DB=appvault `
            -p 5432:5432 `
            -d postgres:15-alpine
        
        Write-Host "   Waiting for PostgreSQL to start..." -ForegroundColor Yellow
        Start-Sleep -Seconds 5
        Write-Host "✅ PostgreSQL started" -ForegroundColor Green
    } else {
        Write-Host "❌ Docker not found. Please start PostgreSQL manually or install Docker" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "✅ PostgreSQL is running" -ForegroundColor Green
}

# Set environment variables
Write-Host ""
Write-Host "🔧 Configuring environment..." -ForegroundColor Yellow
$env:SERVER_PORT = "8888"
$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/appvault?sslmode=disable"
$env:JWT_SECRET = "dev-jwt-secret-change-in-production-$(Get-Random)"
$env:RATE_LIMIT_ENABLED = "true"
$env:RATE_LIMIT_MAX = "100"
Write-Host "✅ Environment configured" -ForegroundColor Green

# Check if backend is already running
Write-Host ""
Write-Host "🔍 Checking if services are already running..." -ForegroundColor Yellow
if (Test-Port -Port 8888) {
    Write-Host "⚠️  Backend already running on port 8888" -ForegroundColor Yellow
} else {
    Write-Host "   Starting Go backend..." -ForegroundColor Yellow
    $backendProcess = Start-ProcessInNewWindow `
        -Title "App Vault Backend (Port 8888)" `
        -Command "go run cmd/server/main.go" `
        -WorkingDirectory $PSScriptRoot
    
    Write-Host "✅ Backend starting in new window (PID: $($backendProcess.Id))" -ForegroundColor Green
    Start-Sleep -Seconds 3
}

# Install frontend dependencies if needed
$webPath = Join-Path $PSScriptRoot "web"
$nodeModulesPath = Join-Path $webPath "node_modules"

Write-Host ""
Write-Host "📦 Checking frontend dependencies..." -ForegroundColor Yellow
if (-not (Test-Path $nodeModulesPath)) {
    Write-Host "   Installing npm packages..." -ForegroundColor Yellow
    Push-Location $webPath
    npm install
    Pop-Location
    Write-Host "✅ Dependencies installed" -ForegroundColor Green
} else {
    Write-Host "✅ Dependencies already installed" -ForegroundColor Green
}

# Check if frontend is already running
if (Test-Port -Port 5173) {
    Write-Host "⚠️  Frontend already running on port 5173" -ForegroundColor Yellow
} else {
    Write-Host ""
    Write-Host "   Starting Vue.js frontend..." -ForegroundColor Yellow
    $frontendProcess = Start-ProcessInNewWindow `
        -Title "App Vault Frontend (Port 5173)" `
        -Command "npm run dev" `
        -WorkingDirectory $webPath
    
    Write-Host "✅ Frontend starting in new window (PID: $($frontendProcess.Id))" -ForegroundColor Green
}

# Wait for services to be ready
Write-Host ""
Write-Host "⏳ Waiting for services to be ready..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0

while ($attempt -lt $maxAttempts) {
    $backendReady = Test-Port -Port 8888
    $frontendReady = Test-Port -Port 5173
    
    if ($backendReady -and $frontendReady) {
        break
    }
    
    Start-Sleep -Seconds 1
    $attempt++
    Write-Host "." -NoNewline
}

Write-Host ""

if ($attempt -eq $maxAttempts) {
    Write-Host "⚠️  Services took longer than expected to start" -ForegroundColor Yellow
    Write-Host "   Check the terminal windows for errors" -ForegroundColor Yellow
} else {
    Write-Host ""
    Write-Host "✅ All services are ready!" -ForegroundColor Green
}

# Display summary
Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "🎉 App Vault is running!" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""
Write-Host "📍 Services:" -ForegroundColor White
Write-Host "   Frontend:  http://localhost:5173" -ForegroundColor Green
Write-Host "   Backend:   http://localhost:8888" -ForegroundColor Green
Write-Host "   Health:    http://localhost:8888/health" -ForegroundColor Green
Write-Host "   Metrics:   http://localhost:8888/metrics" -ForegroundColor Green
Write-Host ""
Write-Host "🔐 First time setup:" -ForegroundColor White
Write-Host "   1. Go to http://localhost:5173" -ForegroundColor Gray
Write-Host "   2. Click 'Create Account'" -ForegroundColor Gray
Write-Host "   3. Register with email and password" -ForegroundColor Gray
Write-Host "   4. SAVE the generated secret key!" -ForegroundColor Yellow
Write-Host "   5. Login with email, password, and secret key" -ForegroundColor Gray
Write-Host ""
Write-Host "💡 Tips:" -ForegroundColor White
Write-Host "   • Backend logs are in the 'Backend' terminal window" -ForegroundColor Gray
Write-Host "   • Frontend logs are in the 'Frontend' terminal window" -ForegroundColor Gray
Write-Host "   • Press Ctrl+C in those windows to stop services" -ForegroundColor Gray
Write-Host ""
Write-Host "📚 Documentation:" -ForegroundColor White
Write-Host "   • QUICKSTART.md - Quick start guide" -ForegroundColor Gray
Write-Host "   • web/README.md - Frontend documentation" -ForegroundColor Gray
Write-Host "   • README.md - Project overview" -ForegroundColor Gray
Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""

# Open browser
Write-Host "🌐 Opening browser..." -ForegroundColor Yellow
Start-Sleep -Seconds 2
Start-Process "http://localhost:5173"

Write-Host ""
Write-Host "✨ Happy coding!" -ForegroundColor Green
Write-Host ""
