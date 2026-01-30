package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	if len(salt1) != SaltLen {
		t.Errorf("Expected salt length %d, got %d", SaltLen, len(salt1))
	}

	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	if bytes.Equal(salt1, salt2) {
		t.Error("Two generated salts should not be equal")
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce1, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce failed: %v", err)
	}

	if len(nonce1) != NonceLen {
		t.Errorf("Expected nonce length %d, got %d", NonceLen, len(nonce1))
	}

	nonce2, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce failed: %v", err)
	}

	if bytes.Equal(nonce1, nonce2) {
		t.Error("Two generated nonces should not be equal")
	}
}

func TestEncryptDecryptData(t *testing.T) {
	key, err := GenerateVaultKey()
	if err != nil {
		t.Fatalf("GenerateVaultKey failed: %v", err)
	}

	nonce, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce failed: %v", err)
	}

	plaintext := []byte("This is a secret message")

	ciphertext, err := EncryptData(plaintext, key, nonce)
	if err != nil {
		t.Fatalf("EncryptData failed: %v", err)
	}

	if bytes.Equal(plaintext, ciphertext) {
		t.Error("Ciphertext should not equal plaintext")
	}

	decrypted, err := DecryptData(ciphertext, key, nonce)
	if err != nil {
		t.Fatalf("DecryptData failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("Decrypted data does not match original: got %s, want %s", decrypted, plaintext)
	}
}

func TestEncryptDecryptWithWrongKey(t *testing.T) {
	key1, _ := GenerateVaultKey()
	key2, _ := GenerateVaultKey()
	nonce, _ := GenerateNonce()

	plaintext := []byte("Secret data")

	ciphertext, err := EncryptData(plaintext, key1, nonce)
	if err != nil {
		t.Fatalf("EncryptData failed: %v", err)
	}

	_, err = DecryptData(ciphertext, key2, nonce)
	if err == nil {
		t.Error("Expected decryption to fail with wrong key")
	}
}

func TestDeriveMasterKey(t *testing.T) {
	password := "mypassword"
	secretKey := "A3-abcd1234-efgh5678-ijkl9012-mnop3456"
	salt, _ := GenerateSalt()

	key1 := DeriveMasterKey(password, secretKey, salt)
	key2 := DeriveMasterKey(password, secretKey, salt)

	if !bytes.Equal(key1, key2) {
		t.Error("Same inputs should produce same master key")
	}

	if len(key1) != Argon2KeyLen {
		t.Errorf("Expected key length %d, got %d", Argon2KeyLen, len(key1))
	}

	key3 := DeriveMasterKey("different", secretKey, salt)
	if bytes.Equal(key1, key3) {
		t.Error("Different passwords should produce different keys")
	}
}

func TestHashSecretKey(t *testing.T) {
	secretKey := "A3-abcd1234-efgh5678-ijkl9012-mnop3456"

	hash1 := HashSecretKey(secretKey)
	hash2 := HashSecretKey(secretKey)

	if hash1 != hash2 {
		t.Error("Same secret key should produce same hash")
	}

	hash3 := HashSecretKey("different-key")
	if hash1 == hash3 {
		t.Error("Different keys should produce different hashes")
	}
}

func TestCompareHashConstantTime(t *testing.T) {
	secretKey := "my-secret-key"
	hash := HashSecretKey(secretKey)

	if !CompareHashConstantTime(hash, secretKey) {
		t.Error("Hash comparison should return true for correct key")
	}

	if CompareHashConstantTime(hash, "wrong-key") {
		t.Error("Hash comparison should return false for incorrect key")
	}
}

func TestGenerateSecretKey(t *testing.T) {
	key1, err := GenerateSecretKey()
	if err != nil {
		t.Fatalf("GenerateSecretKey failed: %v", err)
	}

	if len(key1) < 20 {
		t.Error("Secret key should be reasonably long")
	}

	if key1[:3] != "A3-" {
		t.Error("Secret key should start with A3-")
	}

	key2, err := GenerateSecretKey()
	if err != nil {
		t.Fatalf("GenerateSecretKey failed: %v", err)
	}

	if key1 == key2 {
		t.Error("Two generated secret keys should not be equal")
	}
}
