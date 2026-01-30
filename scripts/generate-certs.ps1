#Requires -Version 5.1

<#
.SYNOPSIS
    Generate TLS certificates for App Vault

.DESCRIPTION
    This script generates TLS certificates for App Vault server.
    - For development: Creates self-signed certificates
    - For production: Generates CSR for CA signing or uses Let's Encrypt

.PARAMETER Environment
    Target environment: dev, staging, or prod

.PARAMETER Domain
    Domain name for the certificate (e.g., appvault.example.com)

.PARAMETER OutputDir
    Directory to store generated certificates (default: ./certs)

.PARAMETER DaysValid
    Certificate validity period in days (default: 365)

.PARAMETER Type
    Certificate type: self-signed, csr, or letsencrypt

.EXAMPLE
    .\generate-certs.ps1 -Environment dev
    
.EXAMPLE
    .\generate-certs.ps1 -Environment prod -Domain appvault.example.com -Type csr

.EXAMPLE
    .\generate-certs.ps1 -Domain localhost -DaysValid 90
#>

param(
    [Parameter()]
    [ValidateSet('dev', 'staging', 'prod')]
    [string]$Environment = 'dev',
    
    [Parameter()]
    [string]$Domain = 'localhost',
    
    [Parameter()]
    [string]$OutputDir = '.\certs',
    
    [Parameter()]
    [int]$DaysValid = 365,
    
    [Parameter()]
    [ValidateSet('self-signed', 'csr', 'letsencrypt')]
    [string]$Type = 'self-signed'
)

$ErrorActionPreference = 'Stop'

# Color output functions
function Write-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
}

function Write-Info {
    param([string]$Message)
    Write-Host "ℹ $Message" -ForegroundColor Cyan
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠ $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor Red
}

function Write-Header {
    param([string]$Message)
    Write-Host "`n═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host "  $Message" -ForegroundColor White
    Write-Host "═══════════════════════════════════════════════════════════════`n" -ForegroundColor Cyan
}

# Create output directory
function Initialize-CertDirectory {
    if (-not (Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
        Write-Success "Created certificate directory: $OutputDir"
    } else {
        Write-Info "Using existing certificate directory: $OutputDir"
    }
}

# Generate self-signed certificate using OpenSSL
function New-SelfSignedCertOpenSSL {
    param(
        [string]$Domain,
        [string]$OutputPath,
        [int]$Days
    )
    
    Write-Info "Generating self-signed certificate using OpenSSL..."
    
    # Check if OpenSSL is available
    $opensslPath = Get-Command openssl -ErrorAction SilentlyContinue
    if (-not $opensslPath) {
        Write-Error "OpenSSL not found. Please install OpenSSL or use -Type self-signed with PowerShell method."
        return $false
    }
    
    $keyFile = Join-Path $OutputPath "key.pem"
    $certFile = Join-Path $OutputPath "cert.pem"
    $configFile = Join-Path $OutputPath "openssl.cnf"
    
    # Create OpenSSL config with SAN
    $opensslConfig = @"
[req]
default_bits = 2048
prompt = no
default_md = sha256
distinguished_name = dn
req_extensions = v3_req

[dn]
C=US
ST=State
L=City
O=App Vault
OU=Development
CN=$Domain

[v3_req]
subjectAltName = @alt_names
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth

[alt_names]
DNS.1 = $Domain
DNS.2 = *.$Domain
DNS.3 = localhost
DNS.4 = 127.0.0.1
IP.1 = 127.0.0.1
IP.2 = ::1
"@
    
    Set-Content -Path $configFile -Value $opensslConfig
    
    # Generate private key and certificate
    $opensslArgs = @(
        'req', '-x509', '-newkey', 'rsa:2048', '-nodes',
        '-keyout', $keyFile,
        '-out', $certFile,
        '-days', $Days,
        '-config', $configFile
    )
    
    & openssl $opensslArgs 2>&1 | Out-Null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Certificate generated successfully"
        Write-Info "Certificate: $certFile"
        Write-Info "Private Key: $keyFile"
        
        # Display certificate info
        Write-Host "`nCertificate Details:" -ForegroundColor Yellow
        & openssl x509 -in $certFile -noout -subject -dates -ext subjectAltName
        
        return $true
    } else {
        Write-Error "Failed to generate certificate with OpenSSL"
        return $false
    }
}

# Generate self-signed certificate using PowerShell
function New-SelfSignedCertPowerShell {
    param(
        [string]$Domain,
        [string]$OutputPath,
        [int]$Days
    )
    
    Write-Info "Generating self-signed certificate using PowerShell..."
    
    $certFile = Join-Path $OutputPath "cert.pem"
    $keyFile = Join-Path $OutputPath "key.pem"
    $pfxFile = Join-Path $OutputPath "certificate.pfx"
    
    # Create certificate
    $notAfter = (Get-Date).AddDays($Days)
    $dnsNames = @($Domain, "localhost", "127.0.0.1")
    
    # Create self-signed certificate
    $cert = New-SelfSignedCertificate `
        -Subject "CN=$Domain" `
        -DnsName $dnsNames `
        -CertStoreLocation "Cert:\CurrentUser\My" `
        -KeyExportPolicy Exportable `
        -KeySpec Signature `
        -KeyLength 2048 `
        -HashAlgorithm SHA256 `
        -NotAfter $notAfter `
        -TextExtension @("2.5.29.37={text}1.3.6.1.5.5.7.3.1")
    
    # Export to PFX with password
    $password = ConvertTo-SecureString -String "appvault" -Force -AsPlainText
    Export-PfxCertificate -Cert $cert -FilePath $pfxFile -Password $password | Out-Null
    
    # Convert PFX to PEM format using OpenSSL (if available)
    $opensslPath = Get-Command openssl -ErrorAction SilentlyContinue
    if ($opensslPath) {
        # Export certificate
        & openssl pkcs12 -in $pfxFile -clcerts -nokeys -out $certFile -passin pass:appvault -passout pass: 2>&1 | Out-Null
        # Export private key
        & openssl pkcs12 -in $pfxFile -nocerts -nodes -out $keyFile -passin pass:appvault 2>&1 | Out-Null
        
        Write-Success "Certificate generated successfully"
        Write-Info "Certificate: $certFile"
        Write-Info "Private Key: $keyFile"
        Write-Info "PFX (Windows): $pfxFile (password: appvault)"
    } else {
        Write-Warning "OpenSSL not found. PFX file created but PEM conversion skipped."
        Write-Info "PFX File: $pfxFile (password: appvault)"
        Write-Info "Install OpenSSL to convert to PEM format for App Vault."
    }
    
    # Remove from certificate store
    Remove-Item -Path "Cert:\CurrentUser\My\$($cert.Thumbprint)" -Force
    
    Write-Host "`nCertificate Details:" -ForegroundColor Yellow
    Write-Host "  Subject: CN=$Domain" -ForegroundColor Gray
    Write-Host "  DNS Names: $($dnsNames -join ', ')" -ForegroundColor Gray
    Write-Host "  Valid From: $($cert.NotBefore)" -ForegroundColor Gray
    Write-Host "  Valid Until: $($cert.NotAfter)" -ForegroundColor Gray
    Write-Host "  Thumbprint: $($cert.Thumbprint)" -ForegroundColor Gray
    
    return $true
}

# Generate Certificate Signing Request (CSR)
function New-CertificateRequest {
    param(
        [string]$Domain,
        [string]$OutputPath
    )
    
    Write-Info "Generating Certificate Signing Request (CSR)..."
    
    # Check if OpenSSL is available
    $opensslPath = Get-Command openssl -ErrorAction SilentlyContinue
    if (-not $opensslPath) {
        Write-Error "OpenSSL is required for CSR generation. Please install OpenSSL."
        return $false
    }
    
    $keyFile = Join-Path $OutputPath "key.pem"
    $csrFile = Join-Path $OutputPath "request.csr"
    $configFile = Join-Path $OutputPath "csr.cnf"
    
    # Create CSR config
    $csrConfig = @"
[req]
default_bits = 2048
prompt = no
default_md = sha256
distinguished_name = dn
req_extensions = v3_req

[dn]
C=US
ST=State
L=City
O=App Vault
OU=Production
CN=$Domain
emailAddress=admin@$Domain

[v3_req]
subjectAltName = @alt_names
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth

[alt_names]
DNS.1 = $Domain
DNS.2 = www.$Domain
"@
    
    Set-Content -Path $configFile -Value $csrConfig
    
    # Generate private key
    Write-Info "Generating private key..."
    & openssl genrsa -out $keyFile 2048 2>&1 | Out-Null
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to generate private key"
        return $false
    }
    
    # Generate CSR
    Write-Info "Generating CSR..."
    & openssl req -new -key $keyFile -out $csrFile -config $configFile 2>&1 | Out-Null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "CSR generated successfully"
        Write-Info "Private Key: $keyFile"
        Write-Info "CSR File: $csrFile"
        
        Write-Host "`nCSR Details:" -ForegroundColor Yellow
        & openssl req -in $csrFile -noout -text | Select-String "Subject:", "DNS:"
        
        Write-Host "`n" -NoNewline
        Write-Warning "IMPORTANT: Keep the private key secure!"
        Write-Info "Submit the CSR ($csrFile) to your Certificate Authority"
        Write-Info "Once signed, save the certificate as 'cert.pem' in: $OutputPath"
        
        return $true
    } else {
        Write-Error "Failed to generate CSR"
        return $false
    }
}

# Generate Let's Encrypt instructions
function Show-LetsEncryptInstructions {
    param(
        [string]$Domain,
        [string]$OutputPath
    )
    
    Write-Header "Let's Encrypt Certificate Generation"
    
    Write-Info "To obtain a Let's Encrypt certificate, follow these steps:"
    Write-Host ""
    
    Write-Host "1. Install Certbot:" -ForegroundColor Yellow
    Write-Host "   # Windows (using Chocolatey)" -ForegroundColor Gray
    Write-Host "   choco install certbot" -ForegroundColor Gray
    Write-Host ""
    Write-Host "   # Or download from: https://certbot.eff.org/" -ForegroundColor Gray
    Write-Host ""
    
    Write-Host "2. Obtain Certificate (Standalone):" -ForegroundColor Yellow
    Write-Host "   certbot certonly --standalone -d $Domain" -ForegroundColor Gray
    Write-Host ""
    
    Write-Host "3. Or using DNS challenge:" -ForegroundColor Yellow
    Write-Host "   certbot certonly --manual --preferred-challenges dns -d $Domain" -ForegroundColor Gray
    Write-Host ""
    
    Write-Host "4. Certificate files will be in:" -ForegroundColor Yellow
    Write-Host "   C:\Certbot\live\$Domain\" -ForegroundColor Gray
    Write-Host ""
    
    Write-Host "5. Copy certificates to App Vault:" -ForegroundColor Yellow
    Write-Host "   copy C:\Certbot\live\$Domain\fullchain.pem $OutputPath\cert.pem" -ForegroundColor Gray
    Write-Host "   copy C:\Certbot\live\$Domain\privkey.pem $OutputPath\key.pem" -ForegroundColor Gray
    Write-Host ""
    
    Write-Host "6. Set up auto-renewal:" -ForegroundColor Yellow
    Write-Host "   certbot renew --dry-run" -ForegroundColor Gray
    Write-Host ""
    
    Write-Info "Let's Encrypt certificates are valid for 90 days"
    Write-Warning "Remember to set up automatic renewal!"
}

# Update .env file with certificate paths
function Update-EnvFile {
    param(
        [string]$CertPath,
        [string]$KeyPath
    )
    
    $envFile = ".\.env"
    $envExample = ".\.env.example"
    
    if (Test-Path $envFile) {
        $content = Get-Content $envFile -Raw
        
        if ($content -match 'TLS_CERT_FILE=') {
            $content = $content -replace 'TLS_CERT_FILE=.*', "TLS_CERT_FILE=$CertPath"
        } else {
            $content += "`nTLS_CERT_FILE=$CertPath"
        }
        
        if ($content -match 'TLS_KEY_FILE=') {
            $content = $content -replace 'TLS_KEY_FILE=.*', "TLS_KEY_FILE=$KeyPath"
        } else {
            $content += "`nTLS_KEY_FILE=$KeyPath"
        }
        
        if ($content -match 'ENABLE_TLS=') {
            $content = $content -replace 'ENABLE_TLS=.*', 'ENABLE_TLS=true'
        } else {
            $content += "`nENABLE_TLS=true"
        }
        
        Set-Content -Path $envFile -Value $content
        Write-Success "Updated .env file with certificate paths"
    } else {
        Write-Warning ".env file not found. Please create one from .env.example"
    }
}

# Main execution
Write-Header "App Vault - TLS Certificate Generator"

Write-Info "Environment: $Environment"
Write-Info "Domain: $Domain"
Write-Info "Certificate Type: $Type"
Write-Info "Output Directory: $OutputDir"

# Initialize directory
Initialize-CertDirectory

# Resolve full output path
$fullOutputPath = (Resolve-Path $OutputDir).Path

# Generate certificates based on type
$success = $false

switch ($Type) {
    'self-signed' {
        # Try OpenSSL first, fallback to PowerShell
        $opensslPath = Get-Command openssl -ErrorAction SilentlyContinue
        if ($opensslPath) {
            $success = New-SelfSignedCertOpenSSL -Domain $Domain -OutputPath $fullOutputPath -Days $DaysValid
        } else {
            Write-Warning "OpenSSL not found. Using PowerShell method (PEM conversion limited)"
            $success = New-SelfSignedCertPowerShell -Domain $Domain -OutputPath $fullOutputPath -Days $DaysValid
        }
    }
    'csr' {
        $success = New-CertificateRequest -Domain $Domain -OutputPath $fullOutputPath
    }
    'letsencrypt' {
        Show-LetsEncryptInstructions -Domain $Domain -OutputPath $fullOutputPath
        $success = $true
    }
}

if ($success) {
    Write-Host ""
    Write-Header "Certificate Generation Complete"
    
    if ($Type -ne 'letsencrypt') {
        $certFile = Join-Path $fullOutputPath "cert.pem"
        $keyFile = Join-Path $fullOutputPath "key.pem"
        
        if ((Test-Path $certFile) -and (Test-Path $keyFile)) {
            # Update .env file
            $relativeCertPath = Resolve-Path -Relative $certFile
            $relativeKeyPath = Resolve-Path -Relative $keyFile
            Update-EnvFile -CertPath $relativeCertPath -KeyPath $relativeKeyPath
            
            Write-Host "`nNext Steps:" -ForegroundColor Yellow
            Write-Host "  1. Review the generated certificates in: $fullOutputPath" -ForegroundColor Gray
            Write-Host "  2. Update your .env file with ENABLE_TLS=true" -ForegroundColor Gray
            Write-Host "  3. Configure TLS_CERT_FILE and TLS_KEY_FILE paths" -ForegroundColor Gray
            Write-Host "  4. Restart App Vault server" -ForegroundColor Gray
            Write-Host ""
            Write-Host "Start server with TLS:" -ForegroundColor Yellow
            Write-Host "  .\bin\appvault.exe" -ForegroundColor Gray
            Write-Host ""
            Write-Host "Test TLS connection:" -ForegroundColor Yellow
            Write-Host "  curl -k https://$Domain`:8443/health" -ForegroundColor Gray
            Write-Host ""
            
            if ($Type -eq 'self-signed') {
                Write-Warning "Self-signed certificates are for development only!"
                Write-Info "For production, use a trusted CA or Let's Encrypt"
            }
        }
    }
    
    Write-Success "All done!"
} else {
    Write-Error "Certificate generation failed"
    exit 1
}
