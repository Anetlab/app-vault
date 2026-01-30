# Security

## Reporting Security Vulnerabilities

**Please do not report security vulnerabilities through public GitHub issues.**

If you discover a security vulnerability, please send an email to security@example.com with:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

We will acknowledge your email within 48 hours and provide a detailed response within 7 days.

## Security Features

### Encryption

**Data at Rest:**
- All secrets encrypted using XChaCha20-Poly1305 AEAD
- 256-bit encryption keys derived from master key
- 192-bit random nonces (never reused)
- Authenticated encryption prevents tampering

**Key Management:**
- Master key stored in environment variable (not in database)
- Key versioning supports rotation without data loss
- Old keys retained for decryption of existing secrets
- Keys derived using Argon2id for password-based operations

**Argon2id Parameters:**
- Time cost: 3 iterations
- Memory: 64 MB
- Parallelism: 4 threads
- Output: 32 bytes

### Authentication

**JWT Tokens:**
- HMAC-SHA256 signatures
- User tokens: 24-hour expiry
- Service principal tokens: 1-hour expiry
- Token claims include user ID, type, permissions
- Tokens validated on every request

**Multi-Factor Authentication:**
- TOTP-based (Time-based One-Time Password)
- 6-digit codes, 30-second window
- Compatible with Google Authenticator, Authy
- Optional per-user enablement

**Service Principals:**
- Client ID and client secret authentication
- Scoped permissions (e.g., secrets:read, secrets:create)
- Rate limits configurable per service principal
- Secrets hashed with Argon2id before storage

### Transport Security

**TLS Configuration:**
- TLS 1.3 minimum version
- Strong cipher suites only:
  - TLS_AES_256_GCM_SHA384
  - TLS_AES_128_GCM_SHA256
  - TLS_CHACHA20_POLY1305_SHA256
- Curve preferences: X25519, P-256
- Perfect Forward Secrecy (PFS)

**Security Headers:**
- `Strict-Transport-Security`: 1-year HSTS with subdomains
- `X-Frame-Options`: DENY (clickjacking protection)
- `X-Content-Type-Options`: nosniff
- `X-XSS-Protection`: 1; mode=block
- `Content-Security-Policy`: Restrictive policy
- `Referrer-Policy`: no-referrer
- `Permissions-Policy`: Disables unnecessary features

### Rate Limiting

**Protection Against Abuse:**
- Per-IP rate limiting with sliding window
- Configurable limits (default: 100 requests/minute)
- HTTP 429 responses when exceeded
- Custom limits for service principals
- Automatic cleanup of stale entries

**Implementation:**
- In-memory cache for performance
- X-Forwarded-For support for reverse proxies
- Rate limit headers on every response
- No shared state (stateless across instances)

### Input Validation

**Request Validation:**
- Maximum request body size (default: 1 MB)
- JSON schema validation
- SQL injection prevention via parameterized queries
- Path traversal prevention
- Strict username/email format validation

**Data Sanitization:**
- HTML escaping in error messages
- No user input in SQL queries
- Type-safe database operations
- Bounds checking on array access

### Database Security

**Connection Security:**
- SSL/TLS connections required in production
- Connection pooling with limits
- Prepared statements prevent SQL injection
- Row-level security policies

**Password Storage:**
- Argon2id hashing (never store plaintext)
- Unique salt per password
- Constant-time comparison prevents timing attacks

**Audit Logging:**
- All operations logged to `audit_logs` table
- Includes user ID, action, timestamp, IP address
- Immutable records (no updates, only inserts)
- Retention policy configurable

### Authorization

**Role-Based Access Control (RBAC):**
- User roles: regular, admin
- Service principal permissions: granular scopes
- Resource ownership checks
- Permission validation on every request

**Secret Access Control:**
- Users can only access their own secrets
- Service principals require explicit permissions
- Admin users have full access
- Audit trail for all access

## Security Best Practices

### Deployment

1. **Never commit secrets to version control**
   - Use `.env` files (excluded by `.gitignore`)
   - Use environment variables in production
   - Rotate secrets if accidentally committed

2. **Use strong, unique passwords**
   - Minimum 12 characters
   - Mix of uppercase, lowercase, numbers, symbols
   - Consider passphrase-based passwords

3. **Enable TLS in production**
   - Use certificates from trusted CA (Let's Encrypt)
   - Configure HSTS header
   - Test with SSL Labs

4. **Protect the master encryption key**
   - Store in secure environment variable
   - Never log or expose in errors
   - Rotate periodically (requires key rotation)
   - Consider using KMS (AWS KMS, Azure Key Vault)

5. **Harden database access**
   - Use SSL connections
   - Restrict network access
   - Use dedicated database user with minimal privileges
   - Regular backups with encryption

6. **Configure reverse proxy properly**
   - Set X-Forwarded-For correctly
   - Enable rate limiting at proxy level too
   - Restrict metrics endpoint access
   - Use fail2ban for brute force protection

7. **Monitor and alert**
   - High error rates
   - Failed authentication attempts
   - Rate limit violations
   - Database connection issues
   - Certificate expiry warnings

### Development

1. **Keep dependencies updated**
   ```bash
   go get -u ./...
   go mod tidy
   ```

2. **Run security scanners**
   ```bash
   # Vulnerability scanning
   go install golang.org/x/vuln/cmd/govulncheck@latest
   govulncheck ./...

   # Static analysis
   go install github.com/securego/gosec/v2/cmd/gosec@latest
   gosec ./...
   ```

3. **Review code for security issues**
   - SQL injection risks
   - Path traversal vulnerabilities
   - Insecure cryptography
   - Hardcoded secrets
   - Information leakage in errors

4. **Write security tests**
   - Test authentication bypass attempts
   - Test authorization boundaries
   - Test input validation
   - Test rate limiting

## Vulnerability Disclosure Policy

We are committed to working with security researchers to identify and resolve security issues.

**Our Promise:**
- We will respond to your report within 48 hours
- We will keep you informed of our progress
- We will credit you for responsible disclosure (if desired)
- We will not pursue legal action for good faith research

**We Ask That You:**
- Give us reasonable time to fix issues before public disclosure
- Avoid privacy violations, data destruction, or service disruption
- Only test against your own accounts or with explicit permission

## Security Audit History

### Internal Security Review (January 2024)

**Scope:** Full application security review

**Findings:**
- ✅ Encryption implementation follows best practices
- ✅ Authentication mechanisms are secure
- ✅ Rate limiting effectively prevents abuse
- ✅ No SQL injection vulnerabilities found
- ✅ TLS configuration meets industry standards

**Recommendations Implemented:**
- Added request size limits
- Enhanced security headers
- Improved audit logging
- Added service principal rate limits

### Pending Audits

- [ ] Third-party penetration test (Q2 2024)
- [ ] Cryptographic review by independent expert
- [ ] OWASP compliance audit
- [ ] GDPR compliance review

## Compliance

### OWASP Top 10 (2021)

| Risk | Status | Mitigation |
|------|--------|------------|
| A01: Broken Access Control | ✅ Protected | RBAC, ownership checks, audit logs |
| A02: Cryptographic Failures | ✅ Protected | XChaCha20-Poly1305, Argon2id, TLS 1.3 |
| A03: Injection | ✅ Protected | Parameterized queries, input validation |
| A04: Insecure Design | ✅ Protected | Security-first architecture, threat modeling |
| A05: Security Misconfiguration | ✅ Protected | Secure defaults, security headers |
| A06: Vulnerable Components | ✅ Protected | Regular updates, vulnerability scanning |
| A07: Authentication Failures | ✅ Protected | JWT, MFA support, rate limiting |
| A08: Data Integrity Failures | ✅ Protected | AEAD encryption, audit logs |
| A09: Logging Failures | ✅ Protected | Comprehensive audit logging |
| A10: SSRF | ✅ Protected | No external requests based on user input |

### CIS Controls

Implemented CIS Controls:
- **Control 3:** Data Protection - Encryption at rest and in transit
- **Control 4:** Secure Configuration - Hardened defaults
- **Control 6:** Access Control - RBAC and authentication
- **Control 8:** Audit Log Management - Comprehensive logging
- **Control 14:** Security Monitoring - Metrics and alerts

## Known Limitations

1. **In-Memory Rate Limiting**
   - Does not persist across restarts
   - Not shared across multiple instances
   - **Mitigation:** Use Redis for distributed rate limiting in production

2. **Master Key in Environment**
   - Single point of failure
   - No automatic rotation
   - **Mitigation:** Consider integrating with KMS services

3. **No Built-in Key Backup**
   - Key rotation requires manual key preservation
   - **Mitigation:** Implement key backup procedures

4. **Service Principal Secrets**
   - Cannot be recovered if lost
   - **Mitigation:** Store securely during creation

## Security Checklist for Production

Before deploying to production:

### Configuration
- [ ] Strong JWT secret (32+ bytes, randomly generated)
- [ ] TLS enabled with valid certificates
- [ ] Database SSL connections enabled
- [ ] Rate limiting configured appropriately
- [ ] Max request size set to reasonable value
- [ ] Metrics endpoint access restricted

### Infrastructure
- [ ] Firewall rules configured (only necessary ports open)
- [ ] Database on private network
- [ ] Regular automated backups configured
- [ ] Backup encryption enabled
- [ ] Monitoring and alerting set up

### Application
- [ ] Latest security patches applied
- [ ] Dependency vulnerabilities scanned
- [ ] Security headers verified
- [ ] Audit logging enabled
- [ ] Error messages don't leak sensitive info

### Operations
- [ ] Incident response plan documented
- [ ] Key rotation procedures established
- [ ] Backup restoration tested
- [ ] Access controls reviewed
- [ ] Security team trained

## Additional Resources

- [OWASP Cheat Sheet Series](https://cheatsheetseries.owasp.org/)
- [Go Security Best Practices](https://github.com/golang/go/wiki/Security)
- [NIST Cryptographic Standards](https://csrc.nist.gov/publications)
- [CIS Controls](https://www.cisecurity.org/controls)

## Contact

For security concerns, please contact: security@example.com

For general questions: support@example.com
