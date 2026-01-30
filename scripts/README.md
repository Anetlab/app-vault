# App Vault Scripts

This directory contains utility scripts for App Vault deployment and maintenance.

## Certificate Generation

### `generate-certs.ps1`

PowerShell script to generate TLS certificates for App Vault.

#### Prerequisites

- PowerShell 5.1 or later
- OpenSSL (optional but recommended for PEM format)
  - Windows: `choco install openssl` or download from [OpenSSL for Windows](https://slproweb.com/products/Win32OpenSSL.html)
  - macOS: `brew install openssl`
  - Linux: Usually pre-installed

#### Usage

**Development (Self-Signed Certificate)**

```powershell
# Generate certificate for localhost
.\scripts\generate-certs.ps1

# Specify custom domain and validity
.\scripts\generate-certs.ps1 -Domain "dev.appvault.local" -DaysValid 90

# Generate for specific environment
.\scripts\generate-certs.ps1 -Environment dev -Domain "localhost"
```

**Production (Certificate Signing Request)**

```powershell
# Generate CSR for CA signing
.\scripts\generate-certs.ps1 -Type csr -Domain "appvault.example.com" -Environment prod

# Submit the generated request.csr to your Certificate Authority
# Save the signed certificate as cert.pem in the certs directory
```

**Let's Encrypt**

```powershell
# Show Let's Encrypt instructions
.\scripts\generate-certs.ps1 -Type letsencrypt -Domain "appvault.example.com"
```

#### Parameters

| Parameter | Description | Default | Values |
|-----------|-------------|---------|--------|
| `-Environment` | Target environment | `dev` | `dev`, `staging`, `prod` |
| `-Domain` | Domain name for certificate | `localhost` | Any valid domain |
| `-OutputDir` | Certificate output directory | `.\certs` | Any valid path |
| `-DaysValid` | Certificate validity (days) | `365` | Any positive integer |
| `-Type` | Certificate type | `self-signed` | `self-signed`, `csr`, `letsencrypt` |

#### Output Files

The script generates the following files in the output directory:

**Self-Signed Certificate:**
- `cert.pem` - Certificate file (public key)
- `key.pem` - Private key file
- `certificate.pfx` - Windows certificate format (if PowerShell method used)
- `openssl.cnf` - OpenSSL configuration (if OpenSSL method used)

**Certificate Signing Request:**
- `key.pem` - Private key file (keep secure!)
- `request.csr` - Certificate signing request (submit to CA)
- `csr.cnf` - CSR configuration file

#### Examples

**Quick Start for Development**

```powershell
# Generate self-signed certificate
.\scripts\generate-certs.ps1

# The script will:
# - Create ./certs directory
# - Generate cert.pem and key.pem
# - Update .env file with certificate paths
```

**Production with Custom CA**

```powershell
# Step 1: Generate CSR and private key
.\scripts\generate-certs.ps1 -Type csr -Domain "appvault.company.com" -Environment prod

# Step 2: Submit certs/request.csr to your CA
# Step 3: Save signed certificate as certs/cert.pem

# Step 4: Update .env file
# ENABLE_TLS=true
# TLS_CERT_FILE=./certs/cert.pem
# TLS_KEY_FILE=./certs/key.pem
```

**Let's Encrypt (Recommended for Production)**

```powershell
# Install Certbot
choco install certbot

# Generate Let's Encrypt certificate
certbot certonly --standalone -d appvault.example.com

# Copy certificates
copy C:\Certbot\live\appvault.example.com\fullchain.pem .\certs\cert.pem
copy C:\Certbot\live\appvault.example.com\privkey.pem .\certs\key.pem

# Set up auto-renewal
certbot renew --dry-run
```

#### Security Notes

⚠️ **Important Security Considerations:**

1. **Private Keys**: Never commit `*.key` or `*.pem` files to version control
2. **Self-Signed Certificates**: Only use for development/testing
3. **Certificate Rotation**: Regularly rotate certificates (90 days for Let's Encrypt)
4. **File Permissions**: Restrict access to private keys (600/400 permissions on Unix)
5. **Production**: Always use trusted CA certificates in production

#### Troubleshooting

**OpenSSL Not Found**
```powershell
# Install OpenSSL on Windows
choco install openssl

# Or download from: https://slproweb.com/products/Win32OpenSSL.html
# Add to PATH: C:\Program Files\OpenSSL-Win64\bin
```

**PowerShell Execution Policy**
```powershell
# If script execution is blocked
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Or run once with bypass
powershell -ExecutionPolicy Bypass -File .\scripts\generate-certs.ps1
```

**Certificate Verification**
```powershell
# View certificate details (requires OpenSSL)
openssl x509 -in .\certs\cert.pem -noout -text

# Test TLS connection
curl -k https://localhost:8443/health

# Test with OpenSSL
openssl s_client -connect localhost:8443
```

#### Related Documentation

- [DEPLOYMENT.md](../DEPLOYMENT.md) - Production deployment guide
- [SECURITY.md](../SECURITY.md) - Security best practices
- [QUICKREF.md](../QUICKREF.md) - Quick reference commands

## Future Scripts

Additional scripts planned for this directory:

- `backup-db.ps1` - Database backup automation
- `restore-db.ps1` - Database restore utility
- `health-check.ps1` - Server health monitoring
- `rotate-keys.ps1` - Automated key rotation
- `setup-systemd.ps1` - Linux service setup
