# Copilot Instructions — SecureVault (Go + PostgreSQL)

> **Purpose**: This document tells GitHub Copilot (and contributors) how to write code for this repository. Follow these rules when generating or modifying code.

## Project Summary

SecureVault is a Key Vault–like secret management service written in **Go** with **PostgreSQL** storage. It aims for strong security properties inspired by Azure Key Vault and 1Password:

* Secrets are encrypted at rest using **XChaCha20-Poly1305**.
* A **Master Key** is derived using **Argon2id** from `(password + secretKey)`.
* A per-user **Vault Key** encrypts secrets. Vault Keys are stored encrypted and support **key rotation**.
* Supports **user sessions** and **service principals** (client credentials).
* Provides **RBAC** and optional restrictions (`allowed_secrets`, `allowed_tags`, `ip_whitelist`, rate limits).
* Maintains **audit logs**.
  
  ## Core Rules (must-follow)
  
  ### Security
1. **Never log secrets** or decrypted plaintext.
* No plaintext in logs, errors, metrics, traces.
* Avoid dumping request bodies.
1. **No secrets in config files**.
* Config may include only **references** (secret names/IDs).
* Credentials are provided via env vars/secret stores.
1. **Use constant-time comparisons** for sensitive equality checks.
* Client secrets, tokens, derived keys.
1. **Prefer authenticated encryption (AEAD)**.
* Use XChaCha20-Poly1305 (or ChaCha20-Poly1305 if required).
* Always verify authentication tags before using plaintext.
1. **Key material handling**.
* Minimize time in memory.
* Zero buffers where feasible (best-effort in Go).
* Do not store VaultKey in DB plaintext.
1. **Do not introduce insecure crypto**.
* No custom crypto.
* Do not use AES-CBC, raw RSA, MD5, SHA1, or homegrown KDF.
1. **Validate all inputs**.
* Length limits, allowed character sets, JSON schema-like validation.
* Guard against path traversal and injection.
  
  ### Database
1. All DB access must be parameterized.
2. Use explicit transactions for multi-step changes (e.g., key rotation re-encryption).
3. Keep migrations deterministic and idempotent.
4. Index fields used for lookups: user_id, client_id, secret name, key_version, created_at.
   
   ### API
5. Follow a consistent REST style:
* `POST /api/v1/secrets` create
* `GET /api/v1/secrets` list
* `GET /api/v1/secrets/by-name?name=...` fetch by name
* `GET /api/v1/secrets/{id}` fetch by id
* `DELETE /api/v1/secrets/{id}` delete
1. Always return JSON with proper HTTP status codes.
2. Errors should be structured: `{ "error": "..." }`.
3. Never return secret plaintext unless endpoint explicitly returns secret data and caller has `read`.
   
   ## Coding Standards (Go)
   
   ### Style & Structure
* Use `go fmt` formatting.
* Keep functions small and testable.
* Prefer interfaces at boundaries (DB, crypto, session store).
* Keep crypto operations in a dedicated package/module (e.g., `crypto/`).
  
  ### Error Handling
* Wrap errors with context using `%w`.
* Avoid leaking internal error details to clients.
* Differentiate between:
  * 400: validation errors
  * 401: auth failures
  * 403: permission denied
  * 404: not found
  * 409: conflict
  * 429: rate limit
  * 500: internal error
    
    ### Concurrency
* Ensure session store access is safe (mutex or concurrent map).
* Avoid data races, especially around rotation and re-encryption.
  
  ## Repository Conventions
  
  ### Suggested Layout
  
  > If the repository differs, follow existing structure.
* `cmd/server/` — entrypoint(s)
* `internal/api/` — HTTP handlers, request/response DTOs
* `internal/service/` — business logic (vault, key rotation, auth)
* `internal/db/` — SQL and repository layer
* `internal/crypto/` — KDF, AEAD, hashing
* `internal/models/` — models used across layers
* `migrations/` — SQL migration files
* `client/` — Go client SDK (optional)
  
  ### Naming
* Use clear names: `RotateVaultKey`, `AuthenticateServicePrincipal`, `CreateSecret`.
* DTOs: `CreateSecretRequest`, `SecretResponse`.
* Keep JSON tags snake_case if that’s the API style.
  
  ## Key System Expectations
  
  ### Key Derivation
* Master Key: Argon2id(password + secretKey, salt, params)
* Subkeys: HKDF for contextual derivation, e.g. `HKDF(masterKey, info="vault-key-wrap")`
  
  ### Vault Key Storage
* Store encrypted vault key + nonce + metadata.
* Store multiple key versions with statuses: `active`, `deprecated`, `destroyed`.
  
  ### Rotation Behavior
* Vault Key rotation must:
  * Create a new key version (active)
  * Re-encrypt all secrets from old key version to new
  * Mark old as deprecated
  * Schedule destruction after grace period
  * Audit log the event with count of affected secrets
* Master Key rotation (password change) must:
  * Re-wrap current vault key with new master key
  * Invalidate sessions
  * Audit log the event
    
    ## Service Principals Expectations
* Client secret must be **stored hashed**, never plaintext.
* Authentication uses client_id + client_secret.
* Enforce:
  * active flag
  * expiration
  * ip whitelist (if configured)
  * rate limits (if configured)
    
    ## Testing Requirements
    
    Whenever Copilot adds or changes behavior:
1. Add unit tests for:
* crypto (encrypt/decrypt roundtrip, wrong key failure)
* permission checks
* service principal auth checks
1. Add integration tests for:
* key rotation (re-encryption correctness)
* token auth flow
1. Ensure deterministic tests (no random timeouts).
   
   ## Documentation Requirements
   
   When adding endpoints or changing request/response:
* Update `README.md` and any API docs.
* Add example `curl` usage.
  
  ## Dependencies Policy
* Prefer standard library.
* Approved common deps (example):
  * PostgreSQL driver (`lib/pq` or `pgx`)
  * JWT library
  * `x/crypto` for Argon2id / HKDF
  * `google/uuid`
    Do not add heavy frameworks unless requested.
    
    ## Safe Defaults
* Default deny: if permissions are missing, reject.
* Min password length: 12+.
* Strong Argon2id params (tunable via env).
* JWT TTL short (e.g., 15–60 min) with refresh strategy if needed.
  
  ## Copilot Prompting Guidance (how to generate changes)
  
  When asked to implement a feature, Copilot should:
1. Identify impacted layers: API → service → db → crypto.
2. Add/modify DB migration if schema changes.
3. Add tests.
4. Update README/API tables.
5. Ensure no secret leakage in logs/errors.
   
   ## Non-Goals (avoid unless explicitly requested)
* Building a UI dashboard
* Multi-region replication
* Full OAuth2 authorization server
* HSM/KMS integration (future roadmap)
  
  ## PR Checklist (for generated code)
* \[ \] `go test ./...` passes
* \[ \] `go vet ./...` passes
* \[ \] No secrets in logs
* \[ \] DB queries are parameterized
* \[ \] Rotation is transactional and audited
* \[ \] README/API docs updated
