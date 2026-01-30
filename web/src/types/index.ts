export interface User {
  id: string
  email: string
  createdAt: string
  lastLogin?: string
}

export interface LoginRequest {
  email: string
  password: string
  secretKey: string
}

export interface RegisterRequest {
  email: string
  password: string
}

export interface AuthResponse {
  token: string
  user: User
  secretKey?: string
}

export interface Secret {
  id: string
  userId: string
  name: string
  type: string
  tags: string[]
  version: number
  createdAt: string
  updatedAt: string
  expiresAt?: string
  accessCount: number
  lastAccessed?: string
}

export interface SecretWithValue extends Secret {
  value: string
}

export interface CreateSecretRequest {
  name: string
  value: string
  type: string
  tags?: string[]
  expiresAt?: string
}

export interface UpdateSecretRequest {
  value: string
  version: number
}

export interface ServicePrincipal {
  id: string
  userId: string
  name: string
  clientId: string
  permissions: string[]
  allowedSecrets: string[]
  allowedTags: string[]
  ipWhitelist: string[]
  rateLimit: number
  isActive: boolean
  expiresAt?: string
  lastUsedAt?: string
  createdAt: string
}

export interface CreateServicePrincipalRequest {
  name: string
  permissions: string[]
  allowedSecrets?: string[]
  allowedTags?: string[]
  ipWhitelist?: string[]
  rateLimit?: number
  expiresAt?: string
}

export interface ServicePrincipalResponse {
  servicePrincipal: ServicePrincipal
  clientSecret: string
}

export interface KeyVersion {
  id: string
  userId: string
  versionNumber: number
  status: 'active' | 'deprecated' | 'destroyed'
  createdAt: string
  deprecatedAt?: string
  destroyedAt?: string
}

export interface KeyRotationRequest {
  reason: string
  gracePeriodDays?: number
}

export interface KeyRotationResponse {
  newVersionId: string
  secretsReEncrypted: number
  durationMs: number
}

export interface KeyStatus {
  currentVersion: KeyVersion
  allVersions: KeyVersion[]
  nextScheduledRotation?: string
}

export interface AuditLog {
  id: string
  userId?: string
  servicePrincipalId?: string
  action: string
  resourceType: string
  resourceId?: string
  ipAddress?: string
  userAgent?: string
  status: string
  details?: Record<string, unknown>
  createdAt: string
}

export interface ApiError {
  error: string
  details?: Record<string, unknown>
}

export interface HealthCheck {
  status: string
  database: string
  timestamp: string
}
