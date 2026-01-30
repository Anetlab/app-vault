# Testing Guide

This guide covers testing procedures for SecureVault, including unit tests, integration tests, and production validation.

## Table of Contents

- [Running Tests](#running-tests)
- [Load Testing](#load-testing)
- [Security Testing](#security-testing)
- [TLS Verification](#tls-verification)
- [Rate Limiting Tests](#rate-limiting-tests)
- [Metrics Validation](#metrics-validation)
- [API Testing](#api-testing)

## Running Tests

### Unit Tests

Run all unit tests:

```bash
go test ./...
```

Run with coverage:

```bash
go test -cover ./...
```

Generate coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

Run specific package:

```bash
go test ./internal/crypto -v
```

### Integration Tests

Integration tests require a running PostgreSQL instance.

```bash
# Set test database URL
export DATABASE_URL="postgres://test_user:test_pass@localhost:5432/securevault_test?sslmode=disable"

# Run migrations
go run cmd/migrate/main.go

# Run integration tests
go test -tags=integration ./...
```

## Load Testing

### Using Apache Bench (ab)

Test basic endpoint:

```bash
ab -n 10000 -c 100 -H "Authorization: Bearer YOUR_JWT_TOKEN" \
   https://localhost:8080/api/v1/secrets
```

### Using hey

Install hey:

```bash
go install github.com/rakyll/hey@latest
```

Load test with authentication:

```bash
hey -n 10000 -c 100 -m GET \
    -H "Authorization: Bearer YOUR_JWT_TOKEN" \
    https://localhost:8080/api/v1/secrets
```

### Using k6

Create `load-test.js`:

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '1m', target: 50 },
    { duration: '30s', target: 0 },
  ],
};

const BASE_URL = 'https://localhost:8080';
const TOKEN = 'YOUR_JWT_TOKEN';

export default function () {
  const params = {
    headers: {
      'Authorization': `Bearer ${TOKEN}`,
      'Content-Type': 'application/json',
    },
  };

  const res = http.get(`${BASE_URL}/api/v1/secrets`, params);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
  });

  sleep(1);
}
```

Run k6 test:

```bash
k6 run load-test.js
```

## Security Testing

### OWASP ZAP

Automated security scan:

```bash
docker run -t owasp/zap2docker-stable zap-baseline.py \
    -t https://localhost:8080 \
    -r zap-report.html
```

### SQL Injection Testing

Test with sqlmap:

```bash
sqlmap -u "https://localhost:8080/api/v1/secrets?name=test" \
       --cookie="token=YOUR_JWT_TOKEN" \
       --level=5 --risk=3
```

Expected result: No SQL injection vulnerabilities.

### Authentication Testing

Test invalid tokens:

```bash
# No token
curl -X GET https://localhost:8080/api/v1/secrets

# Invalid token
curl -X GET https://localhost:8080/api/v1/secrets \
     -H "Authorization: Bearer invalid_token"

# Expired token
curl -X GET https://localhost:8080/api/v1/secrets \
     -H "Authorization: Bearer EXPIRED_TOKEN"
```

All should return 401 Unauthorized.

## TLS Verification

### SSL Labs Test

For public deployments, use SSL Labs:

```
https://www.ssllabs.com/ssltest/analyze.html?d=yourdomain.com
```

Expected grade: A or A+

### OpenSSL Testing

Test TLS connection:

```bash
openssl s_client -connect localhost:8080 -tls1_3
```

Verify:
- Protocol: TLSv1.3
- Cipher: Strong suite (AES-GCM or ChaCha20-Poly1305)
- Certificate chain valid

Test weak protocols (should fail):

```bash
# These should be rejected
openssl s_client -connect localhost:8080 -tls1_2
openssl s_client -connect localhost:8080 -tls1_1
openssl s_client -connect localhost:8080 -ssl3
```

### Certificate Validation

Check certificate expiry:

```bash
echo | openssl s_client -servername localhost -connect localhost:8080 2>/dev/null | \
      openssl x509 -noout -dates
```

Verify certificate chain:

```bash
openssl verify -CAfile certs/ca.crt certs/server.crt
```

## Rate Limiting Tests

### Manual Testing

Test rate limit with curl loop:

```bash
# Send 150 requests (should hit 100 limit)
for i in {1..150}; do
  curl -s -o /dev/null -w "%{http_code}\n" \
       -H "Authorization: Bearer YOUR_JWT_TOKEN" \
       https://localhost:8080/api/v1/secrets
done
```

Expected: First 100 return 200, remaining return 429.

### Check Rate Limit Headers

```bash
curl -i -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     https://localhost:8080/api/v1/secrets
```

Look for headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Window: 60
```

### Rate Limit Recovery

After hitting limit, wait for window to expire:

```bash
# Hit limit
for i in {1..101}; do curl -s https://localhost:8080/health > /dev/null; done

# Wait 61 seconds
sleep 61

# Should work again
curl https://localhost:8080/health
```

### Load Test Rate Limiter

Using hey to test rate limiter performance:

```bash
# Should trigger rate limiting
hey -n 1000 -c 10 -q 20 https://localhost:8080/health
```

Check metrics for rate limit rejections.

## Metrics Validation

### Health Check

```bash
curl https://localhost:8080/health
```

Expected:
```json
{
  "status": "healthy",
  "database": "connected"
}
```

### Metrics Endpoint

```bash
curl https://localhost:8080/metrics
```

Verify metrics present:
- `securevault_requests_total`
- `securevault_auth_success_total`
- `securevault_secrets_created_total`
- `securevault_uptime_seconds`
- `securevault_memory_alloc_bytes`

### Prometheus Integration

Test Prometheus can scrape:

```bash
# Add to prometheus.yml
scrape_configs:
  - job_name: 'securevault'
    static_configs:
      - targets: ['localhost:8080']

# Check Prometheus targets page
curl http://localhost:9090/targets
```

### Grafana Dashboard

Import dashboard and verify:
- Request rate graph
- Error rate graph
- P95 latency
- Memory usage
- Active connections

## API Testing

### Authentication Flow

Register new user:

```bash
curl -X POST https://localhost:8080/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{
       "username": "testuser",
       "email": "test@example.com",
       "password": "SecurePass123!",
       "mfa_enabled": false
     }'
```

Login:

```bash
TOKEN=$(curl -X POST https://localhost:8080/api/v1/auth/login \
             -H "Content-Type: application/json" \
             -d '{
               "username": "testuser",
               "password": "SecurePass123!"
             }' | jq -r '.token')
```

### Secret Management

Create secret:

```bash
curl -X POST https://localhost:8080/api/v1/secrets \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "name": "test_secret",
       "value": "sensitive_data",
       "expires_in_days": 30,
       "tags": ["production", "api-key"]
     }'
```

Retrieve secret:

```bash
curl -X GET https://localhost:8080/api/v1/secrets/test_secret \
     -H "Authorization: Bearer $TOKEN"
```

List secrets:

```bash
curl -X GET https://localhost:8080/api/v1/secrets \
     -H "Authorization: Bearer $TOKEN"
```

Delete secret:

```bash
curl -X DELETE https://localhost:8080/api/v1/secrets/test_secret \
     -H "Authorization: Bearer $TOKEN"
```

### Service Principal Testing

Create service principal:

```bash
curl -X POST https://localhost:8080/api/v1/service-principals \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "name": "ci_pipeline",
       "permissions": ["secrets:read"],
       "rate_limit": 1000
     }'
```

Expected response includes `client_secret` (save it).

Authenticate as service principal:

```bash
SP_TOKEN=$(curl -X POST https://localhost:8080/api/v1/auth/sp-login \
                -H "Content-Type: application/json" \
                -d '{
                  "client_id": "CLIENT_ID_FROM_RESPONSE",
                  "client_secret": "SECRET_FROM_RESPONSE"
                }' | jq -r '.token')
```

Test permissions:

```bash
# Should work (has secrets:read)
curl -X GET https://localhost:8080/api/v1/secrets \
     -H "Authorization: Bearer $SP_TOKEN"

# Should fail (doesn't have secrets:create)
curl -X POST https://localhost:8080/api/v1/secrets \
     -H "Authorization: Bearer $SP_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"name": "test", "value": "data"}'
```

### Key Rotation Testing

Trigger key rotation:

```bash
curl -X POST https://localhost:8080/api/v1/vault/rotate-key \
     -H "Authorization: Bearer $TOKEN"
```

Verify old secrets still accessible:

```bash
curl -X GET https://localhost:8080/api/v1/secrets/test_secret \
     -H "Authorization: Bearer $TOKEN"
```

Create new secret with new key:

```bash
curl -X POST https://localhost:8080/api/v1/secrets \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "name": "new_secret",
       "value": "encrypted_with_new_key"
     }'
```

## Performance Benchmarks

### Expected Performance

Under normal load:
- p50 latency: < 50ms
- p95 latency: < 150ms
- p99 latency: < 300ms
- Throughput: > 1000 req/s (single instance)

### Crypto Performance

Benchmark encryption:

```bash
go test -bench=BenchmarkEncrypt -benchmem ./internal/crypto
```

Expected: > 10,000 ops/sec

### Database Performance

Benchmark database queries:

```bash
go test -bench=BenchmarkDatabase -benchmem ./internal/db
```

Monitor with:
```sql
SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      
      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...
        env:
          DATABASE_URL: postgres://postgres:test@localhost:5432/securevault_test?sslmode=disable
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

## Security Checklist

Before deploying to production:

- [ ] All unit tests pass
- [ ] Integration tests pass
- [ ] Load tests meet performance requirements
- [ ] TLS configuration verified (A+ rating)
- [ ] Rate limiting works as expected
- [ ] Metrics endpoint accessible
- [ ] Health check endpoint works
- [ ] No SQL injection vulnerabilities
- [ ] Authentication cannot be bypassed
- [ ] Service principals work with correct permissions
- [ ] Key rotation doesn't break existing secrets
- [ ] Audit logs capture all operations
- [ ] Error messages don't leak sensitive information
- [ ] Dependencies scanned for vulnerabilities
- [ ] Code reviewed by security team

## Troubleshooting Tests

### Tests Fail to Connect to Database

```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Check connection string
echo $DATABASE_URL

# Test connection manually
psql $DATABASE_URL -c "SELECT 1"
```

### Load Tests Show High Latency

- Check database connection pool settings
- Monitor database queries with EXPLAIN ANALYZE
- Check for slow queries in pg_stat_statements
- Review application logs for errors
- Monitor system resources (CPU, memory, I/O)

### Rate Limiting Not Working

- Verify middleware is enabled in main.go
- Check rate limit configuration in .env
- Inspect X-RateLimit-* headers in responses
- Monitor rate limit cache size
- Check if using correct client IP (X-Forwarded-For)

### Metrics Missing

- Ensure metrics endpoint is accessible
- Check CORS settings for cross-origin access
- Verify Prometheus scrape configuration
- Check firewall rules
- Review application logs for errors
