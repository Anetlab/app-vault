package service

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/app-vault/app-vault/internal/crypto"
	"github.com/app-vault/app-vault/internal/db"
	"github.com/google/uuid"
)

// selectUserByEmailRegex matches the SELECT statement used by
// GetUserByEmail while ignoring whitespace differences. Using a regex
// matcher keeps the tests resilient to incidental formatting changes in
// the production code.
var selectUserByEmailRegex = regexp.MustCompile(`(?is)SELECT\s+id,\s+email,\s+secret_key_hash,\s+master_key_salt,` +
	`\s+encrypted_vault_key,\s+vault_key_nonce,\s+srp_verifier,` +
	`\s+current_key_version_id,\s+created_at,\s+updated_at\s+FROM users WHERE email = \$1`)

var insertUserRegex = regexp.MustCompile(`(?is)INSERT INTO users`)

var insertKeyVersionRegex = regexp.MustCompile(`(?is)INSERT INTO key_versions`)

var insertAuditLogRegex = regexp.MustCompile(`(?is)INSERT INTO audit_logs`)

// newAuthServiceWithMock wires an AuthService backed by a sqlmock
// connection.
func newAuthServiceWithMock(t *testing.T) (*AuthService, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	database := db.NewWithDB(sqlDB)
	svc := NewAuthService(database, "test-secret-key-must-be-long-enough")

	cleanup := func() { _ = sqlDB.Close() }
	return svc, mock, cleanup
}

// userColumns mirrors the column order used by GetUserByEmail in the
// production code.
var userColumns = []string{
	"id", "email", "secret_key_hash", "master_key_salt",
	"encrypted_vault_key", "vault_key_nonce", "srp_verifier",
	"current_key_version_id", "created_at", "updated_at",
}

func TestRegister_RejectsShortPassword(t *testing.T) {
	svc, _, cleanup := newAuthServiceWithMock(t)
	defer cleanup()

	_, err := svc.Register(context.Background(), &RegisterRequest{
		Email:    "user@example.com",
		Password: "short",
	})
	if err == nil {
		t.Fatal("expected error for short password, got nil")
	}
}

func TestRegister_HappyPath(t *testing.T) {
	svc, mock, cleanup := newAuthServiceWithMock(t)
	defer cleanup()

	email := "user@example.com"
	password := "verysecurepassword123"

	mock.ExpectQuery(selectUserByEmailRegex.String()).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(insertUserRegex.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(insertKeyVersionRegex.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(insertAuditLogRegex.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := svc.Register(context.Background(), &RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("Register returned nil response")
	}
	if resp.Email != email {
		t.Errorf("expected email %q, got %q", email, resp.Email)
	}
	if len(resp.SecretKey) < 4 || resp.SecretKey[:3] != "A3-" {
		t.Errorf("expected A3-format secret key, got %q", resp.SecretKey)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

func TestRegister_RejectsDuplicateEmail(t *testing.T) {
	svc, mock, cleanup := newAuthServiceWithMock(t)
	defer cleanup()

	email := "dup@example.com"
	password := "verysecurepassword123"
	now := time.Now()
	userID := uuid.New()

	rows := sqlmock.NewRows(userColumns).
		AddRow(userID, email, "hash", []byte("salt"), []byte("enc"), []byte("nonce"),
			[]byte("srp"), nil, now, now)

	mock.ExpectQuery(selectUserByEmailRegex.String()).
		WithArgs(email).
		WillReturnRows(rows)

	_, err := svc.Register(context.Background(), &RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
}

// TestLogin_UniformErrorForAllFailures is the regression test for the
// user-enumeration fix. Whichever credential is wrong, the caller must
// observe the exact same error so a probe cannot tell accounts apart.
func TestLogin_UniformErrorForAllFailures(t *testing.T) {
	const expectedErr = "authentication failed: invalid credentials"

	t.Run("unknown user", func(t *testing.T) {
		svc, mock, cleanup := newAuthServiceWithMock(t)
		defer cleanup()

		mock.ExpectQuery(selectUserByEmailRegex.String()).
			WithArgs("nobody@example.com").
			WillReturnError(sql.ErrNoRows)

		_, err := svc.Login(context.Background(), &LoginRequest{
			Email:     "nobody@example.com",
			Password:  "verysecurepassword123",
			SecretKey: "A3-aaaaaaaaaa-bbbbbbbbbb-cccccccc-dddddddd",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != expectedErr {
			t.Errorf("expected %q, got %q", expectedErr, err.Error())
		}
	})

	t.Run("wrong secret key", func(t *testing.T) {
		svc, mock, cleanup := newAuthServiceWithMock(t)
		defer cleanup()

		email := "user@example.com"
		now := time.Now()
		userID := uuid.New()
		salt := make([]byte, crypto.SaltLen)

		rows := sqlmock.NewRows(userColumns).
			AddRow(userID, email, "hash-that-wont-match", salt,
				[]byte("enc-vault"), make([]byte, crypto.NonceLen),
				[]byte("srp"), nil, now, now)

		mock.ExpectQuery(selectUserByEmailRegex.String()).
			WithArgs(email).
			WillReturnRows(rows)

		_, err := svc.Login(context.Background(), &LoginRequest{
			Email:     email,
			Password:  "verysecurepassword123",
			SecretKey: "A3-aaaaaaaaaa-bbbbbbbbbb-cccccccc-dddddddd",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != expectedErr {
			t.Errorf("expected %q, got %q", expectedErr, err.Error())
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		svc, mock, cleanup := newAuthServiceWithMock(t)
		defer cleanup()

		email := "user@example.com"
		now := time.Now()
		userID := uuid.New()
		salt := make([]byte, crypto.SaltLen)
		realSecretKey := "A3-aaaaaaaaaa-bbbbbbbbbb-cccccccc-dddddddd"
		// Store a hash that matches the real secret key so the code
		// accepts the secret key and reaches the vault-key
		// decryption step (where it must still fail uniformly).
		storedHash := crypto.HashSecretKey(realSecretKey)

		rows := sqlmock.NewRows(userColumns).
			AddRow(userID, email, storedHash, salt,
				[]byte("garbage-enc-vault-bytes"), make([]byte, crypto.NonceLen),
				[]byte("srp"), nil, now, now)

		mock.ExpectQuery(selectUserByEmailRegex.String()).
			WithArgs(email).
			WillReturnRows(rows)

		_, err := svc.Login(context.Background(), &LoginRequest{
			Email:     email,
			Password:  "verysecurepassword123",
			SecretKey: realSecretKey,
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != expectedErr {
			t.Errorf("expected %q, got %q", expectedErr, err.Error())
		}
	})
}

func TestLogin_HappyPath(t *testing.T) {
	svc, mock, cleanup := newAuthServiceWithMock(t)
	defer cleanup()

	email := "user@example.com"
	password := "verysecurepassword123"
	secretKey := "A3-aaaaaaaaaa-bbbbbbbbbb-cccccccc-dddddddd"
	now := time.Now()
	userID := uuid.New()
	salt := make([]byte, crypto.SaltLen)

	// Pre-compute a vault key wrapped with a KEK derived from
	// (password, secretKey, salt), so the production code can
	// successfully decrypt it.
	masterKey := crypto.DeriveMasterKey(password, secretKey, salt)
	kek, err := crypto.DeriveKeyEncryptionKey(masterKey, "vault-key-wrap")
	if err != nil {
		t.Fatalf("DeriveKeyEncryptionKey: %v", err)
	}
	plainVaultKey := make([]byte, 32) // chacha20poly1305.KeySize
	for i := range plainVaultKey {
		plainVaultKey[i] = byte(i + 1)
	}
	nonce := make([]byte, crypto.NonceLen)
	encryptedVaultKey, err := crypto.EncryptVaultKey(plainVaultKey, kek, nonce)
	if err != nil {
		t.Fatalf("EncryptVaultKey: %v", err)
	}
	crypto.ZeroBytes(masterKey)
	crypto.ZeroBytes(kek)
	crypto.ZeroBytes(plainVaultKey)

	storedHash := crypto.HashSecretKey(secretKey)

	rows := sqlmock.NewRows(userColumns).
		AddRow(userID, email, storedHash, salt, encryptedVaultKey, nonce,
			[]byte("srp"), nil, now, now)

	mock.ExpectQuery(selectUserByEmailRegex.String()).
		WithArgs(email).
		WillReturnRows(rows)
	mock.ExpectExec(insertAuditLogRegex.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := svc.Login(context.Background(), &LoginRequest{
		Email:     email,
		Password:  password,
		SecretKey: secretKey,
	})
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("Login returned nil response")
	}
	if resp.UserID != userID {
		t.Errorf("expected userID %v, got %v", userID, resp.UserID)
	}
	if resp.Token == "" {
		t.Error("expected non-empty JWT token")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
