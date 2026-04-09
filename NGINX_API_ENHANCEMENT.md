# Nginx API v1 Enhancement

## Overview
Enhanced nginx configuration to properly route `/api/v1/` endpoints to the backend with optimized performance, security, and reliability features.

## Changes Made

### 1. **Upstream Backend Pool**
- Created upstream configuration for connection pooling
- Added health checks with configurable failover
- Keepalive connections (32) for performance

```nginx
upstream appvault_backend {
    server appvault:8089 max_fails=3 fail_timeout=30s;
    keepalive 32;
}
```

### 2. **Rate Limiting Configuration**
Implemented three rate-limiting zones:
- **API General**: 100 requests/second with 20-request burst
- **Auth Endpoints**: 20 requests/minute (strict for security)
- **Connection Limits**: Max 10 concurrent connections per IP

```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/s;
limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=20r/m;
limit_conn_zone $binary_remote_addr zone=conn_limit:10m;
```

### 3. **API v1 Routing**
Two separate location blocks for granular control:

#### Authentication Endpoints (`/api/v1/auth/`)
- Stricter rate limiting (20 req/min)
- Connection limit of 5 concurrent connections
- Shorter timeouts for auth security

#### General API Endpoints (`/api/v1/`)
- Standard rate limiting (100 req/s)
- Connection limit of 10 concurrent connections
- Longer timeouts for data transfer

### 4. **Performance Optimizations**
- **Connection Pooling**: Reuses connections to backend
- **Buffering**: Optimized for small to large responses (up to 2GB)
- **Gzip Compression**: Level 6 compression on all text/JSON content
- **Error Handling**: Automatic retry on backend failures (2 attempts, 10s timeout)
- **Timeouts**:
  - Connect: 10-30 seconds
  - Send/Read: 30-60 seconds

### 5. **Security Headers**
- All existing security headers maintained
- Added request ID tracking (`X-Request-ID`)
- Hidden file protection (deny access to `/.git`, `/.env`, etc.)

### 6. **Proxy Headers**
Complete client information forwarding:
- `X-Real-IP`: Original client IP
- `X-Forwarded-For`: Proxy chain
- `X-Forwarded-Proto`: Original protocol
- `X-Forwarded-Host`: Original host
- `X-Request-ID`: Unique request tracking

### 7. **Asset Caching Strategy**
- **Static Assets**: 1-year cache with immutable flag
- **index.html**: No cache (always check for updates)
- **Hidden Files**: Blocked from access

## Access Points

### Frontend
- **URL**: `http://localhost`
- **Port**: 80 (HTTP), 443 (HTTPS)

### API v1
- **Base URL**: `http://localhost/api/v1`
- **Examples**:
  - `POST http://localhost/api/v1/auth/login`
  - `POST http://localhost/api/v1/auth/register`
  - `GET http://localhost/api/v1/secrets`

### Health Checks
- **Endpoint**: `http://localhost/health`
- **Response**: `{"status":"healthy","database":"connected"}`

## Docker Configuration Updates

### Service Changes
1. Uncommented PostgreSQL service with `with-postgres` profile
2. Added `depends_on` for appvault service with health check
3. Fixed health check port from 8888 to 8089

### Running the System

**With Local PostgreSQL**:
```bash
docker compose --profile with-postgres up -d
```

**Without Local Database** (using external DB):
```bash
# Update .env with DATABASE_URL
docker compose up -d
```

## File Changes

| File | Change |
|------|--------|
| `web/nginx.conf` | Enhanced with upstream pools, rate limiting, and optimized routing |
| `docker-compose.yml` | Uncommented postgres service, added depends_on |

## Performance Metrics

- **Upstream Failover**: Max 3 failures before marking server down (30s timeout)
- **Rate Limit Burst**: 20 requests allowed above limit before rejection
- **Buffer Size**: Up to 16-64KB per request, 2GB maximum temp file
- **Keepalive**: 32 connections per pool for connection reuse
- **Compression**: All responses >1KB compressed with gzip

## Security Features

- **Rate Limiting**: Protects against brute force and DoS attacks
- **Connection Limits**: Prevents resource exhaustion
- **Security Headers**: HSTS, CSP, X-Frame-Options, etc.
- **Hidden Files**: `.gitignore`, `.env`, etc. blocked
- **Request Tracking**: Unique ID for all requests for logging/debugging

## Testing

### Test the API
```bash
# Health check
curl http://localhost/health

# Test auth endpoint
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# Test secrets endpoint
curl http://localhost/api/v1/secrets \
  -H "Authorization: Bearer {token}"
```

### Monitor Logs
```bash
# nginx logs
docker logs appvault-web --tail 50 -f

# backend logs
docker logs appvault-server --tail 50 -f

# database logs
docker logs appvault-postgres --tail 50 -f
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| 502 Bad Gateway | Check backend is running: `docker logs appvault-server` |
| 404 Not Found | Verify endpoint exists and uses `/api/v1` prefix |
| Rate limit exceeded | Wait for rate limit window to reset or increase limits |
| CORS errors | Update CSP header in nginx.conf if needed |
| High latency | Check buffer settings and increase timeouts if necessary |

## Future Enhancements

- [ ] Add SSL/TLS termination at nginx
- [ ] Implement request/response caching
- [ ] Add distributed rate limiting (across multiple nginx instances)
- [ ] Metrics collection (Prometheus integration)
- [ ] Request logging to ELK stack
- [ ] Load balancing across multiple backend instances

---

**Last Updated**: 2026-04-09
**Status**: ✅ Active and Healthy
