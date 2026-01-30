# SecureVault Quick Reference

Quick command reference for common tasks with SecureVault.

## Environment Setup

```bash
# Copy example environment file
cp .env.example .env

# Required variables
DATABASE_URL=postgres://user:pass@localhost:5432/securevault?sslmode=require
JWT_SECRET=$(openssl rand -base64 32)
SERVER_PORT=8080

# Optional production variables
ENABLE_TLS=true
TLS_CERT_FILE=certs/server.crt
TLS_KEY_FILE=certs/server.key
RATE_LIMIT_MAX=100
RATE_LIMIT_WINDOW_SECONDS=60
```

## Building & Running

```bash
# Install dependencies
go mod download

# Build
go build -o bin/securevault cmd/server/main.go

# Run
./bin/securevault

# Run with custom port
SERVER_PORT=9090 ./bin/securevault
```

## Testing

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/crypto -v

# Benchmarks
go test -bench=. ./internal/crypto

# Race detection
go test -race ./...

# Load test (requires hey)
hey -n 1000 -c 10 http://localhost:8080/health
```

## API Quick Reference

### Authentication

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user","email":"user@example.com","password":"pass123","mfa_enabled":false}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass123"}'

# Save token
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Secrets

```bash
# Create secret
curl -X POST http://localhost:8080/api/v1/secrets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"db_password","value":"secret123","expires_in_days":30}'

# Get secret
curl -X GET http://localhost:8080/api/v1/secrets/db_password \
  -H "Authorization: Bearer $TOKEN"

# List secrets
curl -X GET http://localhost:8080/api/v1/secrets \
  -H "Authorization: Bearer $TOKEN"

# Delete secret
curl -X DELETE http://localhost:8080/api/v1/secrets/db_password \
  -H "Authorization: Bearer $TOKEN"
```

### Service Principals

```bash
# Create service principal
curl -X POST http://localhost:8080/api/v1/service-principals \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"ci_pipeline","permissions":["secrets:read"],"rate_limit":1000}'

# SP login
curl -X POST http://localhost:8080/api/v1/auth/sp-login \
  -H "Content-Type: application/json" \
  -d '{"client_id":"CLIENT_ID","client_secret":"CLIENT_SECRET"}'
```

### Key Management

```bash
# Rotate key
curl -X POST http://localhost:8080/api/v1/vault/rotate-key \
  -H "Authorization: Bearer $TOKEN"

# Key status
curl -X GET http://localhost:8080/api/v1/keys/status \
  -H "Authorization: Bearer $TOKEN"
```

### Monitoring

```bash
# Health check
curl http://localhost:8080/health

# Metrics (Prometheus format)
curl http://localhost:8080/metrics

# Check rate limit headers
curl -i http://localhost:8080/health | grep X-RateLimit
```

## Database Operations

```bash
# Create database
createdb securevault

# Connect
psql securevault

# Run migrations
psql securevault < migrations/001_initial_schema.sql

# Backup
pg_dump securevault > backup.sql

# Restore
psql securevault < backup.sql

# Check tables
psql securevault -c "\dt"

# View audit logs
psql securevault -c "SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT 10;"
```

## TLS Setup

```bash
# Generate self-signed certificates (development)
mkdir -p certs
openssl req -x509 -newkey rsa:4096 -keyout certs/server.key \
  -out certs/server.crt -days 365 -nodes \
  -subj "/CN=localhost"

# Let's Encrypt (production)
sudo certbot certonly --standalone -d yourdomain.com
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem certs/server.crt
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem certs/server.key

# Test TLS
openssl s_client -connect localhost:8080 -tls1_3
```

## Docker (Optional)

```dockerfile
# Dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o securevault cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/securevault .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./securevault"]
```

```bash
# Build
docker build -t securevault:latest .

# Run
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://..." \
  -e JWT_SECRET="..." \
  securevault:latest
```

## Common Workflows

### Initial Setup

```bash
# 1. Setup environment
cp .env.example .env
# Edit .env

# 2. Create database
createdb securevault

# 3. Build
go build -o bin/securevault cmd/server/main.go

# 4. Run (migrations run automatically)
./bin/securevault
```

### Adding a New User

```bash
# 1. Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"SecurePass123!","mfa_enabled":false}'

# 2. Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"SecurePass123!"}' | jq -r '.token')

# 3. Create first secret
curl -X POST http://localhost:8080/api/v1/secrets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"test","value":"hello"}'
```

### CI/CD Integration

```bash
# 1. Create service principal for CI
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"AdminPass123!"}' | jq -r '.token')

SP=$(curl -s -X POST http://localhost:8080/api/v1/service-principals \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"github_actions","permissions":["secrets:read"],"rate_limit":5000}')

CLIENT_ID=$(echo $SP | jq -r '.client_id')
CLIENT_SECRET=$(echo $SP | jq -r '.client_secret')

# 2. Add to CI secrets
echo "SECUREVAULT_CLIENT_ID=$CLIENT_ID"
echo "SECUREVAULT_CLIENT_SECRET=$CLIENT_SECRET"

# 3. Use in CI pipeline
SP_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/sp-login \
  -H "Content-Type: application/json" \
  -d "{\"client_id\":\"$CLIENT_ID\",\"client_secret\":\"$CLIENT_SECRET\"}" | jq -r '.token')

# Fetch secrets
curl -X GET http://localhost:8080/api/v1/secrets/db_password \
  -H "Authorization: Bearer $SP_TOKEN"
```

### Key Rotation

```bash
# 1. Login as admin
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"AdminPass123!"}' | jq -r '.token')

# 2. Check current key status
curl -X GET http://localhost:8080/api/v1/keys/status \
  -H "Authorization: Bearer $TOKEN"

# 3. Rotate key
curl -X POST http://localhost:8080/api/v1/vault/rotate-key \
  -H "Authorization: Bearer $TOKEN"

# 4. Verify all secrets still accessible
curl -X GET http://localhost:8080/api/v1/secrets \
  -H "Authorization: Bearer $TOKEN"
```

## Troubleshooting

```bash
# Check server is running
curl http://localhost:8080/health

# Check database connection
psql $DATABASE_URL -c "SELECT 1"

# View logs (if using systemd)
sudo journalctl -u securevault -f

# Check TLS certificate
openssl x509 -in certs/server.crt -text -noout

# Test rate limiting
for i in {1..150}; do curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/health; done

# Check metrics
curl http://localhost:8080/metrics | grep securevault_requests_total

# Database: Find secrets by tag
psql securevault -c "SELECT name, tags FROM secrets WHERE 'production' = ANY(tags);"

# Database: Audit trail for user
psql securevault -c "SELECT * FROM audit_logs WHERE user_id = 'USER_UUID' ORDER BY created_at DESC LIMIT 20;"
```

## Useful SQL Queries

```sql
-- List all users
SELECT id, username, email, created_at FROM users;

-- Count secrets per user
SELECT u.username, COUNT(s.id) as secret_count
FROM users u
LEFT JOIN secrets s ON u.id = s.user_id
GROUP BY u.id, u.username;

-- Find expired secrets
SELECT name, expires_at FROM secrets WHERE expires_at < NOW();

-- Recent audit activity
SELECT action, ip_address, created_at FROM audit_logs ORDER BY created_at DESC LIMIT 50;

-- Service principals with rate limits
SELECT name, permissions, rate_limit FROM service_principals;

-- Key rotation history
SELECT version, status, created_at, rotation_reason FROM key_versions ORDER BY version DESC;
```

## Performance Tips

```bash
# Enable connection pooling (already configured)
# Max open: 25, Max idle: 5

# Monitor database performance
psql securevault -c "SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"

# Check memory usage
curl http://localhost:8080/metrics | grep securevault_memory

# Check goroutines
curl http://localhost:8080/metrics | grep securevault_goroutines

# Profile CPU
curl http://localhost:8080/debug/pprof/profile > cpu.prof
go tool pprof cpu.prof

# Profile memory
curl http://localhost:8080/debug/pprof/heap > mem.prof
go tool pprof mem.prof
```

## Security Checklist

```bash
# ✓ Strong JWT secret (32+ bytes)
openssl rand -base64 32

# ✓ TLS enabled
grep ENABLE_TLS .env

# ✓ Database SSL
echo $DATABASE_URL | grep sslmode=require

# ✓ Rate limiting enabled
curl -i http://localhost:8080/health | grep X-RateLimit

# ✓ Security headers
curl -i http://localhost:8080/health | grep -E "(Strict-Transport|X-Frame|X-Content)"

# ✓ No vulnerabilities
govulncheck ./...

# ✓ Static analysis
gosec ./...
```

## Links

- [Full Documentation](README.md)
- [Deployment Guide](DEPLOYMENT.md)
- [Testing Guide](TESTING.md)
- [Security](SECURITY.md)
- [Examples](EXAMPLES.md)
- [Contributing](CONTRIBUTING.md)
