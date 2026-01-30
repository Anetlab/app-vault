# Production Deployment Guide

This guide covers deploying App Vault in a production environment with all security features enabled.

## Table of Contents

- [Prerequisites](#prerequisites)
- [TLS Configuration](#tls-configuration)
- [Rate Limiting](#rate-limiting)
- [Monitoring & Metrics](#monitoring--metrics)
- [Security Headers](#security-headers)
- [Database Setup](#database-setup)
- [Environment Variables](#environment-variables)
- [Deployment Checklist](#deployment-checklist)

## Prerequisites

- Go 1.23 or higher
- PostgreSQL 13 or higher
- TLS certificates (Let''s Encrypt recommended)
- Reverse proxy (Nginx/Caddy recommended)

## TLS Configuration

### Using Let''s Encrypt with Certbot

```bash
# Install certbot
sudo apt-get install certbot

# Generate certificate
sudo certbot certonly --standalone -d yourdomain.com

# Copy certificates to certs directory
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem certs/server.crt
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem certs/server.key
sudo chown $(whoami):$(whoami) certs/*
```

### Enable TLS in .env

```env
ENABLE_TLS=true
TLS_CERT_FILE=certs/server.crt
TLS_KEY_FILE=certs/server.key
```

### TLS Configuration Details

App Vault uses:
- **TLS 1.3** minimum version
- **Strong cipher suites**: AES-256-GCM, AES-128-GCM, ChaCha20-Poly1305
- **Curve preferences**: X25519, P-256
- **HSTS** with 1-year max-age

## Rate Limiting

Rate limiting is enabled by default to prevent abuse.

### Configuration

```env
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX=100                # requests per window
RATE_LIMIT_WINDOW_SECONDS=60      # window size in seconds
```

### How it Works

- **Per-IP tracking**: Uses client IP address (supports X-Forwarded-For)
- **Sliding window algorithm**: More accurate than fixed windows
- **HTTP headers**: Returns X-RateLimit-* headers on every response
- **429 status code**: When limit exceeded

### Response Headers

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Window: 60
```

## Monitoring & Metrics

App Vault exposes Prometheus-compatible metrics at `/metrics`.

### Available Metrics

**Counters:**
- `App Vault_requests_total` - Total HTTP requests
- `App Vault_errors_total` - Total errors
- `App Vault_auth_success_total` - Successful authentications
- `App Vault_auth_failure_total` - Failed authentications
- `App Vault_secrets_created_total` - Secrets created
- `App Vault_secrets_read_total` - Secrets read
- `App Vault_secrets_deleted_total` - Secrets deleted
- `App Vault_key_rotations_total` - Key rotations

**Gauges:**
- `App Vault_uptime_seconds` - Server uptime
- `App Vault_memory_alloc_bytes` - Allocated memory
- `App Vault_memory_sys_bytes` - System memory
- `App Vault_goroutines` - Number of goroutines

**Histograms:**
- `App Vault_request_duration_ms` - Request durations by endpoint

### Prometheus Configuration

```yaml
scrape_configs:
  - job_name: ''App Vault''
    static_configs:
      - targets: [''localhost:8080'']
    metrics_path: /metrics
    scrape_interval: 15s
```

### Health Check

Health check endpoint at `/health` returns:

```json
{
  "status": "healthy",
  "database": "connected"
}
```

Returns HTTP 503 if database is unhealthy.

## Security Headers

App Vault automatically adds the following security headers to all responses:

- **Strict-Transport-Security**: Forces HTTPS for 1 year
- **X-Frame-Options**: Prevents clickjacking (DENY)
- **X-Content-Type-Options**: Prevents MIME sniffing (nosniff)
- **X-XSS-Protection**: Enables XSS filter
- **Content-Security-Policy**: Restricts resource loading
- **Referrer-Policy**: Controls referrer information
- **Permissions-Policy**: Disables unnecessary browser features

### Request Size Limits

Maximum request body size can be configured:

```env
MAX_REQUEST_SIZE_MB=1
```

Default is 1MB. Requests exceeding this will be rejected with HTTP 413.

## Database Setup

### PostgreSQL Configuration for Production

1. **Create dedicated database and user:**

```sql
CREATE DATABASE App Vault;
CREATE USER App Vault_user WITH ENCRYPTED PASSWORD ''strong_password_here'';
GRANT ALL PRIVILEGES ON DATABASE App Vault TO App Vault_user;
```

2. **Enable SSL connections:**

```env
DATABASE_URL=postgres://App Vault_user:password@localhost:5432/App Vault?sslmode=require
```

3. **Connection pooling:**

The application configures:
- Max open connections: 25
- Max idle connections: 5
- Connection lifetime: 5 minutes

4. **Backup strategy:**

```bash
# Daily backup
pg_dump App Vault > backup_$(date +%Y%m%d).sql

# Restore
psql App Vault < backup_20240101.sql
```

## Environment Variables

### Required Variables

```env
SERVER_PORT=8080
DATABASE_URL=postgres://user:pass@host:5432/App Vault?sslmode=require
JWT_SECRET=generate-a-strong-random-secret-here
```

### Optional Variables

```env
MIGRATIONS_PATH=migrations
ENABLE_TLS=true
TLS_CERT_FILE=certs/server.crt
TLS_KEY_FILE=certs/server.key
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX=100
RATE_LIMIT_WINDOW_SECONDS=60
MAX_REQUEST_SIZE_MB=1
```

### Generating Secrets

```bash
# Generate JWT secret (32 bytes, base64)
openssl rand -base64 32

# Alternative using Go
go run -c ''package main; import("crypto/rand"; "encoding/base64"; "os"); func main() { b := make([]byte, 32); rand.Read(b); os.Stdout.Write([]byte(base64.StdEncoding.EncodeToString(b))) }''
```

## Deployment Checklist

### Pre-deployment

- [ ] Change JWT_SECRET from default value
- [ ] Configure production DATABASE_URL with SSL
- [ ] Generate or obtain valid TLS certificates
- [ ] Review and adjust rate limits
- [ ] Set appropriate MAX_REQUEST_SIZE_MB
- [ ] Configure backup strategy
- [ ] Set up monitoring/alerting

### Security Review

- [ ] TLS 1.3 enabled
- [ ] Strong cipher suites configured
- [ ] Rate limiting enabled
- [ ] Security headers present
- [ ] Database using SSL connections
- [ ] Secrets not committed to version control
- [ ] Proper file permissions on certificates (600)
- [ ] Database credentials rotated from defaults

### Deployment

- [ ] Run database migrations
- [ ] Test with health check endpoint
- [ ] Verify metrics endpoint works
- [ ] Test rate limiting behavior
- [ ] Confirm TLS handshake succeeds
- [ ] Check logs for errors
- [ ] Perform end-to-end API tests

### Reverse Proxy (Nginx Example)

```nginx
upstream App Vault {
    server localhost:8080;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /path/to/fullchain.pem;
    ssl_certificate_key /path/to/privkey.pem;

    # Security headers (redundant but safe)
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    location / {
        proxy_pass http://App Vault;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Metrics endpoint (restrict access)
    location /metrics {
        allow 10.0.0.0/8;  # Internal network only
        deny all;
        proxy_pass http://App Vault;
    }
}
```

### Systemd Service

Create `/etc/systemd/system/App Vault.service`:

```ini
[Unit]
Description=App Vault API Server
After=network.target postgresql.service

[Service]
Type=simple
User=App Vault
WorkingDirectory=/opt/App Vault
EnvironmentFile=/opt/App Vault/.env
ExecStart=/opt/App Vault/App Vault
Restart=on-failure
RestartSec=5s

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/App Vault/certs

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable App Vault
sudo systemctl start App Vault
sudo systemctl status App Vault
```

### Monitoring with Prometheus & Grafana

1. **Add App Vault to Prometheus targets**
2. **Import Grafana dashboard** (see grafana-dashboard.json)
3. **Set up alerts:**
   - High error rate
   - Authentication failures spike
   - Database connection issues
   - High memory usage

### Logging

Application logs to stdout. Capture with:

```bash
# Systemd journal
sudo journalctl -u App Vault -f

# Or redirect to file
./App Vault 2>&1 | tee -a App Vault.log
```

## Performance Tuning

### Database

- Enable connection pooling in PostgreSQL
- Create indexes on frequently queried columns
- Regular VACUUM and ANALYZE

### Application

- Adjust rate limits based on traffic patterns
- Monitor memory usage and goroutine count
- Consider horizontal scaling with load balancer

### Caching

For high-traffic deployments, consider:
- Redis for rate limiting (current implementation is in-memory)
- CDN for static assets
- Read replicas for database

## Security Considerations

1. **Rotate credentials regularly:**
   - JWT secrets
   - Database passwords
   - TLS certificates

2. **Audit logs:**
   - All operations are logged to `audit_logs` table
   - Regular review for suspicious activity

3. **Network security:**
   - Firewall rules (only expose port 443)
   - VPC/private network for database
   - VPN for admin access

4. **Backup security:**
   - Encrypt backups at rest
   - Store offsite
   - Test restore procedures

5. **Incident response:**
   - Document procedures
   - Monitor for compromise
   - Have rollback plan

## Troubleshooting

### Common Issues

**TLS handshake failures:**
- Check certificate paths in .env
- Verify certificate not expired
- Ensure proper file permissions (600)

**Database connection errors:**
- Verify DATABASE_URL is correct
- Check PostgreSQL is accepting connections
- Confirm SSL requirements match

**Rate limit issues:**
- Check client IP detection with X-Forwarded-For
- Adjust limits for your traffic patterns
- Monitor rate limit metrics

**High memory usage:**
- Rate limiter cache grows with unique IPs
- Check for connection leaks
- Review goroutine count

## Support

For issues or questions:
- GitHub Issues: https://github.com/App Vault/app-vault/issues
- Documentation: https://github.com/App Vault/app-vault
