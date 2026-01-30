package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

// Argon2id parameters
const (
	Argon2Time    = 3
	Argon2Memory  = 64 * 1024
	Argon2Threads = 4
	Argon2KeyLen  = 32
	SaltLen       = 32
	NonceLen      = 24
)

// GenerateSalt creates a cryptographically secure random salt
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// GenerateNonce creates a cryptographically secure random nonce for XChaCha20-Poly1305
func GenerateNonce() ([]byte, error) {
	nonce := make([]byte, NonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return nonce, nil
}

// GenerateVaultKey creates a new random vault key
func GenerateVaultKey() ([]byte, error) {
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate vault key: %w", err)
	}
	return key, nil
}

// DeriveMasterKey derives the master key from password and secret key using Argon2id
func DeriveMasterKey(password, secretKey string, salt []byte) []byte {
	combined := password + secretKey
	return argon2.IDKey(
		[]byte(combined),
		salt,
		Argon2Time,
		Argon2Memory,
		Argon2Threads,
		Argon2KeyLen,
	)
}

// DeriveKeyEncryptionKey derives a key encryption key from the master key using HKDF
func DeriveKeyEncryptionKey(masterKey []byte, info string) ([]byte, error) {
	reader := hkdf.New(sha256.New, masterKey, nil, []byte(info))
	kek := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(reader, kek); err != nil {
		return nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	return kek, nil
}

// EncryptVaultKey encrypts the vault key with the key encryption key
func EncryptVaultKey(vaultKey, kek, nonce []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(kek)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	ciphertext := aead.Seal(nil, nonce, vaultKey, nil)
	return ciphertext, nil
}

// DecryptVaultKey decrypts the encrypted vault key with the key encryption key
func DecryptVaultKey(encryptedVaultKey, kek, nonce []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(kek)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, encryptedVaultKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt vault key: %w", err)
	}

	return plaintext, nil
}

// EncryptData encrypts data with XChaCha20-Poly1305
func EncryptData(plaintext, key, nonce []byte) ([]byte, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", chacha20poly1305.KeySize, len(key))
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptData decrypts data with XChaCha20-Poly1305
func DecryptData(ciphertext, key, nonce []byte) ([]byte, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", chacha20poly1305.KeySize, len(key))
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// HashSecretKey hashes a secret key for storage (using SHA-256)
func HashSecretKey(secretKey string) string {
	hash := sha256.Sum256([]byte(secretKey))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// HashClientSecret hashes a client secret for service principal authentication
func HashClientSecret(clientSecret string) string {
	hash := sha256.Sum256([]byte(clientSecret))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// CompareHashConstantTime compares a hash with a value using constant-time comparison
func CompareHashConstantTime(hash, value string) bool {
	computedHash := sha256.Sum256([]byte(value))
	computedHashStr := base64.StdEncoding.EncodeToString(computedHash[:])
	return subtle.ConstantTimeCompare([]byte(hash), []byte(computedHashStr)) == 1
}

// GenerateSecretKey generates a random A3-formatted secret key
func GenerateSecretKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", fmt.Errorf("failed to generate secret key: %w", err)
	}

	encoded := base64.RawURLEncoding.EncodeToString(bytes)

	formatted := fmt.Sprintf("A3-%s-%s-%s-%s",
		encoded[0:8],
		encoded[8:16],
		encoded[16:24],
		encoded[24:32])

	return formatted, nil
}

// GenerateClientID generates a unique client ID
func GenerateClientID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", fmt.Errorf("failed to generate client ID: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// GenerateClientSecret generates a secure client secret
func GenerateClientSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", fmt.Errorf("failed to generate client secret: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// ZeroBytes attempts to zero a byte slice in memory
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
