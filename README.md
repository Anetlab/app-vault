# SecureVault

A secure key management service inspired by Azure Key Vault and 1Password, written in Go with PostgreSQL storage. SecureVault provides enterprise-grade secret management with strong encryption and multiple authentication methods.

## Features

- **Strong Encryption**: XChaCha20-Poly1305 AEAD encryption for all secrets
- **Zero-Knowledge Architecture**: Server never sees unencrypted secrets
- **Key Derivation**: Argon2id with secure parameters for master key derivation
- **Vault Key Management**: Per-user encryption keys with rotation support
- **Dual Authentication**: User accounts and service principals (client credentials)
- **Key Rotation**: Automatic re-encryption during vault key rotation
- **Audit Logging**: Complete audit trail of all operations
- **Secret Versioning**: Track changes to secrets over time
- **Expiration Support**: Set expiration dates for secrets
- **RESTful API**: Simple HTTP/JSON API

## Security Architecture

SecureVault implements a 1Password-inspired security model:

1. **Master Key**: Derived from `password + secretKey` using Argon2id
2. **Vault Key**: Per-user encryption key for secrets
3. **Key Encryption**: Vault Key encrypted with Key Encryption Key (KEK) derived from Master Key
4. **Secret Encryption**: All secrets encrypted with Vault Key using XChaCha20-Poly1305

### Key Derivation Parameters

- **Argon2id**: Time=3, Memory=64MB, Threads=4, KeyLength=32 bytes
- **HKDF**: SHA-256 for deriving Key Encryption Key
- **XChaCha20-Poly1305**: 192-bit nonce, 256-bit key

## Quick Start

### Prerequisites

- Go 1.23 or later
- PostgreSQL 12 or later

### Installation

1. Clone the repository:
```bash
git clone https://github.com/securevault/app-vault.git
cd app-vault
```

2. Install dependencies:
```bash
go mod download
```

3. Set up PostgreSQL:
```bash
createdb securevault
```

4. Configure environment variables:
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/securevault?sslmode=disable"
export JWT_SECRET="your-secret-key-change-this-in-production"
export SERVER_PORT="8080"
```

5. Run the server:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`.

## API Documentation

### Authentication

#### Register User

Creates a new user account and returns a secret key.

```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "secret_key": "A3-abcd1234-efgh5678-ijkl9012-mnop3456"
}
```

⚠️ **Important**: Save the `secret_key` securely. It cannot be recovered.

#### Login

Authenticates a user and returns a JWT token.

```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123",
  "secret_key": "A3-abcd1234-efgh5678-ijkl9012-mnop3456"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com"
}
```

### Secret Management

All secret operations require:
- `Authorization: Bearer <token>` header
- `X-Vault-Password: <password>` header
- `X-Vault-Secret-Key: <secret_key>` header

#### Create Secret

Creates or updates a secret.

```bash
POST /api/v1/secrets
Authorization: Bearer <token>
X-Vault-Password: securepassword123
X-Vault-Secret-Key: A3-abcd1234-efgh5678-ijkl9012-mnop3456
Content-Type: application/json

{
  "name": "database-password",
  "value": "super-secret-password",
  "secret_type": "password",
  "tags": ["production", "database"],
  "expires_at": "2026-12-31T23:59:59Z"
}
```

**Response:**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "name": "database-password",
  "version": 1,
  "created_at": "2026-01-30T10:00:00Z"
}
```

#### Get Secret by Name

Retrieves a secret by name.

```bash
GET /api/v1/secrets/by-name?name=database-password
Authorization: Bearer <token>
X-Vault-Password: securepassword123
X-Vault-Secret-Key: A3-abcd1234-efgh5678-ijkl9012-mnop3456
```

**Response:**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "name": "database-password",
  "value": "super-secret-password",
  "secret_type": "password",
  "version": 1,
  "tags": ["production", "database"],
  "created_at": "2026-01-30T10:00:00Z",
  "updated_at": "2026-01-30T10:00:00Z",
  "expires_at": "2026-12-31T23:59:59Z",
  "access_count": 1,
  "last_accessed_at": "2026-01-30T10:05:00Z"
}
```

#### Get Secret by ID

Retrieves a secret by ID.

```bash
GET /api/v1/secrets/660e8400-e29b-41d4-a716-446655440001
Authorization: Bearer <token>
X-Vault-Password: securepassword123
X-Vault-Secret-Key: A3-abcd1234-efgh5678-ijkl9012-mnop3456
```

#### List Secrets

Lists all secrets (without values).

```bash
GET /api/v1/secrets
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "name": "database-password",
    "secret_type": "password",
    "version": 1,
    "tags": ["production", "database"],
    "created_at": "2026-01-30T10:00:00Z",
    "updated_at": "2026-01-30T10:00:00Z",
    "expires_at": "2026-12-31T23:59:59Z",
    "access_count": 5,
    "last_accessed_at": "2026-01-30T12:00:00Z"
  }
]
```

#### Delete Secret

Deletes a secret.

```bash
DELETE /api/v1/secrets/660e8400-e29b-41d4-a716-446655440001
Authorization: Bearer <token>
```

**Response:** `204 No Content`

### Key Management

#### Rotate Vault Key

Rotates the vault key and re-encrypts all secrets.

```bash
POST /api/v1/keys/rotate
Authorization: Bearer <token>
Content-Type: application/json

{
  "password": "securepassword123",
  "secret_key": "A3-abcd1234-efgh5678-ijkl9012-mnop3456",
  "reason": "scheduled rotation"
}
```

**Response:**
```json
{
  "new_key_version_id": "770e8400-e29b-41d4-a716-446655440002",
  "secrets_reencrypted": 42,
  "duration_ms": 1234
}
```

#### Get Key Status

Retrieves key version history.

```bash
GET /api/v1/keys/status
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": "770e8400-e29b-41d4-a716-446655440002",
    "version": 2,
    "status": "active",
    "created_at": "2026-01-30T10:00:00Z",
    "rotation_reason": "scheduled rotation"
  },
  {
    "id": "770e8400-e29b-41d4-a716-446655440001",
    "version": 1,
    "status": "deprecated",
    "created_at": "2026-01-01T00:00:00Z",
    "destroy_at": "2026-01-31T10:00:00Z",
    "rotation_reason": "initial key"
  }
]
```

#### Change Password

Changes the user's password (master key rotation).

```bash
POST /api/v1/auth/change-password
Authorization: Bearer <token>
Content-Type: application/json

{
  "old_password": "securepassword123",
  "new_password": "newsecurepassword456",
  "secret_key": "A3-abcd1234-efgh5678-ijkl9012-mnop3456"
}
```

**Response:**
```json
{
  "success": true
}
```

## Secret Types

- `generic` - Generic secret (default)
- `password` - Password
- `api_key` - API key
- `certificate` - Certificate
- `connection` - Connection string

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:postgres@localhost:5432/securevault?sslmode=disable` |
| `JWT_SECRET` | Secret key for JWT signing | `your-secret-key-change-this-in-production` |
| `SERVER_PORT` | Server port | `8080` |
| `MIGRATIONS_PATH` | Path to SQL migration files | `migrations` |

## Development

### Run Tests

```bash
go test ./...
```

### Build

```bash
go build -o bin/securevault cmd/server/main.go
```

### Run Migrations Manually

Migrations are automatically run on startup. To run them manually:

```bash
psql -U postgres -d securevault -f migrations/001_initial_schema.sql
```

## Security Considerations

1. **Change JWT Secret**: Always use a strong, random JWT secret in production
2. **Use TLS**: Deploy behind a TLS terminator (nginx, load balancer)
3. **Secure Database**: Use strong PostgreSQL passwords and restrict network access
4. **Rate Limiting**: Consider adding rate limiting for production deployments
5. **Backup Secret Key**: Users must securely store their secret key
6. **Key Rotation**: Implement regular vault key rotation policies
7. **Audit Logs**: Monitor audit logs for suspicious activity

## Architecture

```
┌─────────────┐
│  Client     │
└──────┬──────┘
       │ (HTTPS)
       ▼
┌─────────────────┐
│   API Layer     │  ← JWT Auth, Rate Limiting
└──────┬──────────┘
       │
┌──────▼──────────┐
│ Service Layer   │  ← Business Logic
│ - Auth          │
│ - Vault         │
│ - Rotation      │
└──────┬──────────┘
       │
┌──────▼──────────┐
│  Crypto Layer   │  ← Encryption/Decryption
└──────┬──────────┘
       │
┌──────▼──────────┐
│   Database      │  ← PostgreSQL
│   (Encrypted)   │
└─────────────────┘
```

## Database Schema

### Tables

- **users** - User accounts with encrypted vault keys
- **service_principals** - Application credentials for API access
- **key_versions** - Key rotation history
- **secrets** - Encrypted secrets with versioning
- **audit_logs** - Complete audit trail
- **rate_limits** - Rate limiting state

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please read the contributing guidelines before submitting pull requests.

## Support

For issues and questions, please open an issue on GitHub.
