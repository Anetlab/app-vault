-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    secret_key_hash VARCHAR(255) NOT NULL,
    master_key_salt BYTEA NOT NULL,
    encrypted_vault_key BYTEA NOT NULL,
    vault_key_nonce BYTEA NOT NULL,
    srp_verifier BYTEA NOT NULL,
    current_key_version_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create service principals table
CREATE TABLE IF NOT EXISTS service_principals (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id VARCHAR(64) UNIQUE NOT NULL,
    client_secret_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    permissions TEXT[] NOT NULL DEFAULT '{}',
    allowed_secrets TEXT[],
    allowed_tags TEXT[],
    ip_whitelist TEXT[],
    rate_limit INTEGER DEFAULT 60,
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create key versions table
CREATE TABLE IF NOT EXISTS key_versions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    encrypted_vault_key BYTEA NOT NULL,
    vault_key_nonce BYTEA NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    rotated_at TIMESTAMPTZ,
    destroy_at TIMESTAMPTZ,
    destroyed_at TIMESTAMPTZ,
    rotation_reason TEXT,
    UNIQUE(user_id, version)
);

-- Create secrets table
CREATE TABLE IF NOT EXISTS secrets (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_version_id UUID NOT NULL REFERENCES key_versions(id),
    name VARCHAR(255) NOT NULL,
    encrypted_data BYTEA NOT NULL,
    nonce BYTEA NOT NULL,
    secret_type VARCHAR(50) NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    previous_id UUID REFERENCES secrets(id),
    tags TEXT[],
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    access_count INTEGER DEFAULT 0,
    last_accessed_at TIMESTAMPTZ,
    UNIQUE(user_id, name)
);

-- Create audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_principal_id UUID REFERENCES service_principals(id),
    action VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create rate limits table
CREATE TABLE IF NOT EXISTS rate_limits (
    key VARCHAR(255) PRIMARY KEY,
    count INTEGER DEFAULT 0,
    window_start TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_secrets_user_id ON secrets(user_id);
CREATE INDEX IF NOT EXISTS idx_secrets_name ON secrets(user_id, name);
CREATE INDEX IF NOT EXISTS idx_secrets_tags ON secrets USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_secrets_key_version ON secrets(key_version_id);
CREATE INDEX IF NOT EXISTS idx_service_principals_client_id ON service_principals(client_id);
CREATE INDEX IF NOT EXISTS idx_service_principals_user_id ON service_principals(user_id);
CREATE INDEX IF NOT EXISTS idx_key_versions_user_id ON key_versions(user_id);
CREATE INDEX IF NOT EXISTS idx_key_versions_status ON key_versions(user_id, status);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_sp_id ON audit_logs(service_principal_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
