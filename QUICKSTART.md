# App Vault Web - Quick Start Guide

## 🚀 Getting Started

### Prerequisites
- **Go 1.23+** installed
- **Node.js 18+** and npm installed
- **PostgreSQL 12+** running

### Step 1: Start PostgreSQL

```powershell
# Using Docker (recommended)
docker run --name appvault-postgres `
  -e POSTGRES_PASSWORD=postgres `
  -e POSTGRES_DB=appvault `
  -p 5432:5432 `
  -d postgres:15-alpine

# Or use existing PostgreSQL instance
```

### Step 2: Configure Backend

```powershell
# Set environment variables (or use .env file)
$env:SERVER_PORT="8888"
$env:DATABASE_URL="postgres://postgres:postgres@localhost:5432/appvault?sslmode=disable"
$env:JWT_SECRET="your-super-secret-jwt-key-change-in-production"
```

### Step 3: Run Database Migrations

```powershell
# From project root
cd c:\Sources\GitHub\app-vault

# Run migrations (backend will auto-migrate on startup)
```

### Step 4: Start Backend API

```powershell
# Terminal 1: Start the Go backend
cd c:\Sources\GitHub\app-vault
go run cmd/server/main.go

# Backend will start on http://localhost:8888
# Check health: http://localhost:8888/health
```

### Step 5: Install Frontend Dependencies

```powershell
# Terminal 2: Install npm packages
cd c:\Sources\GitHub\app-vault\web
npm install
```

### Step 6: Start Frontend Dev Server

```powershell
# Same terminal (Terminal 2)
npm run dev

# Frontend will start on http://localhost:5173
# Opens automatically in browser
```

## 🎯 Access the Application

1. Open browser to **http://localhost:5173**
2. Click "Create Account"
3. Register with email and password
4. **IMPORTANT**: Save the generated secret key (shown only once!)
5. Login with email, password, and secret key
6. Start managing secrets! 🔐

## 🔧 Development Workflow

### Backend Development

```powershell
# Run with hot reload using air (install: go install github.com/cosmtrek/air@latest)
cd c:\Sources\GitHub\app-vault
air

# Or use go run
go run cmd/server/main.go

# Run tests
go test ./...

# Build binary
go build -o bin/app-vault.exe cmd/server/main.go
```

### Frontend Development

```powershell
cd c:\Sources\GitHub\app-vault\web

# Development server (hot reload)
npm run dev

# Type checking
npm run type-check

# Build for production
npm run build

# Preview production build
npm run preview
```

## 📁 Project Structure

```
app-vault/
├── cmd/server/main.go        # Backend entry point
├── internal/
│   ├── api/                  # HTTP handlers
│   ├── crypto/               # Encryption service
│   ├── db/                   # Database layer
│   ├── service/              # Business logic
│   └── ...
├── migrations/               # SQL migrations
├── web/                      # Frontend application
│   ├── src/
│   │   ├── api/             # API clients
│   │   ├── components/      # Vue components
│   │   ├── views/           # Pages
│   │   ├── stores/          # Pinia stores
│   │   └── ...
│   ├── package.json
│   └── vite.config.ts
└── docker-compose.yml        # Docker setup (optional)
```

## 🐳 Docker Compose (Alternative)

```powershell
# Start everything with Docker Compose
docker-compose up -d

# Backend: http://localhost:8888
# Frontend: Needs to be served separately or built and embedded
```

## 🔐 Configuration

### Backend Environment Variables

```env
SERVER_PORT=8888
DATABASE_URL=postgres://postgres:postgres@localhost:5432/appvault?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key
ARGON_MEMORY=65536
ARGON_ITERATIONS=3
ARGON_PARALLELISM=4
RATE_LIMIT_MAX=100
RATE_LIMIT_WINDOW_SECONDS=60
ENABLE_TLS=false
```

### Frontend Environment Variables

```env
VITE_API_URL=http://localhost:8888/api/v1
VITE_APP_NAME=SecureVault
VITE_VERSION=2.1.0
```

## 📊 API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login with credentials
- `POST /api/v1/auth/logout` - Logout

### Secrets
- `POST /api/v1/secrets` - Create secret
- `GET /api/v1/secrets` - List secrets (metadata only)
- `GET /api/v1/secrets/:id` - Get secret with value
- `GET /api/v1/secrets/by-name?name=X` - Get by name
- `PUT /api/v1/secrets/:id` - Update secret
- `DELETE /api/v1/secrets/:id` - Delete secret

### Key Management
- `POST /api/v1/keys/rotate` - Rotate vault key
- `GET /api/v1/keys/status` - Get key status

### System
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

## 🧪 Testing the API

```powershell
# Health check
Invoke-WebRequest http://localhost:8888/health

# Register user
$body = @{
    email = "admin@example.com"
    password = "SecurePassword123!"
} | ConvertTo-Json

Invoke-WebRequest -Method POST `
  -Uri http://localhost:8888/api/v1/auth/register `
  -ContentType "application/json" `
  -Body $body

# Login
$body = @{
    email = "admin@example.com"
    password = "SecurePassword123!"
    secretKey = "A3-XXXXXX-XXXXXX-XXXXX"
} | ConvertTo-Json

$response = Invoke-WebRequest -Method POST `
  -Uri http://localhost:8888/api/v1/auth/login `
  -ContentType "application/json" `
  -Body $body

$token = ($response.Content | ConvertFrom-Json).token
```

## 🛠️ Troubleshooting

### Backend Won't Start
- Check PostgreSQL is running: `docker ps` or check service
- Verify DATABASE_URL is correct
- Check port 8888 is not in use: `netstat -ano | findstr :8888`

### Frontend Won't Connect to API
- Verify backend is running on http://localhost:8888
- Check CORS configuration in backend
- Inspect browser console for errors
- Check Vite proxy configuration in `vite.config.ts`

### Database Errors
- Run migrations: Backend auto-migrates on startup
- Check PostgreSQL logs
- Verify database exists: `psql -U postgres -c "\l"`

### Build Errors
```powershell
# Clear Go module cache
go clean -modcache

# Clear npm cache
cd web
rm -rf node_modules package-lock.json
npm install
```

## 📝 Next Steps

1. ✅ **Phase 1**: Core secrets management (COMPLETE)
2. 🔄 **Phase 2**: Service Principals endpoints (IN PROGRESS)
3. ⏳ **Phase 3**: Key Rotation UI
4. ⏳ **Phase 4**: Audit Logs UI
5. ⏳ **Phase 5**: Advanced features (export/import, bulk ops)

## 🔗 Useful Links

- Backend API: http://localhost:8888
- Frontend UI: http://localhost:5173
- Health Check: http://localhost:8888/health
- Metrics: http://localhost:8888/metrics

## 💡 Tips

- **Secret Key**: The A3-format secret key is generated ONCE during registration. Save it securely!
- **Session Storage**: Secret key is stored in sessionStorage and cleared when tab closes
- **JWT Tokens**: Valid for 24 hours, stored in localStorage
- **Hot Reload**: Both frontend and backend support hot reload in dev mode
- **Database**: Migrations run automatically on backend startup

## 🆘 Getting Help

- Check the logs: Backend stdout, Browser console
- Verify environment variables are set
- Ensure all services are running
- Check README.md for detailed documentation

---

**Happy Coding! 🚀**
