# Docker Compose Setup Guide

## 🐳 Running App Vault with Docker Compose

The complete stack includes:
- **PostgreSQL** (port 5432) - Database
- **App Vault Backend** (port 8888) - Go API server
- **Web Frontend** (port 80) - Vue 3 application with Nginx

## Quick Start

### 1. Create Environment File

Create a `.env` file in the project root:

```env
# Server Configuration
SERVER_PORT=8888
ENABLE_TLS=false

# Database Configuration
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=appvault
DB_PORT=5432
DATABASE_URL=postgres://postgres:postgres@postgres:5432/appvault?sslmode=disable

# Security
JWT_SECRET=your-super-secret-jwt-key-change-in-production-$(openssl rand -hex 32)

# Encryption
ARGON_MEMORY=65536
ARGON_ITERATIONS=3
ARGON_PARALLELISM=4

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX=100
RATE_LIMIT_WINDOW_SECONDS=60

# Frontend
VITE_API_URL=/api/v1
```

### 2. Start All Services

```powershell
# Start everything
docker-compose up -d

# View logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f web
docker-compose logs -f appvault
docker-compose logs -f postgres
```

### 3. Access the Application

- **Web UI**: http://localhost
- **Backend API**: http://localhost:8888
- **Health Check**: http://localhost/health
- **Metrics**: http://localhost:8888/metrics

### 4. Stop Services

```powershell
# Stop all services
docker-compose down

# Stop and remove volumes (deletes database data)
docker-compose down -v
```

## Service Details

### PostgreSQL Database
- **Container**: `appvault-postgres`
- **Port**: 5432
- **Data**: Persisted in Docker volume `postgres_data`
- **Credentials**: Set via environment variables

### Backend API
- **Container**: `appvault-server`
- **Port**: 8888
- **Built from**: Project root Dockerfile
- **Depends on**: PostgreSQL (waits for healthy status)

### Web Frontend
- **Container**: `appvault-web`
- **Ports**: 80 (HTTP), 443 (HTTPS ready)
- **Built from**: `web/Dockerfile`
- **Serves**: Vue 3 SPA with Nginx
- **Proxies**: All `/api/*` requests to backend

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      Docker Host                        │
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Web Frontend (Nginx)                           │   │
│  │  Port: 80, 443                                  │   │
│  │  • Serves static files                          │   │
│  │  • Proxies /api/* to backend                    │   │
│  └─────────────┬───────────────────────────────────┘   │
│                │                                         │
│                ▼                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Backend API (Go)                               │   │
│  │  Port: 8888                                     │   │
│  │  • REST API endpoints                           │   │
│  │  • JWT authentication                           │   │
│  │  • Secret encryption                            │   │
│  └─────────────┬───────────────────────────────────┘   │
│                │                                         │
│                ▼                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │  PostgreSQL Database                            │   │
│  │  Port: 5432                                     │   │
│  │  • Persistent storage (volume)                  │   │
│  │  • User credentials                             │   │
│  │  • Encrypted secrets                            │   │
│  └─────────────────────────────────────────────────┘   │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## Development vs Production

### Development (Current Setup)
- Single `.env` file
- HTTP only (no TLS)
- Debug logging enabled
- Hot reload not available (use local dev for that)

### Production Recommendations
1. **Enable TLS/HTTPS**:
   ```env
   ENABLE_TLS=true
   TLS_CERT_FILE=/certs/server.crt
   TLS_KEY_FILE=/certs/server.key
   ```

2. **Use Docker Secrets** for sensitive data:
   ```yaml
   services:
     appvault:
       secrets:
         - jwt_secret
         - db_password
   ```

3. **Add Nginx HTTPS configuration**:
   - Mount SSL certificates
   - Configure HTTPS redirect
   - Use Let's Encrypt with certbot

4. **Resource Limits**:
   ```yaml
   services:
     appvault:
       deploy:
         resources:
           limits:
             cpus: '1'
             memory: 512M
   ```

5. **Use External Database** (production):
   - Remove postgres service
   - Point DATABASE_URL to external PostgreSQL
   - Use managed database service

## Useful Commands

### Build and Start
```powershell
# Build images
docker-compose build

# Build without cache
docker-compose build --no-cache

# Start in foreground (see logs)
docker-compose up

# Start in background
docker-compose up -d
```

### Management
```powershell
# View running containers
docker-compose ps

# View logs (all services)
docker-compose logs -f

# View logs (specific service)
docker-compose logs -f web

# Restart a service
docker-compose restart web

# Stop a service
docker-compose stop web

# Remove containers (keeps volumes)
docker-compose down

# Remove everything including volumes
docker-compose down -v
```

### Debugging
```powershell
# Execute command in container
docker-compose exec appvault sh
docker-compose exec web sh
docker-compose exec postgres psql -U postgres -d appvault

# View environment variables
docker-compose exec appvault env

# Check health status
docker-compose ps
```

### Database
```powershell
# Access PostgreSQL
docker-compose exec postgres psql -U postgres -d appvault

# Backup database
docker-compose exec postgres pg_dump -U postgres appvault > backup.sql

# Restore database
cat backup.sql | docker-compose exec -T postgres psql -U postgres -d appvault

# View database logs
docker-compose logs postgres
```

## Networking

All services are on the same Docker network and can communicate using service names:
- `postgres:5432` - Database (from backend)
- `appvault:8888` - API (from web frontend)
- `web:80` - Frontend (from host)

## Volumes

### postgres_data
- **Purpose**: Persistent database storage
- **Location**: Docker managed volume
- **View**: `docker volume ls`
- **Inspect**: `docker volume inspect app-vault_postgres_data`
- **Backup**: `docker run --rm -v app-vault_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres-backup.tar.gz /data`

## Troubleshooting

### Web container fails to start
```powershell
# Check build logs
docker-compose logs web

# Rebuild without cache
docker-compose build --no-cache web
docker-compose up -d web
```

### Backend can't connect to database
```powershell
# Check PostgreSQL is running and healthy
docker-compose ps postgres

# Check DATABASE_URL in backend
docker-compose exec appvault env | grep DATABASE_URL

# Test connection manually
docker-compose exec postgres pg_isready -U postgres
```

### Port already in use
```powershell
# Check what's using the port
netstat -ano | findstr :80
netstat -ano | findstr :8888

# Change port in docker-compose.yml or stop conflicting service
```

### View container resource usage
```powershell
docker stats
```

### Access logs
```powershell
# Backend logs
docker-compose logs -f appvault

# Nginx access logs
docker-compose exec web cat /var/log/nginx/access.log

# Nginx error logs
docker-compose exec web cat /var/log/nginx/error.log
```

## Security Considerations

### ⚠️ Before Production
1. Change default passwords in `.env`
2. Use strong `JWT_SECRET` (32+ random characters)
3. Enable TLS/HTTPS
4. Configure firewall rules
5. Use Docker secrets for sensitive data
6. Enable Nginx rate limiting
7. Set up log monitoring
8. Regular security updates
9. Database backups scheduled
10. Use non-root users in containers

## Health Checks

All services include health checks:
- **PostgreSQL**: `pg_isready` every 5s
- **Backend**: HTTP GET `/health` every 30s
- **Frontend**: HTTP GET `localhost:80` every 30s

Check health status:
```powershell
docker-compose ps
# Look for "healthy" status
```

## Updates and Maintenance

### Update containers
```powershell
# Pull latest images
docker-compose pull

# Rebuild and restart
docker-compose up -d --build

# View what changed
docker-compose ps
```

### Clean up
```powershell
# Remove unused images
docker image prune -a

# Remove unused volumes (careful!)
docker volume prune

# Remove everything (nuclear option)
docker system prune -a --volumes
```

---

**Ready to go!** Just run `docker-compose up -d` and access http://localhost 🚀
