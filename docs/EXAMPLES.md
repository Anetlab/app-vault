# App Vault API Examples

This document provides practical examples of using the App Vault API.

## Prerequisites

- App Vault server running on `http://localhost:8080`
- PostgreSQL database configured
- `curl` or similar HTTP client

## Complete Workflow Example

### 1. Register a New User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "SecurePassword123!"
  }'
```

**Response:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "secret_key": "A3-abcd1234-efgh5678-ijkl9012-mnop3456"
}
```

⚠️ **Save the `secret_key`** - you cannot recover it later!

### 2. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "SecurePassword123!",
    "secret_key": "A3-abcd1234-efgh5678-ijkl9012-mnop3456"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com"
}
```

Save the `token` for subsequent requests.

### 3. Create a Secret

```bash
curl -X POST http://localhost:8080/api/v1/secrets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "X-Vault-Password: SecurePassword123!" \
  -H "X-Vault-Secret-Key: A3-abcd1234-efgh5678-ijkl9012-mnop3456" \
  -d '{
    "name": "aws-access-key",
    "value": "AKIAIOSFODNN7EXAMPLE",
    "secret_type": "api_key",
    "tags": ["production", "aws"]
  }'
```

**Response:**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "name": "aws-access-key",
  "version": 1,
  "created_at": "2026-01-30T10:00:00Z"
}
```

### 4. Create Another Secret

```bash
curl -X POST http://localhost:8080/api/v1/secrets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "X-Vault-Password: SecurePassword123!" \
  -H "X-Vault-Secret-Key: A3-abcd1234-efgh5678-ijkl9012-mnop3456" \
  -d '{
    "name": "database-password",
    "value": "postgres123!secret",
    "secret_type": "password",
    "tags": ["production", "database"],
    "expires_at": "2026-12-31T23:59:59Z"
  }'
```

### 5. List All Secrets

```bash
curl -X GET http://localhost:8080/api/v1/secrets \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
[
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "name": "aws-access-key",
    "secret_type": "api_key",
    "version": 1,
    "tags": ["production", "aws"],
    "created_at": "2026-01-30T10:00:00Z",
    "updated_at": "2026-01-30T10:00:00Z",
    "access_count": 0
  },
  {
    "id": "660e8400-e29b-41d4-a716-446655440002",
    "name": "database-password",
    "secret_type": "password",
    "version": 1,
    "tags": ["production", "database"],
    "created_at": "2026-01-30T10:01:00Z",
    "updated_at": "2026-01-30T10:01:00Z",
    "expires_at": "2026-12-31T23:59:59Z",
    "access_count": 0
  }
]
```

### 6. Retrieve a Secret by Name

```bash
curl -X GET "http://localhost:8080/api/v1/secrets/by-name?name=aws-access-key" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "X-Vault-Password: SecurePassword123!" \
  -H "X-Vault-Secret-Key: A3-abcd1234-efgh5678-ijkl9012-mnop3456"
```

**Response:**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "name": "aws-access-key",
  "value": "AKIAIOSFODNN7EXAMPLE",
  "secret_type": "api_key",
  "version": 1,
  "tags": ["production", "aws"],
  "created_at": "2026-01-30T10:00:00Z",
  "updated_at": "2026-01-30T10:00:00Z",
  "access_count": 1,
  "last_accessed_at": "2026-01-30T10:05:00Z"
}
```

### 7. Retrieve a Secret by ID

```bash
curl -X GET http://localhost:8080/api/v1/secrets/660e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "X-Vault-Password: SecurePassword123!" \
  -H "X-Vault-Secret-Key: A3-abcd1234-efgh5678-ijkl9012-mnop3456"
```

### 8. Update a Secret (Creates New Version)

```bash
curl -X POST http://localhost:8080/api/v1/secrets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "X-Vault-Password: SecurePassword123!" \
  -H "X-Vault-Secret-Key: A3-abcd1234-efgh5678-ijkl9012-mnop3456" \
  -d '{
    "name": "aws-access-key",
    "value": "AKIAIOSFODNN7UPDATED",
    "secret_type": "api_key",
    "tags": ["production", "aws"]
  }'
```

**Response:**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440003",
  "name": "aws-access-key",
  "version": 2,
  "created_at": "2026-01-30T10:10:00Z"
}
```

### 9. Rotate Vault Key

```bash
curl -X POST http://localhost:8080/api/v1/keys/rotate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "password": "SecurePassword123!",
    "secret_key": "A3-abcd1234-efgh5678-ijkl9012-mnop3456",
    "reason": "quarterly rotation"
  }'
```

**Response:**
```json
{
  "new_key_version_id": "770e8400-e29b-41d4-a716-446655440002",
  "secrets_reencrypted": 2,
  "duration_ms": 145
}
```

### 10. Get Key Status

```bash
curl -X GET http://localhost:8080/api/v1/keys/status \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
[
  {
    "id": "770e8400-e29b-41d4-a716-446655440002",
    "version": 2,
    "status": "active",
    "created_at": "2026-01-30T10:15:00Z",
    "rotation_reason": "quarterly rotation"
  },
  {
    "id": "770e8400-e29b-41d4-a716-446655440001",
    "version": 1,
    "status": "deprecated",
    "created_at": "2026-01-30T09:00:00Z",
    "destroy_at": "2026-01-31T10:15:00Z",
    "rotation_reason": "initial key"
  }
]
```

### 11. Change Password

```bash
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "old_password": "SecurePassword123!",
    "new_password": "NewSecurePassword456!",
    "secret_key": "A3-abcd1234-efgh5678-ijkl9012-mnop3456"
  }'
```

**Response:**
```json
{
  "success": true
}
```

### 12. Delete a Secret

```bash
curl -X DELETE http://localhost:8080/api/v1/secrets/660e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:** `204 No Content`

## Environment Setup for Examples

Create a `.env` file:

```bash
SERVER_PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/App Vault?sslmode=disable
JWT_SECRET=my-super-secret-jwt-key-change-in-production
MIGRATIONS_PATH=migrations
```

Start PostgreSQL:

```bash
docker run -d --name App Vault-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=App Vault \
  -p 5432:5432 \
  postgres:15
```

Start the server:

```bash
go run cmd/server/main.go
```

## Using with HTTPie

HTTPie provides a more user-friendly syntax:

```bash
# Register
http POST :8080/api/v1/auth/register \
  email=alice@example.com \
  password=SecurePassword123!

# Login
http POST :8080/api/v1/auth/login \
  email=alice@example.com \
  password=SecurePassword123! \
  secret_key=A3-abcd1234-efgh5678-ijkl9012-mnop3456

# Create secret
http POST :8080/api/v1/secrets \
  Authorization:"Bearer TOKEN" \
  X-Vault-Password:SecurePassword123! \
  X-Vault-Secret-Key:A3-abcd1234-efgh5678-ijkl9012-mnop3456 \
  name=my-secret \
  value=secret-value

# Get secret
http GET :8080/api/v1/secrets/by-name?name=my-secret \
  Authorization:"Bearer TOKEN" \
  X-Vault-Password:SecurePassword123! \
  X-Vault-Secret-Key:A3-abcd1234-efgh5678-ijkl9012-mnop3456
```

## Error Handling

### 400 Bad Request
```json
{
  "error": "secret name is required"
}
```

### 401 Unauthorized
```json
{
  "error": "authentication failed"
}
```

### 404 Not Found
```json
{
  "error": "secret not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

## Best Practices

1. **Store credentials securely**: Never commit tokens or secret keys to version control
2. **Use environment variables**: Store sensitive configuration in environment variables
3. **Rotate keys regularly**: Use the key rotation endpoint to rotate vault keys
4. **Set expiration dates**: Add expiration dates to secrets that should be time-limited
5. **Use tags**: Organize secrets with meaningful tags
6. **Monitor audit logs**: Check database audit_logs table for security monitoring

## Testing with a Script

Create `test-api.sh`:

```bash
#!/bin/bash

API_URL="http://localhost:8080"

# Register
echo "Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST $API_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"TestPass123!"}')

EMAIL=$(echo $REGISTER_RESPONSE | jq -r '.email')
SECRET_KEY=$(echo $REGISTER_RESPONSE | jq -r '.secret_key')

echo "Email: $EMAIL"
echo "Secret Key: $SECRET_KEY"

# Login
echo "Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST $API_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"TestPass123!\",\"secret_key\":\"$SECRET_KEY\"}")

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
echo "Token: $TOKEN"

# Create secret
echo "Creating secret..."
curl -X POST $API_URL/api/v1/secrets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Vault-Password: TestPass123!" \
  -H "X-Vault-Secret-Key: $SECRET_KEY" \
  -d '{"name":"test-secret","value":"my-secret-value"}'

echo ""
echo "Test complete!"
```

Run it:
```bash
chmod +x test-api.sh
./test-api.sh
```
