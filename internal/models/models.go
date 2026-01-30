package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user account in the system
type User struct {
	ID                  uuid.UUID
	Email               string
	SecretKeyHash       string
	MasterKeySalt       []byte
	EncryptedVaultKey   []byte
	VaultKeyNonce       []byte
	SRPVerifier         []byte
	CurrentKeyVersionID *uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// ServicePrincipal represents an application credential for API access
type ServicePrincipal struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	ClientID         string
	ClientSecretHash string
	Name             string
	Description      string
	Permissions      []string
	AllowedSecrets   []string
	AllowedTags      []string
	IPWhitelist      []string
	RateLimit        int
	IsActive         bool
	ExpiresAt        *time.Time
	LastUsedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// KeyVersion represents a version of the vault encryption key
type KeyVersion struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	Version           int
	EncryptedVaultKey []byte
	VaultKeyNonce     []byte
	Status            string
	CreatedAt         time.Time
	RotatedAt         *time.Time
	DestroyAt         *time.Time
	DestroyedAt       *time.Time
	RotationReason    string
}

// Secret represents an encrypted secret
type Secret struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	KeyVersionID   uuid.UUID
	Name           string
	EncryptedData  []byte
	Nonce          []byte
	PublicKey      []byte // Public key for sharing with other projects
	SecretType     string
	Version        int
	PreviousID     *uuid.UUID
	Tags           []string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ExpiresAt      *time.Time
	AccessCount    int
	LastAccessedAt *time.Time
}

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	ServicePrincipalID *uuid.UUID
	Action             string
	ResourceID         string
	Details            map[string]interface{}
	IPAddress          string
	UserAgent          string
	CreatedAt          time.Time
}

// RateLimit represents rate limiting state
type RateLimit struct {
	Key         string
	Count       int
	WindowStart time.Time
}

// Key version statuses
const (
	KeyStatusActive     = "active"
	KeyStatusDeprecated = "deprecated"
	KeyStatusDestroyed  = "destroyed"
)

// Secret types
const (
	SecretTypeGeneric     = "generic"
	SecretTypePassword    = "password"
	SecretTypeAPIKey      = "api_key"
	SecretTypeCertificate = "certificate"
	SecretTypeConnection  = "connection"
)

// Permissions
const (
	PermissionReadSecrets   = "secrets:read"
	PermissionWriteSecrets  = "secrets:write"
	PermissionDeleteSecrets = "secrets:delete"
	PermissionListSecrets   = "secrets:list"
	PermissionRotateKeys    = "keys:rotate"
	PermissionAdmin         = "admin"
)
