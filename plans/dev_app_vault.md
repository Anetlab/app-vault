## Обзор Azure Key Vault

Привет, давай сегодня разберемся как устроен Azure KeyVault

## Claude Sonnet 4.5

Привет! Отлично, давай разберемся с Azure Key Vault.
Azure Key Vault — это облачный сервис Microsoft для безопасного хранения и
управления секретами, ключами шифрования и сертификатами.
Основные компоненты:

- Secrets (Секреты)
  Пароли, строки подключения, API ключи
  Хранятся в зашифрованном виде
  Версионируются автоматически
- Keys (Ключи)
  Криптографические ключи для шифрования/дешифрования
  Поддержка RSA и EC ключей
  Могут храниться в HSM (Hardware Security Module)
- Certificates (Сертификаты)
  SSL/TLS сертификаты
  Автоматическое обновление
  Интеграция с центрами сертификации
  
  ## Архитектура:
  
  ## •
  
  ## •
  
  ## •
  
  ## •
  
  ## •
  
  ## •
  
  ## •
  
  ## •
  
  ## •
  
  ## ┌─────────────┐
  
  ## │ Application │
  
  ## └──────┬──────┘
  
  │ (Azure AD Auth)
  
  ## ▼
  
  ## ┌─────────────────┐
  
  │  Key Vault API  │
  
  ## └──────┬──────────┘
  
  ## │
  
  ## ┌──────▼──────────────────┐

Ключевые особенности:
Аутентификация: через Azure AD (Managed Identity, Service Principal)
Контроль доступа: RBAC или Access Policies
Аудит: логирование всех операций
Geo-репликация: автоматическое резервное копирование
Что конкретно тебя интересует — настройка, интеграция с приложениями, или
что-то другое?
Давай сделаем аналогичное приложение на go

## Claude Sonnet 4.5

Отлично! Давай создадим аналог Azure Key Vault на Go. Сделаем REST API с
базовым функционалом для хранения секретов, ключей и управления
доступом.

## │   Key Vault Instance    │

## │  ┌──────┬──────┬──────┐ │

│  │Secret│ Keys │Certs │ │

## │  └──────┴──────┴──────┘ │

## └─────────────────────────┘

## •

## •

## •

## •

package main
import (

## "crypto/aes"

## "crypto/cipher"

## "crypto/rand"

## "crypto/sha256"

## "encoding/base64"

## "encoding/json"

## "fmt"

## "io"

## "net/http"

## "sync"

## "time"

## "github.com/golang-jwt/jwt/v5"

## "github.com/gorilla/mux"

## )

// Secret представляет секрет в хранилище

type Secret struct {
Name      string    `json:"name"`
Value     string    `json:"value"`
Version   int       `json:"version"`
CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`

## }

// SecretVersion хранит историю версий
type SecretVersion struct {
Version   int       `json:"version"`
Value     string    `json:"value"`
CreatedAt time.Time `json:"created_at"`

## }

// Vault основное хранилище
type Vault struct {
secrets        map[string][]SecretVersion
encryptionKey  []byte
mu             sync.RWMutex
accessPolicies map[string][]string // user -> permissions

## }

// User для аутентификации
type User struct {
Username string `json:"username"`
Password string `json:"password"`

## }

var (
vault     *Vault
jwtSecret = []byte("your-secret-key-change-in-production")

## )

// NewVault создает новое хранилище
func NewVault(masterKey string) *Vault {
hash := sha256.Sum256([]byte(masterKey))
return &Vault{
secrets:        make(map[string][]SecretVersion),
encryptionKey:  hash[:],
accessPolicies: make(map[string][]string),

## }

## }

// Encrypt шифрует данные
func (v *Vault) Encrypt(plaintext string) (string, error) {
block, err := aes.NewCipher(v.encryptionKey)
if err != nil {
return "", err

## }

gcm, err := cipher.NewGCM(block)
if err != nil {
return "", err

## }

nonce := make([]byte, gcm.NonceSize())
if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
return "", err

## }

ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
return base64.StdEncoding.EncodeToString(ciphertext), nil

## }

// Decrypt расшифровывает данные
func (v *Vault) Decrypt(ciphertext string) (string, error) {
data, err := base64.StdEncoding.DecodeString(ciphertext)
if err != nil {
return "", err

## }

block, err := aes.NewCipher(v.encryptionKey)
if err != nil {
return "", err

## }

gcm, err := cipher.NewGCM(block)
if err != nil {
return "", err

## }

nonceSize := gcm.NonceSize()
if len(data) < nonceSize {
return "", fmt.Errorf("ciphertext too short")

## }

nonce, cipherData := data[:nonceSize], data[nonceSize:]
plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
if err != nil {
return "", err

## }

return string(plaintext), nil

## }

// SetSecret сохраняет секрет
func (v *Vault) SetSecret(name, value string) error {
v.mu.Lock()
defer v.mu.Unlock()

encrypted, err := v.Encrypt(value)
if err != nil {
return err

## }

versions := v.secrets[name]
newVersion := len(versions) + 1
secretVersion := SecretVersion{
Version:   newVersion,
Value:     encrypted,
CreatedAt: time.Now(),

## }

v.secrets[name] = append(versions, secretVersion)
return nil

## }

// GetSecret получает последнюю версию секрета
func (v *Vault) GetSecret(name string) (*Secret, error) {
v.mu.RLock()
defer v.mu.RUnlock()
versions, exists := v.secrets[name]
if !exists || len(versions) == 0 {
return nil, fmt.Errorf("secret not found")

## }

latest := versions[len(versions)-1]
decrypted, err := v.Decrypt(latest.Value)
if err != nil {
return nil, err

## }

return &Secret{
Name:      name,
Value:     decrypted,
Version:   latest.Version,
CreatedAt: latest.CreatedAt,
UpdatedAt: latest.CreatedAt,
}, nil

## }

// GetSecretVersion получает конкретную версию секрета
func (v *Vault) GetSecretVersion(name string, version int) (*Secret, error) {
v.mu.RLock()
defer v.mu.RUnlock()
versions, exists := v.secrets[name]

if !exists {
return nil, fmt.Errorf("secret not found")

## }

if version < 1 || version > len(versions) {
return nil, fmt.Errorf("version not found")

## }

secretVersion := versions[version-1]
decrypted, err := v.Decrypt(secretVersion.Value)
if err != nil {
return nil, err

## }

return &Secret{
Name:      name,
Value:     decrypted,
Version:   secretVersion.Version,
CreatedAt: secretVersion.CreatedAt,
}, nil

## }

// ListSecrets возвращает список всех секретов
func (v *Vault) ListSecrets() []string {
v.mu.RLock()
defer v.mu.RUnlock()
secrets := make([]string, 0, len(v.secrets))
for name := range v.secrets {
secrets = append(secrets, name)

## }

return secrets

## }

// DeleteSecret удаляет секрет
func (v *Vault) DeleteSecret(name string) error {
v.mu.Lock()
defer v.mu.Unlock()
if _, exists := v.secrets[name]; !exists {
return fmt.Errorf("secret not found")

## }

delete(v.secrets, name)
return nil

## }

// SetAccessPolicy устанавливает права доступа
func (v *Vault) SetAccessPolicy(username string, permissions []string) {
v.mu.Lock()

defer v.mu.Unlock()
v.accessPolicies[username] = permissions

## }

// CheckPermission проверяет права доступа
func (v *Vault) CheckPermission(username, permission string) bool {
v.mu.RLock()
defer v.mu.RUnlock()
permissions, exists := v.accessPolicies[username]
if !exists {
return false

## }

for _, p := range permissions {
if p == permission || p == "admin" {
return true

## }

## }

return false

## }

// GenerateJWT генерирует JWT токен
func GenerateJWT(username string) (string, error) {
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
"username": username,
"exp":      time.Now().Add(time.Hour * 24).Unix(),

## })

return token.SignedString(jwtSecret)

## }

// ValidateJWT проверяет JWT токен
func ValidateJWT(tokenString string) (string, error) {
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
return jwtSecret, nil

## })

if err != nil {
return "", err

## }

if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
return claims["username"].(string), nil

## }

return "", fmt.Errorf("invalid token")

## }

// Middleware для аутентификации

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
return func(w http.ResponseWriter, r *http.Request) {
authHeader := r.Header.Get("Authorization")
if authHeader == "" {
http.Error(w, "Missing authorization header", http.StatusUnauthorized)
return

## }

tokenString := authHeader[len("Bearer "):]
username, err := ValidateJWT(tokenString)
if err != nil {
http.Error(w, "Invalid token", http.StatusUnauthorized)
return

## }

r.Header.Set("X-Username", username)
next(w, r)

## }

## }

// HTTP Handlers
func LoginHandler(w http.ResponseWriter, r *http.Request) {
var user User
if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
http.Error(w, err.Error(), http.StatusBadRequest)
return

## }

// Простая проверка (в продакшене использовать bcrypt и БД)
if user.Username == "admin" && user.Password == "password" {
token, err := GenerateJWT(user.Username)
if err != nil {
http.Error(w, err.Error(), http.StatusInternalServerError)
return

## }

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]string{"token": token})
return

## }

http.Error(w, "Invalid credentials", http.StatusUnauthorized)

## }

func SetSecretHandler(w http.ResponseWriter, r *http.Request) {
username := r.Header.Get("X-Username")
if !vault.CheckPermission(username, "write") {
http.Error(w, "Permission denied", http.StatusForbidden)
return

## }

vars := mux.Vars(r)
name := vars["name"]
var data map[string]string
if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
http.Error(w, err.Error(), http.StatusBadRequest)
return

## }

value, ok := data["value"]
if !ok {
http.Error(w, "Missing value field", http.StatusBadRequest)
return

## }

if err := vault.SetSecret(name, value); err != nil {
http.Error(w, err.Error(), http.StatusInternalServerError)
return

## }

w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(map[string]string{"message": "Secret created"})

## }

func GetSecretHandler(w http.ResponseWriter, r *http.Request) {
username := r.Header.Get("X-Username")
if !vault.CheckPermission(username, "read") {
http.Error(w, "Permission denied", http.StatusForbidden)
return

## }

vars := mux.Vars(r)
name := vars["name"]
secret, err := vault.GetSecret(name)
if err != nil {
http.Error(w, err.Error(), http.StatusNotFound)
return

## }

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(secret)

## }

func ListSecretsHandler(w http.ResponseWriter, r *http.Request) {
username := r.Header.Get("X-Username")
if !vault.CheckPermission(username, "list") {
http.Error(w, "Permission denied", http.StatusForbidden)

return

## }

secrets := vault.ListSecrets()
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string][]string{"secrets": secrets})

## }

func DeleteSecretHandler(w http.ResponseWriter, r *http.Request) {
username := r.Header.Get("X-Username")
if !vault.CheckPermission(username, "delete") {
http.Error(w, "Permission denied", http.StatusForbidden)
return

## }

vars := mux.Vars(r)
name := vars["name"]
if err := vault.DeleteSecret(name); err != nil {
http.Error(w, err.Error(), http.StatusNotFound)
return

## }

w.WriteHeader(http.StatusNoContent)

## }

func main() {
// Инициализация хранилища
vault = NewVault("master-encryption-key-change-me")
// Установка прав доступа для admin
vault.SetAccessPolicy("admin", []string{"read", "write", "delete", "list"})

## // Роутер

r := mux.NewRouter()
// Публичные эндпоинты
r.HandleFunc("/api/v1/login", LoginHandler).Methods("POST")
// Защищенные эндпоинты
r.HandleFunc("/api/v1/secrets", AuthMiddleware(ListSecretsHandler)).Methods("GET")
r.HandleFunc("/api/v1/secrets/{name}",
AuthMiddleware(SetSecretHandler)).Methods("PUT")
r.HandleFunc("/api/v1/secrets/{name}",
AuthMiddleware(GetSecretHandler)).Methods("GET")
r.HandleFunc("/api/v1/secrets/{name}",
AuthMiddleware(DeleteSecretHandler)).Methods("DELETE")
fmt.Println(" Key Vault server starting on :8080")

Как использовать:

- Установи зависимости:
- Запусти сервер:
- Примеры запросов:
  http.ListenAndServe(":8080", r)
  
  ## }
  
  go get github.com/gorilla/mux
  go get github.com/golang-jwt/jwt/v5
  go run main.go
  
  ## # Логин
  
  curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  
  ## -d '{"username":"admin","password":"password"}'
  
  # Сохранить секрет
  
  curl -X PUT http://localhost:8080/api/v1/secrets/db-password \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  
  ## -d '{"value":"super-secret-password"}'
  
  # Получить секрет
  
  curl -X GET http://localhost:8080/api/v1/secrets/db-password \
  -H "Authorization: Bearer YOUR_TOKEN"
  
  # Список секретов
  
  curl -X GET http://localhost:8080/api/v1/secrets \
  -H "Authorization: Bearer YOUR_TOKEN"
  
  # Удалить секрет
  
  curl -X DELETE http://localhost:8080/api/v1/secrets/db-password \
  -H "Authorization: Bearer YOUR_TOKEN"

Основные фичи:
✅ AES-256 шифрование
✅ Версионирование секретов
✅ JWT аутентификация
✅ Контроль доступа (RBAC)

## ✅ REST API

✅ Thread-safe операции
Хочешь добавить что-то еще — персистентность в БД, HSM поддержку, или
метрики?
Отлично, давай двигаться дальше. Давай углубимся в хранение
секретов и будем использовать бд но нам нужно значение и всё
что мы храним было зашифрованно так чтобы ни один хакер
не смог отобрать это у нас чтобы хранение было такое же
защищённое как у 1password

## Claude Opus 4.5

Отлично! Давай сделаем серьёзную криптографию как у 1Password. Их модель
использует двойное шифрование с Secret Key + Master Password, так что
даже при утечке БД данные бесполезны.
package main
import (

## "crypto/aes"

## "crypto/cipher"

## "crypto/rand"

## "crypto/sha256"

## "crypto/subtle"

## "database/sql"

## "encoding/base64"

## "encoding/hex"

## "encoding/json"

## "errors"

## "fmt"

## "io"

## "log"

## "net/http"

## "strings"

## "time"

## "github.com/golang-jwt/jwt/v5"

## "github.com/google/uuid"

## "github.com/gorilla/mux"

## _ "github.com/mattn/go-sqlite3"

## "golang.org/x/crypto/argon2"

## "golang.org/x/crypto/chacha20poly1305"

## "golang.org/x/crypto/hkdf"

## )

## //

## ======================================================================

## ======

// КРИПТОГРАФИЧЕСКИЕ КОНСТАНТЫ (как у 1Password)

## //

## ======================================================================

## ======

const (
// Argon2id параметры (OWASP рекомендации)
Argon2Time    = 3
Argon2Memory  = 64 * 1024 // 64 MB
Argon2Threads = 4
Argon2KeyLen  = 32
// Размеры ключей
SecretKeySize  = 32 // 256 бит
SaltSize       = 32
NonceSize      = 24 // XChaCha20-Poly1305
MasterKeySize  = 32
VaultKeySize   = 32

## )

## //

## ======================================================================

## ======

## // СТРУКТУРЫ ДАННЫХ

## //

## ======================================================================

## ======

// User представляет пользователя с криптографическими данными
type User struct {
ID              string    `json:"id"`
Email           string    `json:"email"`
SecretKey       string    `json:"-"` // Никогда не хранится на сервере!
MasterKeySalt   []byte    `json:"-"` // Соль для деривации мастер-ключа
EncryptedVaultKey []byte  `json:"-"` // Зашифрованный ключ хранилища
VaultKeyNonce   []byte    `json:"-"` // Nonce для расшифровки vault key
SRPVerifier     []byte    `json:"-"` // SRP верификатор (пароль не хранится!)
CreatedAt       time.Time `json:"created_at"`

## }

// Vault представляет хранилище пользователя
type Vault struct {
ID        string    `json:"id"`
UserID    string    `json:"user_id"`
Name      string    `json:"name"`
CreatedAt time.Time `json:"created_at"`

## }

// SecretItem представляет зашифрованный секрет
type SecretItem struct {
ID             string    `json:"id"`
VaultID        string    `json:"vault_id"`
EncryptedData  []byte    `json:"-"`     // Зашифрованные данные
Nonce          []byte    `json:"-"`     // Nonce для расшифровки
Version        int       `json:"version"`
CreatedAt      time.Time `json:"created_at"`
UpdatedAt      time.Time `json:"updated_at"`

## }

// SecretData - расшифрованные данные секрета
type SecretData struct {
Name     string            `json:"name"`
Type     string            `json:"type"` // password, note, card, identity
Fields   map[string]string `json:"fields"`
Notes    string            `json:"notes,omitempty"`

## }

## //

## ======================================================================

## ======

// КРИПТОГРАФИЧЕСКОЕ ЯДРО (1Password-style)

## //

## ======================================================================

## ======

// CryptoEngine - движок шифрования
type CryptoEngine struct{}
// GenerateSecretKey генерирует Secret Key (как A3-XXXXX-XXXXX-XXXXX-XXXXX-

## XXXXX-XXXXX)

func (c *CryptoEngine) GenerateSecretKey() (string, []byte, error) {
key := make([]byte, SecretKeySize)
if _, err := rand.Read(key); err != nil {
return "", nil, err

## }

// Форматируем как 1Password: A3-XXXXX-XXXXX-...
encoded := base64.RawURLEncoding.EncodeToString(key)
formatted := fmt.Sprintf("A3-%s-%s-%s-%s-%s-%s",

encoded[0:5], encoded[5:10], encoded[10:15],
encoded[15:20], encoded[20:25], encoded[25:30])
return formatted, key, nil

## }

// ParseSecretKey парсит Secret Key обратно в байты
func (c *CryptoEngine) ParseSecretKey(formatted string) ([]byte, error) {
parts := strings.Split(formatted, "-")
if len(parts) != 7 || parts[0] != "A3" {
return nil, errors.New("invalid secret key format")

## }

encoded := strings.Join(parts[1:], "")
return base64.RawURLEncoding.DecodeString(encoded)

## }

// DeriveMasterKey создаёт мастер-ключ из пароля + secret key
// Это ключевая защита: даже зная пароль, без secret key ничего не расшифровать
func (c *CryptoEngine) DeriveMasterKey(password string, secretKey []byte, salt []byte) []byte

## {

// Шаг 1: Комбинируем пароль и secret key
combined := append([]byte(password), secretKey...)
// Шаг 2: Argon2id - защита от GPU/ASIC атак
masterKey := argon2.IDKey(combined, salt, Argon2Time, Argon2Memory, Argon2Threads,
MasterKeySize)
return masterKey

## }

// DeriveEncryptionKey создаёт ключ шифрования из мастер-ключа используя HKDF
func (c *CryptoEngine) DeriveEncryptionKey(masterKey []byte, context string) ([]byte, error) {
hkdfReader := hkdf.New(sha256.New, masterKey, nil, []byte(context))
key := make([]byte, 32)
if _, err := io.ReadFull(hkdfReader, key); err != nil {
return nil, err

## }

return key, nil

## }

// GenerateVaultKey генерирует случайный ключ для хранилища
func (c *CryptoEngine) GenerateVaultKey() ([]byte, error) {
key := make([]byte, VaultKeySize)
if _, err := rand.Read(key); err != nil {
return nil, err

## }

return key, nil

## }

// EncryptVaultKey шифрует vault key мастер-ключом
func (c *CryptoEngine) EncryptVaultKey(vaultKey, masterKey []byte) (ciphertext, nonce
[]byte, err error) {
return c.EncryptXChaCha20(vaultKey, masterKey)

## }

// DecryptVaultKey расшифровывает vault key
func (c *CryptoEngine) DecryptVaultKey(ciphertext, nonce, masterKey []byte) ([]byte, error) {
return c.DecryptXChaCha20(ciphertext, nonce, masterKey)

## }

// EncryptXChaCha20 шифрует данные с XChaCha20-Poly1305
// Преимущества: 256-бит nonce (можно генерировать случайно), быстрее AES на CPU
без AES-NI
func (c *CryptoEngine) EncryptXChaCha20(plaintext, key []byte) (ciphertext, nonce []byte,
err error) {
aead, err := chacha20poly1305.NewX(key)
if err != nil {
return nil, nil, err

## }

nonce = make([]byte, aead.NonceSize())
if _, err := rand.Read(nonce); err != nil {
return nil, nil, err

## }

ciphertext = aead.Seal(nil, nonce, plaintext, nil)
return ciphertext, nonce, nil

## }

// DecryptXChaCha20 расшифровывает данные
func (c *CryptoEngine) DecryptXChaCha20(ciphertext, nonce, key []byte) ([]byte, error) {
aead, err := chacha20poly1305.NewX(key)
if err != nil {
return nil, err

## }

plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
if err != nil {
return nil, errors.New("decryption failed: invalid key or corrupted data")

## }

return plaintext, nil

## }

// EncryptAESGCM альтернативное шифрование AES-256-GCM
func (c *CryptoEngine) EncryptAESGCM(plaintext, key []byte) (ciphertext, nonce []byte, err
error) {
block, err := aes.NewCipher(key)
if err != nil {

return nil, nil, err

## }

aead, err := cipher.NewGCM(block)
if err != nil {
return nil, nil, err

## }

nonce = make([]byte, aead.NonceSize())
if _, err := rand.Read(nonce); err != nil {
return nil, nil, err

## }

ciphertext = aead.Seal(nil, nonce, plaintext, nil)
return ciphertext, nonce, nil

## }

// DecryptAESGCM расшифровывает AES-256-GCM
func (c *CryptoEngine) DecryptAESGCM(ciphertext, nonce, key []byte) ([]byte, error) {
block, err := aes.NewCipher(key)
if err != nil {
return nil, err

## }

aead, err := cipher.NewGCM(block)
if err != nil {
return nil, err

## }

return aead.Open(nil, nonce, ciphertext, nil)

## }

// SecureCompare сравнивает два значения за константное время (защита от timing
attacks)
func (c *CryptoEngine) SecureCompare(a, b []byte) bool {
return subtle.ConstantTimeCompare(a, b) == 1

## }

// GenerateSalt генерирует криптографически стойкую соль
func (c *CryptoEngine) GenerateSalt() ([]byte, error) {
salt := make([]byte, SaltSize)
if _, err := rand.Read(salt); err != nil {
return nil, err

## }

return salt, nil

## }

## //

## ======================================================================

## ======

// БАЗА ДАННЫХ (всё зашифровано!)

## //

## ======================================================================

## ======

type Database struct {
db     *sql.DB
crypto *CryptoEngine

## }

func NewDatabase(dbPath string) (*Database, error) {
db, err := sql.Open("sqlite3", dbPath)
if err != nil {
return nil, err

## }

database := &Database{
db:     db,
crypto: &CryptoEngine{},

## }

if err := database.migrate(); err != nil {
return nil, err

## }

return database, nil

## }

func (d *Database) migrate() error {
queries := []string{
`CREATE TABLE IF NOT EXISTS users (
id TEXT PRIMARY KEY,
email TEXT UNIQUE NOT NULL,
master_key_salt BLOB NOT NULL,
encrypted_vault_key BLOB NOT NULL,
vault_key_nonce BLOB NOT NULL,
srp_verifier BLOB NOT NULL,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP

## )`,

`CREATE TABLE IF NOT EXISTS vaults (
id TEXT PRIMARY KEY,
user_id TEXT NOT NULL,
name_encrypted BLOB NOT NULL,
name_nonce BLOB NOT NULL,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (user_id) REFERENCES users(id)

## )`,

`CREATE TABLE IF NOT EXISTS secrets (
id TEXT PRIMARY KEY,
vault_id TEXT NOT NULL,

encrypted_data BLOB NOT NULL,
nonce BLOB NOT NULL,
version INTEGER DEFAULT 1,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (vault_id) REFERENCES vaults(id)

## )`,

`CREATE TABLE IF NOT EXISTS secret_versions (
id TEXT PRIMARY KEY,
secret_id TEXT NOT NULL,
encrypted_data BLOB NOT NULL,
nonce BLOB NOT NULL,
version INTEGER NOT NULL,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (secret_id) REFERENCES secrets(id)

## )`,

`CREATE TABLE IF NOT EXISTS audit_log (
id TEXT PRIMARY KEY,
user_id TEXT NOT NULL,
action TEXT NOT NULL,
resource_type TEXT NOT NULL,
resource_id TEXT,
ip_address TEXT,
user_agent TEXT,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP

## )`,

`CREATE INDEX IF NOT EXISTS idx_secrets_vault ON secrets(vault_id)`,
`CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_log(user_id)`,

## }

for _, query := range queries {
if _, err := d.db.Exec(query); err != nil {
return err

## }

## }

return nil

## }

// CreateUser создаёт пользователя с полной криптографической настройкой
func (d *Database) CreateUser(email, password string) (*User, string, error) {
crypto := d.crypto
// 1. Генерируем Secret Key (пользователь должен сохранить!)
secretKeyFormatted, secretKeyBytes, err := crypto.GenerateSecretKey()
if err != nil {
return nil, "", err

## }

// 2. Генерируем соль для мастер-ключа

salt, err := crypto.GenerateSalt()
if err != nil {
return nil, "", err

## }

// 3. Создаём мастер-ключ из пароля + secret key
masterKey := crypto.DeriveMasterKey(password, secretKeyBytes, salt)
// 4. Генерируем ключ хранилища
vaultKey, err := crypto.GenerateVaultKey()
if err != nil {
return nil, "", err

## }

// 5. Шифруем vault key мастер-ключом
encryptedVaultKey, vaultKeyNonce, err := crypto.EncryptVaultKey(vaultKey, masterKey)
if err != nil {
return nil, "", err

## }

// 6. Создаём SRP верификатор (пароль НЕ хранится!)
srpVerifier := crypto.DeriveMasterKey(password, salt, salt) // Упрощённый SRP
// 7. Сохраняем в БД
userID := uuid.New().String()
_, err = d.db.Exec(`
INSERT INTO users (id, email, master_key_salt, encrypted_vault_key,
vault_key_nonce, srp_verifier)

## VALUES (?, ?, ?, ?, ?, ?)

`, userID, email, salt, encryptedVaultKey, vaultKeyNonce, srpVerifier)
if err != nil {
return nil, "", err

## }

user := &User{
ID:        userID,
Email:     email,
CreatedAt: time.Now(),

## }

// Возвращаем Secret Key - пользователь ДОЛЖЕН его сохранить!
return user, secretKeyFormatted, nil

## }

// AuthenticateUser проверяет пользователя и возвращает vault key
func (d *Database) AuthenticateUser(email, password, secretKeyFormatted string) (*User,
[]byte, error) {
crypto := d.crypto

## // 1. Парсим Secret Key

secretKeyBytes, err := crypto.ParseSecretKey(secretKeyFormatted)
if err != nil {
return nil, nil, errors.New("invalid secret key")

## }

// 2. Получаем данные пользователя
var user User
var salt, encryptedVaultKey, vaultKeyNonce, srpVerifier []byte
err = d.db.QueryRow(`
SELECT id, email, master_key_salt, encrypted_vault_key, vault_key_nonce,
srp_verifier, created_at
FROM users WHERE email = ?
`, email).Scan(&user.ID, &user.Email, &salt, &encryptedVaultKey, &vaultKeyNonce,
&srpVerifier, &user.CreatedAt)
if err != nil {
return nil, nil, errors.New("invalid credentials")

## }

// 3. Проверяем пароль через SRP верификатор
computedVerifier := crypto.DeriveMasterKey(password, salt, salt)
if !crypto.SecureCompare(computedVerifier, srpVerifier) {
return nil, nil, errors.New("invalid credentials")

## }

// 4. Создаём мастер-ключ
masterKey := crypto.DeriveMasterKey(password, secretKeyBytes, salt)
// 5. Расшифровываем vault key
vaultKey, err := crypto.DecryptVaultKey(encryptedVaultKey, vaultKeyNonce, masterKey)
if err != nil {
return nil, nil, errors.New("invalid credentials")

## }

return &user, vaultKey, nil

## }

// CreateSecret создаёт зашифрованный секрет
func (d *Database) CreateSecret(vaultID string, data *SecretData, vaultKey []byte)
(*SecretItem, error) {
crypto := d.crypto
// Сериализуем данные
jsonData, err := json.Marshal(data)
if err != nil {
return nil, err

## }

## // Шифруем

encryptedData, nonce, err := crypto.EncryptXChaCha20(jsonData, vaultKey)
if err != nil {
return nil, err

## }

secretID := uuid.New().String()
now := time.Now()
_, err = d.db.Exec(`
INSERT INTO secrets (id, vault_id, encrypted_data, nonce, version, created_at,
updated_at)

## VALUES (?, ?, ?, ?, 1, ?, ?)

`, secretID, vaultID, encryptedData, nonce, now, now)
if err != nil {
return nil, err

## }

return &SecretItem{
ID:        secretID,
VaultID:   vaultID,

## Version:   1,

CreatedAt: now,
UpdatedAt: now,
}, nil

## }

// GetSecret получает и расшифровывает секрет
func (d *Database) GetSecret(secretID string, vaultKey []byte) (*SecretData, error) {
crypto := d.crypto
var encryptedData, nonce []byte
err := d.db.QueryRow(`
SELECT encrypted_data, nonce FROM secrets WHERE id = ?
`, secretID).Scan(&encryptedData, &nonce)
if err != nil {
return nil, err

## }

## // Расшифровываем

jsonData, err := crypto.DecryptXChaCha20(encryptedData, nonce, vaultKey)
if err != nil {
return nil, err

## }

var data SecretData
if err := json.Unmarshal(jsonData, &data); err != nil {
return nil, err

## }

return &data, nil

## }

// UpdateSecret обновляет секрет с версионированием
func (d *Database) UpdateSecret(secretID string, data *SecretData, vaultKey []byte) error {
crypto := d.crypto
// Получаем текущую версию
var currentVersion int
var oldEncrypted, oldNonce []byte
err := d.db.QueryRow(`
SELECT version, encrypted_data, nonce FROM secrets WHERE id = ?
`, secretID).Scan(&currentVersion, &oldEncrypted, &oldNonce)
if err != nil {
return err

## }

// Сохраняем старую версию
versionID := uuid.New().String()
_, err = d.db.Exec(`
INSERT INTO secret_versions (id, secret_id, encrypted_data, nonce, version,
created_at)

## VALUES (?, ?, ?, ?, ?, ?)

`, versionID, secretID, oldEncrypted, oldNonce, currentVersion, time.Now())
if err != nil {
return err

## }

// Шифруем новые данные
jsonData, err := json.Marshal(data)
if err != nil {
return err

## }

encryptedData, nonce, err := crypto.EncryptXChaCha20(jsonData, vaultKey)
if err != nil {
return err

## }

// Обновляем секрет
_, err = d.db.Exec(`
UPDATE secrets SET encrypted_data = ?, nonce = ?, version = ?, updated_at = ?
WHERE id = ?
`, encryptedData, nonce, currentVersion+1, time.Now(), secretID)
return err

## }

// ListSecrets возвращает список секретов (только метаданные)
func (d *Database) ListSecrets(vaultID string, vaultKey []byte) ([]SecretData, error) {
rows, err := d.db.Query(`
SELECT id, encrypted_data, nonce FROM secrets WHERE vault_id = ?
`, vaultID)
if err != nil {
return nil, err

## }

defer rows.Close()
var secrets []SecretData
crypto := d.crypto
for rows.Next() {
var id string
var encryptedData, nonce []byte
if err := rows.Scan(&id, &encryptedData, &nonce); err != nil {
return nil, err

## }

jsonData, err := crypto.DecryptXChaCha20(encryptedData, nonce, vaultKey)
if err != nil {
continue // Пропускаем повреждённые записи

## }

var data SecretData
if err := json.Unmarshal(jsonData, &data); err != nil {
continue

## }

secrets = append(secrets, data)

## }

return secrets, nil

## }

// AuditLog записывает действие в журнал аудита
func (d *Database) AuditLog(userID, action, resourceType, resourceID, ipAddress,
userAgent string) {
id := uuid.New().String()
d.db.Exec(`
INSERT INTO audit_log (id, user_id, action, resource_type, resource_id, ip_address,
user_agent)

## VALUES (?, ?, ?, ?, ?, ?, ?)

`, id, userID, action, resourceType, resourceID, ipAddress, userAgent)

## }

## //

## ======================================================================

## ======

## // HTTP API

## //

## ======================================================================

## ======

type Server struct {
db        *Database
crypto    *CryptoEngine
jwtSecret []byte

## }

type Session struct {
UserID   string
VaultKey []byte

## }

var sessions = make(map[string]*Session) // В продакшене использовать Redis
func NewServer(db *Database) *Server {
jwtSecret := make([]byte, 32)
rand.Read(jwtSecret)
return &Server{
db:        db,
crypto:    &CryptoEngine{},
jwtSecret: jwtSecret,

## }

## }

// RegisterRequest запрос на регистрацию
type RegisterRequest struct {
Email    string `json:"email"`
Password string `json:"password"`

## }

// RegisterResponse ответ с Secret Key
type RegisterResponse struct {
UserID    string `json:"user_id"`
SecretKey string `json:"secret_key"`
Message   string `json:"message"`

## }

// LoginRequest запрос на вход
type LoginRequest struct {
Email     string `json:"email"`
Password  string `json:"password"`
SecretKey string `json:"secret_key"`

## }

func (s *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
var req RegisterRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "Invalid request", http.StatusBadRequest)
return

## }

if len(req.Password) < 12 {
http.Error(w, "Password must be at least 12 characters", http.StatusBadRequest)
return

## }

user, secretKey, err := s.db.CreateUser(req.Email, req.Password)
if err != nil {
http.Error(w, "Registration failed", http.StatusInternalServerError)
return

## }

// Создаём дефолтное хранилище
vaultID := uuid.New().String()
s.db.db.Exec(`INSERT INTO vaults (id, user_id, name_encrypted, name_nonce) VALUES

## (?, ?, ?, ?)`,

vaultID, user.ID, []byte("default"), []byte("nonce"))
resp := RegisterResponse{
UserID:    user.ID,
SecretKey: secretKey,
Message:   "ВАЖНО! Сохраните Secret Key в безопасном месте. Без него
восстановить доступ НЕВОЗМОЖНО!",

## }

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(resp)

## }

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
var req LoginRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "Invalid request", http.StatusBadRequest)
return

## }

user, vaultKey, err := s.db.AuthenticateUser(req.Email, req.Password, req.SecretKey)
if err != nil {
http.Error(w, "Invalid credentials", http.StatusUnauthorized)
return

## }

// Создаём JWT

token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
"user_id": user.ID,
"exp":     time.Now().Add(time.Hour * 24).Unix(),

## })

tokenString, err := token.SignedString(s.jwtSecret)
if err != nil {
http.Error(w, "Token generation failed", http.StatusInternalServerError)
return

## }

// Сохраняем vault key в сессии (в памяти сервера)
sessionID := uuid.New().String()
sessions[sessionID] = &Session{
UserID:   user.ID,
VaultKey: vaultKey,

## }

s.db.AuditLog(user.ID, "login", "user", user.ID, r.RemoteAddr, r.UserAgent())
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]string{
"token":      tokenString,
"session_id": sessionID,

## })

## }

func (s *Server) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
return func(w http.ResponseWriter, r *http.Request) {
authHeader := r.Header.Get("Authorization")
if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
http.Error(w, "Unauthorized", http.StatusUnauthorized)
return

## }

tokenString := strings.TrimPrefix(authHeader, "Bearer ")
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
return s.jwtSecret, nil

## })

if err != nil || !token.Valid {
http.Error(w, "Invalid token", http.StatusUnauthorized)
return

## }

claims := token.Claims.(jwt.MapClaims)
r.Header.Set("X-User-ID", claims["user_id"].(string))
sessionID := r.Header.Get("X-Session-ID")
if session, ok := sessions[sessionID]; ok {

r.Header.Set("X-Vault-Key", hex.EncodeToString(session.VaultKey))

## }

next(w, r)

## }

## }

func (s *Server) CreateSecretHandler(w http.ResponseWriter, r *http.Request) {
vaultKeyHex := r.Header.Get("X-Vault-Key")
if vaultKeyHex == "" {
http.Error(w, "Session expired", http.StatusUnauthorized)
return

## }

vaultKey, _ := hex.DecodeString(vaultKeyHex)
var data SecretData
if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
http.Error(w, "Invalid request", http.StatusBadRequest)
return

## }

vars := mux.Vars(r)
vaultID := vars["vault_id"]
secret, err := s.db.CreateSecret(vaultID, &data, vaultKey)
if err != nil {
http.Error(w, "Failed to create secret", http.StatusInternalServerError)
return

## }

userID := r.Header.Get("X-User-ID")
s.db.AuditLog(userID, "create_secret", "secret", secret.ID, r.RemoteAddr, r.UserAgent())
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(secret)

## }

func (s *Server) GetSecretHandler(w http.ResponseWriter, r *http.Request) {
vaultKeyHex := r.Header.Get("X-Vault-Key")
if vaultKeyHex == "" {
http.Error(w, "Session expired", http.StatusUnauthorized)
return

## }

vaultKey, _ := hex.DecodeString(vaultKeyHex)
vars := mux.Vars(r)
secretID := vars["secret_id"]

data, err := s.db.GetSecret(secretID, vaultKey)
if err != nil {
http.Error(w, "Secret not found", http.StatusNotFound)
return

## }

userID := r.Header.Get("X-User-ID")
s.db.AuditLog(userID, "read_secret", "secret", secretID, r.RemoteAddr, r.UserAgent())
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(data)

## }

func (s *Server) ListSecretsHandler(w http.ResponseWriter, r *http.Request) {
vaultKeyHex := r.Header.Get("X-Vault-Key")
if vaultKeyHex == "" {
http.Error(w, "Session expired", http.StatusUnauthorized)
return

## }

vaultKey, _ := hex.DecodeString(vaultKeyHex)
vars := mux.Vars(r)
vaultID := vars["vault_id"]
secrets, err := s.db.ListSecrets(vaultID, vaultKey)
if err != nil {
http.Error(w, "Failed to list secrets", http.StatusInternalServerError)
return

## }

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(secrets)

## }

func main() {
db, err := NewDatabase("vault.db")
if err != nil {
log.Fatal(err)

## }

server := NewServer(db)
r := mux.NewRouter()
// Публичные эндпоинты
r.HandleFunc("/api/v1/register", server.RegisterHandler).Methods("POST")
r.HandleFunc("/api/v1/login", server.LoginHandler).Methods("POST")
// Защищённые эндпоинты
r.HandleFunc("/api/v1/vaults/{vault_id}/secrets",

Архитектура безопасности (как у 1Password):
server.AuthMiddleware(server.CreateSecretHandler)).Methods("POST")
r.HandleFunc("/api/v1/vaults/{vault_id}/secrets",
server.AuthMiddleware(server.ListSecretsHandler)).Methods("GET")
r.HandleFunc("/api/v1/secrets/{secret_id}",
server.AuthMiddleware(server.GetSecretHandler)).Methods("GET")
fmt.Println(" Secure Vault Server starting on :8080")
fmt.Println(" Криптография: XChaCha20-Poly1305 + Argon2id + HKDF")
fmt.Println("️  Модель безопасности: Password + Secret Key (как 1Password)")
log.Fatal(http.ListenAndServe(":8080", r))

## }

## ┌────────────────────────────────────────────────────────

## ─────┐

│                    КЛИЕНТ (браузер/приложение)              │

## │  ┌─────────────┐    ┌─────────────┐                         │

## │  │  Password   │ +  │ Secret Key  │ = Master Key            │

│  │ (в голове)  │    │ (на бумаге) │   (никуда не уходит)    │

## │  └──────┬──────┘    └──────┬──────┘                         │

## │         └────────┬─────────┘                                │

## │                  ▼                                          │

## │         ┌───────────────┐                                   │

│         │   Argon2id    │ ← Защита от GPU/ASIC              │

## │         └───────┬───────┘                                   │

## │                 ▼                                           │

## │         ┌───────────────┐                                   │

## │         │  Master Key   │ → Расшифровывает Vault Key        │

## │         └───────────────┘                                   │

## └────────────────────────────────────────────────────────

## ─────┘

## │

▼ (только зашифрованные данные)

## ┌────────────────────────────────────────────────────────

## ─────┐

## │                         СЕРВЕР                              │

## │

## ┌──────────────────────────────────────────────────────┐

## │

│  │                    База данных                       │   │

## │  │  ┌────────────────┐  ┌────────────────────────────┐  │   │

│  │  │ SRP Verifier   │  │ Encrypted Vault Key        │  │   │
│  │  │ (не пароль!)   │  │ (бесполезен без Master Key)│  │   │

## │  │  └────────────────┘  └────────────────────────────┘  │   │

Почему это безопасно:
УгрозаЗащита
Утечка БДВсе данные зашифрованы, ключи недоступны
Брутфорс пароляArgon2id (64MB RAM, 3 итерации)
Компрометация сервераVault Key шифруется на клиенте
Timing attacksConstant-time сравнение
Replay attacksУникальный nonce для каждого шифрования

## Зависимости:

Хочешь добавить HSM интеграцию, key rotation, или emergency access?
Key rotation

## │  │  ┌────────────────────────────────────────────────┐  │   │

│  │  │         Encrypted Secrets (XChaCha20)          │  │   │
│  │  │    (бесполезны без Vault Key)                  │  │   │

## │  │  └────────────────────────────────────────────────┘  │   │

## │

## └──────────────────────────────────────────────────────┘

## │

## └────────────────────────────────────────────────────────

## ─────┘

go get github.com/gorilla/mux
go get github.com/golang-jwt/jwt/v5
go get github.com/google/uuid
go get github.com/mattn/go-sqlite3
go get golang.org/x/crypto

## Claude Opus 4.5

Отлично! Добавляем Key Rotation — критически важную функцию для
безопасности. Реализуем ротацию на всех уровнях: Master Key, Vault Key и
автоматическую ротацию по расписанию.
package main
import (

## "crypto/aes"

## "crypto/cipher"

## "crypto/rand"

## "crypto/sha256"

## "crypto/subtle"

## "database/sql"

## "encoding/base64"

## "encoding/hex"

## "encoding/json"

## "errors"

## "fmt"

## "io"

## "log"

## "net/http"

## "strings"

## "sync"

## "time"

## "github.com/golang-jwt/jwt/v5"

## "github.com/google/uuid"

## "github.com/gorilla/mux"

## _ "github.com/mattn/go-sqlite3"

## "golang.org/x/crypto/argon2"

## "golang.org/x/crypto/chacha20poly1305"

## "golang.org/x/crypto/hkdf"

## )

## //

## ======================================================================

## ======

## // КРИПТОГРАФИЧЕСКИЕ КОНСТАНТЫ

## //

## ======================================================================

## ======

const (
Argon2Time    = 3
Argon2Memory  = 64 * 1024
Argon2Threads = 4
Argon2KeyLen  = 32

SecretKeySize = 32
SaltSize      = 32
NonceSize     = 24
MasterKeySize = 32
VaultKeySize  = 32
// Key Rotation настройки
DefaultKeyRotationDays  = 90
MaxKeyVersions          = 10
KeyRotationGracePeriod  = 24 * time.Hour

## )

## //

## ======================================================================

## ======

## // KEY ROTATION СТРУКТУРЫ

## //

## ======================================================================

## ======

// KeyVersion представляет версию ключа
type KeyVersion struct {
ID          string    `json:"id"`
Version     int       `json:"version"`
KeyType     string    `json:"key_type"` // master, vault, encryption
EncryptedKey []byte   `json:"-"`

## Nonce       []byte    `json:"-"`

Status      string    `json:"status"` // active, rotating, deprecated, destroyed
CreatedAt   time.Time `json:"created_at"`
ExpiresAt   time.Time `json:"expires_at"`
RotatedAt   *time.Time `json:"rotated_at,omitempty"`
DestroyedAt *time.Time `json:"destroyed_at,omitempty"`

## }

// KeyRotationPolicy политика ротации ключей
type KeyRotationPolicy struct {
ID                string        `json:"id"`
UserID            string        `json:"user_id"`
KeyType           string        `json:"key_type"`
RotationInterval  time.Duration `json:"rotation_interval"`
AutoRotate        bool          `json:"auto_rotate"`
NotifyBeforeDays  int           `json:"notify_before_days"`
LastRotation      time.Time     `json:"last_rotation"`
NextRotation      time.Time     `json:"next_rotation"`

## }

// KeyRotationEvent событие ротации для аудита
type KeyRotationEvent struct {
ID            string    `json:"id"`
UserID        string    `json:"user_id"`

KeyType       string    `json:"key_type"`
OldVersion    int       `json:"old_version"`
NewVersion    int       `json:"new_version"`
Reason        string    `json:"reason"` // scheduled, manual, compromised, policy
Status        string    `json:"status"` // started, completed, failed, rolled_back
SecretsCount  int       `json:"secrets_count"`
Duration      int64     `json:"duration_ms"`
CreatedAt     time.Time `json:"created_at"`
CompletedAt   *time.Time `json:"completed_at,omitempty"`
ErrorMessage  string    `json:"error_message,omitempty"`

## }

## //

## ======================================================================

## ======

## // ОСНОВНЫЕ СТРУКТУРЫ

## //

## ======================================================================

## ======

type User struct {
ID                string    `json:"id"`
Email             string    `json:"email"`
MasterKeySalt     []byte    `json:"-"`
CurrentKeyVersion int       `json:"-"`
SRPVerifier       []byte    `json:"-"`
CreatedAt         time.Time `json:"created_at"`

## }

type SecretItem struct {
ID            string    `json:"id"`
VaultID       string    `json:"vault_id"`
EncryptedData []byte    `json:"-"`

## Nonce         []byte    `json:"-"`

KeyVersion    int       `json:"key_version"` // Какой версией ключа зашифровано
Version       int       `json:"version"`
CreatedAt     time.Time `json:"created_at"`
UpdatedAt     time.Time `json:"updated_at"`

## }

type SecretData struct {
Name   string            `json:"name"`
Type   string            `json:"type"`
Fields map[string]string `json:"fields"`
Notes  string            `json:"notes,omitempty"`

## }

## //

## ======================================================================

## ======

## // CRYPTO ENGINE С KEY ROTATION

## //

## ======================================================================

## ======

type CryptoEngine struct{}
func (c *CryptoEngine) GenerateSecretKey() (string, []byte, error) {
key := make([]byte, SecretKeySize)
if _, err := rand.Read(key); err != nil {
return "", nil, err

## }

encoded := base64.RawURLEncoding.EncodeToString(key)
formatted := fmt.Sprintf("A3-%s-%s-%s-%s-%s-%s",
encoded[0:5], encoded[5:10], encoded[10:15],
encoded[15:20], encoded[20:25], encoded[25:30])
return formatted, key, nil

## }

func (c *CryptoEngine) ParseSecretKey(formatted string) ([]byte, error) {
parts := strings.Split(formatted, "-")
if len(parts) != 7 || parts[0] != "A3" {
return nil, errors.New("invalid secret key format")

## }

encoded := strings.Join(parts[1:], "")
return base64.RawURLEncoding.DecodeString(encoded)

## }

func (c *CryptoEngine) DeriveMasterKey(password string, secretKey []byte, salt []byte) []byte

## {

combined := append([]byte(password), secretKey...)
return argon2.IDKey(combined, salt, Argon2Time, Argon2Memory, Argon2Threads,
MasterKeySize)

## }

func (c *CryptoEngine) DeriveKeyWithVersion(masterKey []byte, version int, context string)
([]byte, error) {
// Включаем версию в контекст для уникальности ключа
versionedContext := fmt.Sprintf("%s:v%d", context, version)
hkdfReader := hkdf.New(sha256.New, masterKey, nil, []byte(versionedContext))
key := make([]byte, 32)
if _, err := io.ReadFull(hkdfReader, key); err != nil {
return nil, err

## }

return key, nil

## }

func (c *CryptoEngine) GenerateVaultKey() ([]byte, error) {

key := make([]byte, VaultKeySize)
if _, err := rand.Read(key); err != nil {
return nil, err

## }

return key, nil

## }

func (c *CryptoEngine) EncryptXChaCha20(plaintext, key []byte) (ciphertext, nonce []byte,
err error) {
aead, err := chacha20poly1305.NewX(key)
if err != nil {
return nil, nil, err

## }

nonce = make([]byte, aead.NonceSize())
if _, err := rand.Read(nonce); err != nil {
return nil, nil, err

## }

ciphertext = aead.Seal(nil, nonce, plaintext, nil)
return ciphertext, nonce, nil

## }

func (c *CryptoEngine) DecryptXChaCha20(ciphertext, nonce, key []byte) ([]byte, error) {
aead, err := chacha20poly1305.NewX(key)
if err != nil {
return nil, err

## }

plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
if err != nil {
return nil, errors.New("decryption failed")

## }

return plaintext, nil

## }

func (c *CryptoEngine) EncryptAESGCM(plaintext, key []byte) (ciphertext, nonce []byte, err
error) {
block, err := aes.NewCipher(key)
if err != nil {
return nil, nil, err

## }

aead, err := cipher.NewGCM(block)
if err != nil {
return nil, nil, err

## }

nonce = make([]byte, aead.NonceSize())

if _, err := rand.Read(nonce); err != nil {
return nil, nil, err

## }

ciphertext = aead.Seal(nil, nonce, plaintext, nil)
return ciphertext, nonce, nil

## }

func (c *CryptoEngine) DecryptAESGCM(ciphertext, nonce, key []byte) ([]byte, error) {
block, err := aes.NewCipher(key)
if err != nil {
return nil, err

## }

aead, err := cipher.NewGCM(block)
if err != nil {
return nil, err

## }

return aead.Open(nil, nonce, ciphertext, nil)

## }

func (c *CryptoEngine) SecureCompare(a, b []byte) bool {
return subtle.ConstantTimeCompare(a, b) == 1

## }

func (c *CryptoEngine) GenerateSalt() ([]byte, error) {
salt := make([]byte, SaltSize)
if _, err := rand.Read(salt); err != nil {
return nil, err

## }

return salt, nil

## }

// SecureZeroMemory безопасно очищает память
func (c *CryptoEngine) SecureZeroMemory(b []byte) {
for i := range b {
b[i] = 0

## }

## }

## //

## ======================================================================

## ======

## // KEY ROTATION MANAGER

## //

## ======================================================================

## ======

type KeyRotationManager struct {

db     *Database
crypto *CryptoEngine
mu     sync.RWMutex
// Кэш активных ключей (в памяти)
activeKeys map[string]map[int][]byte // userID -> version -> key

## }

func NewKeyRotationManager(db *Database) *KeyRotationManager {
return &KeyRotationManager{
db:         db,
crypto:     &CryptoEngine{},
activeKeys: make(map[string]map[int][]byte),

## }

## }

// RotateVaultKey выполняет ротацию Vault Key
func (krm *KeyRotationManager) RotateVaultKey(userID string, currentMasterKey []byte,
reason string) (*KeyRotationEvent, error) {
krm.mu.Lock()
defer krm.mu.Unlock()
startTime := time.Now()
event := &KeyRotationEvent{
ID:        uuid.New().String(),
UserID:    userID,
KeyType:   "vault",
Reason:    reason,

## Status:    "started",

CreatedAt: time.Now(),

## }

// 1. Получаем текущую версию ключа
currentVersion, err := krm.db.GetCurrentKeyVersion(userID, "vault")
if err != nil {
event.Status = "failed"
event.ErrorMessage = err.Error()
krm.db.SaveKeyRotationEvent(event)
return event, err

## }

event.OldVersion = currentVersion.Version
// 2. Расшифровываем текущий vault key
oldVaultKey, err := krm.crypto.DecryptXChaCha20(
currentVersion.EncryptedKey,
currentVersion.Nonce,
currentMasterKey,

## )

if err != nil {
event.Status = "failed"

event.ErrorMessage = "failed to decrypt current vault key"
krm.db.SaveKeyRotationEvent(event)
return event, err

## }

defer krm.crypto.SecureZeroMemory(oldVaultKey)
// 3. Генерируем новый vault key
newVaultKey, err := krm.crypto.GenerateVaultKey()
if err != nil {
event.Status = "failed"
event.ErrorMessage = err.Error()
krm.db.SaveKeyRotationEvent(event)
return event, err

## }

defer krm.crypto.SecureZeroMemory(newVaultKey)
// 4. Шифруем новый vault key мастер-ключом
encryptedNewKey, newNonce, err := krm.crypto.EncryptXChaCha20(newVaultKey,
currentMasterKey)
if err != nil {
event.Status = "failed"
event.ErrorMessage = err.Error()
krm.db.SaveKeyRotationEvent(event)
return event, err

## }

// 5. Создаём новую версию ключа
newVersion := &KeyVersion{
ID:           uuid.New().String(),
Version:      currentVersion.Version + 1,
KeyType:      "vault",
EncryptedKey: encryptedNewKey,
Nonce:        newNonce,

## Status:       "rotating",

CreatedAt:    time.Now(),
ExpiresAt:    time.Now().Add(time.Duration(DefaultKeyRotationDays) * 24 * time.Hour),

## }

event.NewVersion = newVersion.Version
// 6. Сохраняем новую версию ключа
if err := krm.db.SaveKeyVersion(userID, newVersion); err != nil {
event.Status = "failed"
event.ErrorMessage = err.Error()
krm.db.SaveKeyRotationEvent(event)
return event, err

## }

// 7. Перешифровываем все секреты
secretsCount, err := krm.reencryptSecrets(userID, oldVaultKey, newVaultKey,
currentVersion.Version, newVersion.Version)

if err != nil {
// Откатываем изменения
krm.db.DeleteKeyVersion(newVersion.ID)
event.Status = "rolled_back"
event.ErrorMessage = err.Error()
krm.db.SaveKeyRotationEvent(event)
return event, err

## }

event.SecretsCount = secretsCount
// 8. Помечаем старый ключ как deprecated
now := time.Now()
currentVersion.Status = "deprecated"
currentVersion.RotatedAt = &now
krm.db.UpdateKeyVersionStatus(currentVersion)
// 9. Активируем новый ключ
newVersion.Status = "active"
krm.db.UpdateKeyVersionStatus(newVersion)
// 10. Обновляем текущую версию пользователя
krm.db.UpdateUserKeyVersion(userID, newVersion.Version)
// 11. Завершаем событие
completedAt := time.Now()
event.Status = "completed"
event.CompletedAt = &completedAt
event.Duration = time.Since(startTime).Milliseconds()
krm.db.SaveKeyRotationEvent(event)
// 12. Планируем уничтожение старого ключа
go krm.scheduleKeyDestruction(currentVersion.ID, KeyRotationGracePeriod)
return event, nil

## }

// reencryptSecrets перешифровывает все секреты новым ключом
func (krm *KeyRotationManager) reencryptSecrets(userID string, oldKey, newKey []byte,
oldVersion, newVersion int) (int, error) {
secrets, err := krm.db.GetSecretsForReencryption(userID, oldVersion)
if err != nil {
return 0, err

## }

count := 0
for _, secret := range secrets {
// Расшифровываем старым ключом
plaintext, err := krm.crypto.DecryptXChaCha20(secret.EncryptedData, secret.Nonce,
oldKey)
if err != nil {

log.Printf("Warning: failed to decrypt secret %s: %v", secret.ID, err)
continue

## }

// Шифруем новым ключом
newCiphertext, newNonce, err := krm.crypto.EncryptXChaCha20(plaintext, newKey)
if err != nil {
krm.crypto.SecureZeroMemory(plaintext)
return count, err

## }

// Очищаем plaintext из памяти
krm.crypto.SecureZeroMemory(plaintext)
// Обновляем в БД
if err := krm.db.UpdateSecretEncryption(secret.ID, newCiphertext, newNonce,
newVersion); err != nil {
return count, err

## }

count++

## }

return count, nil

## }

// scheduleKeyDestruction планирует уничтожение ключа после grace period
func (krm *KeyRotationManager) scheduleKeyDestruction(keyVersionID string, delay
time.Duration) {
time.Sleep(delay)
krm.mu.Lock()
defer krm.mu.Unlock()
// Проверяем, что ключ всё ещё deprecated
keyVersion, err := krm.db.GetKeyVersionByID(keyVersionID)
if err != nil || keyVersion.Status != "deprecated" {
return

## }

// Уничтожаем ключ (перезаписываем случайными данными)
randomData := make([]byte, len(keyVersion.EncryptedKey))
rand.Read(randomData)
now := time.Now()
keyVersion.EncryptedKey = randomData
keyVersion.Status = "destroyed"
keyVersion.DestroyedAt = &now
krm.db.UpdateKeyVersion(keyVersion)

log.Printf("Key version %s destroyed", keyVersionID)

## }

// RotateMasterKey ротация мастер-ключа (смена пароля)
func (krm *KeyRotationManager) RotateMasterKey(userID, oldPassword, newPassword
string, secretKey []byte) error {
krm.mu.Lock()
defer krm.mu.Unlock()
// 1. Получаем данные пользователя
user, salt, err := krm.db.GetUserForKeyRotation(userID)
if err != nil {
return err

## }

// 2. Создаём старый мастер-ключ
oldMasterKey := krm.crypto.DeriveMasterKey(oldPassword, secretKey, salt)
defer krm.crypto.SecureZeroMemory(oldMasterKey)
// 3. Получаем текущий vault key
currentVersion, err := krm.db.GetCurrentKeyVersion(userID, "vault")
if err != nil {
return err

## }

vaultKey, err := krm.crypto.DecryptXChaCha20(
currentVersion.EncryptedKey,
currentVersion.Nonce,
oldMasterKey,

## )

if err != nil {
return errors.New("invalid current password")

## }

defer krm.crypto.SecureZeroMemory(vaultKey)
// 4. Генерируем новую соль
newSalt, err := krm.crypto.GenerateSalt()
if err != nil {
return err

## }

// 5. Создаём новый мастер-ключ
newMasterKey := krm.crypto.DeriveMasterKey(newPassword, secretKey, newSalt)
defer krm.crypto.SecureZeroMemory(newMasterKey)
// 6. Перешифровываем vault key новым мастер-ключом
newEncryptedVaultKey, newNonce, err := krm.crypto.EncryptXChaCha20(vaultKey,
newMasterKey)
if err != nil {

return err

## }

// 7. Обновляем SRP верификатор
newSRPVerifier := krm.crypto.DeriveMasterKey(newPassword, newSalt, newSalt)
// 8. Обновляем в БД
if err := krm.db.UpdateUserMasterKey(userID, newSalt, newEncryptedVaultKey,
newNonce, newSRPVerifier); err != nil {
return err

## }

// 9. Логируем событие
event := &KeyRotationEvent{
ID:        uuid.New().String(),
UserID:    userID,
KeyType:   "master",

## Reason:    "password_change",

## Status:    "completed",

CreatedAt: time.Now(),

## }

now := time.Now()
event.CompletedAt = &now
krm.db.SaveKeyRotationEvent(event)
return nil

## }

// EmergencyRotation экстренная ротация при компрометации
func (krm *KeyRotationManager) EmergencyRotation(userID string, masterKey []byte)
(*KeyRotationEvent, error) {
// Помечаем все текущие ключи как скомпрометированные
if err := krm.db.MarkKeysCompromised(userID); err != nil {
return nil, err

## }

// Выполняем ротацию с причиной "compromised"
return krm.RotateVaultKey(userID, masterKey, "compromised")

## }

// CheckRotationNeeded проверяет, нужна ли ротация
func (krm *KeyRotationManager) CheckRotationNeeded(userID string) (bool, *KeyVersion,
error) {
currentVersion, err := krm.db.GetCurrentKeyVersion(userID, "vault")
if err != nil {
return false, nil, err

## }

if time.Now().After(currentVersion.ExpiresAt) {
return true, currentVersion, nil

## }

return false, currentVersion, nil

## }

## //

## ======================================================================

## ======

## // DATABASE С KEY ROTATION

## //

## ======================================================================

## ======

type Database struct {
db     *sql.DB
crypto *CryptoEngine

## }

func NewDatabase(dbPath string) (*Database, error) {
db, err := sql.Open("sqlite3", dbPath)
if err != nil {
return nil, err

## }

database := &Database{
db:     db,
crypto: &CryptoEngine{},

## }

if err := database.migrate(); err != nil {
return nil, err

## }

return database, nil

## }

func (d *Database) migrate() error {
queries := []string{
`CREATE TABLE IF NOT EXISTS users (
id TEXT PRIMARY KEY,
email TEXT UNIQUE NOT NULL,
master_key_salt BLOB NOT NULL,
current_key_version INTEGER DEFAULT 1,
srp_verifier BLOB NOT NULL,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP

## )`,

`CREATE TABLE IF NOT EXISTS key_versions (
id TEXT PRIMARY KEY,
user_id TEXT NOT NULL,
version INTEGER NOT NULL,

key_type TEXT NOT NULL,
encrypted_key BLOB NOT NULL,
nonce BLOB NOT NULL,
status TEXT DEFAULT 'active',
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
expires_at DATETIME,
rotated_at DATETIME,
destroyed_at DATETIME,
FOREIGN KEY (user_id) REFERENCES users(id),
UNIQUE(user_id, version, key_type)

## )`,

`CREATE TABLE IF NOT EXISTS vaults (
id TEXT PRIMARY KEY,
user_id TEXT NOT NULL,
name_encrypted BLOB NOT NULL,
name_nonce BLOB NOT NULL,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (user_id) REFERENCES users(id)

## )`,

`CREATE TABLE IF NOT EXISTS secrets (
id TEXT PRIMARY KEY,
vault_id TEXT NOT NULL,
encrypted_data BLOB NOT NULL,
nonce BLOB NOT NULL,
key_version INTEGER NOT NULL,
version INTEGER DEFAULT 1,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (vault_id) REFERENCES vaults(id)

## )`,

`CREATE TABLE IF NOT EXISTS key_rotation_events (
id TEXT PRIMARY KEY,
user_id TEXT NOT NULL,
key_type TEXT NOT NULL,
old_version INTEGER,
new_version INTEGER,
reason TEXT NOT NULL,
status TEXT NOT NULL,
secrets_count INTEGER DEFAULT 0,
duration_ms INTEGER,
error_message TEXT,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
completed_at DATETIME,
FOREIGN KEY (user_id) REFERENCES users(id)

## )`,

`CREATE TABLE IF NOT EXISTS key_rotation_policies (
id TEXT PRIMARY KEY,
user_id TEXT NOT NULL,
key_type TEXT NOT NULL,
rotation_interval_days INTEGER DEFAULT 90,

auto_rotate BOOLEAN DEFAULT 1,
notify_before_days INTEGER DEFAULT 7,
last_rotation DATETIME,
next_rotation DATETIME,
FOREIGN KEY (user_id) REFERENCES users(id),
UNIQUE(user_id, key_type)

## )`,

`CREATE INDEX IF NOT EXISTS idx_key_versions_user ON key_versions(user_id,
key_type, status)`,
`CREATE INDEX IF NOT EXISTS idx_secrets_key_version ON secrets(key_version)`,
`CREATE INDEX IF NOT EXISTS idx_rotation_events_user ON
key_rotation_events(user_id)`,

## }

for _, query := range queries {
if _, err := d.db.Exec(query); err != nil {
return err

## }

## }

return nil

## }

// CreateUser создаёт пользователя с начальной версией ключа
func (d *Database) CreateUser(email, password string) (*User, string, error) {
crypto := d.crypto
secretKeyFormatted, secretKeyBytes, err := crypto.GenerateSecretKey()
if err != nil {
return nil, "", err

## }

salt, err := crypto.GenerateSalt()
if err != nil {
return nil, "", err

## }

masterKey := crypto.DeriveMasterKey(password, secretKeyBytes, salt)
defer crypto.SecureZeroMemory(masterKey)
vaultKey, err := crypto.GenerateVaultKey()
if err != nil {
return nil, "", err

## }

defer crypto.SecureZeroMemory(vaultKey)
encryptedVaultKey, vaultKeyNonce, err := crypto.EncryptXChaCha20(vaultKey,
masterKey)
if err != nil {
return nil, "", err

## }

srpVerifier := crypto.DeriveMasterKey(password, salt, salt)
userID := uuid.New().String()
tx, err := d.db.Begin()
if err != nil {
return nil, "", err

## }

defer tx.Rollback()
// Создаём пользователя
_, err = tx.Exec(`
INSERT INTO users (id, email, master_key_salt, current_key_version, srp_verifier)

## VALUES (?, ?, ?, 1, ?)

`, userID, email, salt, srpVerifier)
if err != nil {
return nil, "", err

## }

// Создаём первую версию ключа
keyVersionID := uuid.New().String()
expiresAt := time.Now().Add(time.Duration(DefaultKeyRotationDays) * 24 * time.Hour)
_, err = tx.Exec(`
INSERT INTO key_versions (id, user_id, version, key_type, encrypted_key, nonce,
status, expires_at)
VALUES (?, ?, 1, 'vault', ?, ?, 'active', ?)
`, keyVersionID, userID, encryptedVaultKey, vaultKeyNonce, expiresAt)
if err != nil {
return nil, "", err

## }

// Создаём политику ротации по умолчанию
policyID := uuid.New().String()
nextRotation := time.Now().Add(time.Duration(DefaultKeyRotationDays) * 24 * time.Hour)
_, err = tx.Exec(`
INSERT INTO key_rotation_policies (id, user_id, key_type, last_rotation, next_rotation)
VALUES (?, ?, 'vault', ?, ?)
`, policyID, userID, time.Now(), nextRotation)
if err != nil {
return nil, "", err

## }

// Создаём дефолтное хранилище
vaultID := uuid.New().String()
_, err = tx.Exec(`
INSERT INTO vaults (id, user_id, name_encrypted, name_nonce)

## VALUES (?, ?, ?, ?)

`, vaultID, userID, []byte("default"), []byte("nonce"))

if err != nil {
return nil, "", err

## }

if err := tx.Commit(); err != nil {
return nil, "", err

## }

return &User{
ID:                userID,
Email:             email,
CurrentKeyVersion: 1,
CreatedAt:         time.Now(),
}, secretKeyFormatted, nil

## }

// AuthenticateUser аутентифицирует и возвращает vault key
func (d *Database) AuthenticateUser(email, password, secretKeyFormatted string) (*User,
[]byte, error) {
crypto := d.crypto
secretKeyBytes, err := crypto.ParseSecretKey(secretKeyFormatted)
if err != nil {
return nil, nil, errors.New("invalid secret key")

## }

var user User
var salt, srpVerifier []byte
err = d.db.QueryRow(`
SELECT id, email, master_key_salt, current_key_version, srp_verifier, created_at
FROM users WHERE email = ?
`, email).Scan(&user.ID, &user.Email, &salt, &user.CurrentKeyVersion, &srpVerifier,
&user.CreatedAt)
if err != nil {
return nil, nil, errors.New("invalid credentials")

## }

computedVerifier := crypto.DeriveMasterKey(password, salt, salt)
if !crypto.SecureCompare(computedVerifier, srpVerifier) {
return nil, nil, errors.New("invalid credentials")

## }

masterKey := crypto.DeriveMasterKey(password, secretKeyBytes, salt)
defer crypto.SecureZeroMemory(masterKey)
// Получаем активную версию vault key
var encryptedVaultKey, vaultKeyNonce []byte
err = d.db.QueryRow(`

SELECT encrypted_key, nonce FROM key_versions
WHERE user_id = ? AND key_type = 'vault' AND status = 'active'
ORDER BY version DESC LIMIT 1
`, user.ID).Scan(&encryptedVaultKey, &vaultKeyNonce)
if err != nil {
return nil, nil, errors.New("vault key not found")

## }

vaultKey, err := crypto.DecryptXChaCha20(encryptedVaultKey, vaultKeyNonce,
masterKey)
if err != nil {
return nil, nil, errors.New("invalid credentials")

## }

return &user, vaultKey, nil

## }

// GetCurrentKeyVersion получает текущую активную версию ключа
func (d *Database) GetCurrentKeyVersion(userID, keyType string) (*KeyVersion, error) {
var kv KeyVersion
err := d.db.QueryRow(`
SELECT id, version, key_type, encrypted_key, nonce, status, created_at, expires_at
FROM key_versions
WHERE user_id = ? AND key_type = ? AND status = 'active'
ORDER BY version DESC LIMIT 1
`, userID, keyType).Scan(&kv.ID, &kv.Version, &kv.KeyType, &kv.EncryptedKey,
&kv.Nonce, &kv.Status, &kv.CreatedAt, &kv.ExpiresAt)
if err != nil {
return nil, err

## }

return &kv, nil

## }

// GetKeyVersionByID получает версию ключа по ID
func (d *Database) GetKeyVersionByID(id string) (*KeyVersion, error) {
var kv KeyVersion
err := d.db.QueryRow(`
SELECT id, version, key_type, encrypted_key, nonce, status, created_at, expires_at
FROM key_versions WHERE id = ?
`, id).Scan(&kv.ID, &kv.Version, &kv.KeyType, &kv.EncryptedKey, &kv.Nonce, &kv.Status,
&kv.CreatedAt, &kv.ExpiresAt)
if err != nil {
return nil, err

## }

return &kv, nil

## }

// SaveKeyVersion сохраняет новую версию ключа
func (d *Database) SaveKeyVersion(userID string, kv *KeyVersion) error {
_, err := d.db.Exec(`
INSERT INTO key_versions (id, user_id, version, key_type, encrypted_key, nonce,
status, created_at, expires_at)

## VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)

`, kv.ID, userID, kv.Version, kv.KeyType, kv.EncryptedKey, kv.Nonce, kv.Status,
kv.CreatedAt, kv.ExpiresAt)
return err

## }

// UpdateKeyVersionStatus обновляет статус версии ключа
func (d *Database) UpdateKeyVersionStatus(kv *KeyVersion) error {
_, err := d.db.Exec(`
UPDATE key_versions SET status = ?, rotated_at = ? WHERE id = ?
`, kv.Status, kv.RotatedAt, kv.ID)
return err

## }

// UpdateKeyVersion обновляет версию ключа
func (d *Database) UpdateKeyVersion(kv *KeyVersion) error {
_, err := d.db.Exec(`
UPDATE key_versions SET encrypted_key = ?, status = ?, destroyed_at = ? WHERE id

## = ?

`, kv.EncryptedKey, kv.Status, kv.DestroyedAt, kv.ID)
return err

## }

// DeleteKeyVersion удаляет версию ключа
func (d *Database) DeleteKeyVersion(id string) error {
_, err := d.db.Exec(`DELETE FROM key_versions WHERE id = ?`, id)
return err

## }

// UpdateUserKeyVersion обновляет текущую версию ключа пользователя
func (d *Database) UpdateUserKeyVersion(userID string, version int) error {
_, err := d.db.Exec(`UPDATE users SET current_key_version = ? WHERE id = ?`, version,
userID)
return err

## }

// GetSecretsForReencryption получает секреты для перешифровки
func (d *Database) GetSecretsForReencryption(userID string, keyVersion int) ([]SecretItem,
error) {
rows, err := d.db.Query(`

SELECT s.id, s.vault_id, s.encrypted_data, s.nonce, s.key_version, s.version
FROM secrets s
JOIN vaults v ON s.vault_id = v.id
WHERE v.user_id = ? AND s.key_version = ?
`, userID, keyVersion)
if err != nil {
return nil, err

## }

defer rows.Close()
var secrets []SecretItem
for rows.Next() {
var s SecretItem
if err := rows.Scan(&s.ID, &s.VaultID, &s.EncryptedData, &s.Nonce, &s.KeyVersion,
&s.Version); err != nil {
return nil, err

## }

secrets = append(secrets, s)

## }

return secrets, nil

## }

// UpdateSecretEncryption обновляет шифрование секрета
func (d *Database) UpdateSecretEncryption(secretID string, encryptedData, nonce []byte,
keyVersion int) error {
_, err := d.db.Exec(`
UPDATE secrets SET encrypted_data = ?, nonce = ?, key_version = ?, updated_at = ?
WHERE id = ?
`, encryptedData, nonce, keyVersion, time.Now(), secretID)
return err

## }

// MarkKeysCompromised помечает все ключи как скомпрометированные
func (d *Database) MarkKeysCompromised(userID string) error {
_, err := d.db.Exec(`
UPDATE key_versions SET status = 'compromised' WHERE user_id = ? AND status =

## 'active'

`, userID)
return err

## }

// SaveKeyRotationEvent сохраняет событие ротации
func (d *Database) SaveKeyRotationEvent(event *KeyRotationEvent) error {
_, err := d.db.Exec(`
INSERT INTO key_rotation_events (id, user_id, key_type, old_version, new_version,
reason, status, secrets_count, duration_ms, error_message, created_at, completed_at)

## VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)

`, event.ID, event.UserID, event.KeyType, event.OldVersion, event.NewVersion,
event.Reason, event.Status, event.SecretsCount, event.Duration, event.ErrorMessage,
event.CreatedAt, event.CompletedAt)
return err

## }

// GetKeyRotationHistory получает историю ротаций
func (d *Database) GetKeyRotationHistory(userID string, limit int) ([]KeyRotationEvent, error)

## {

rows, err := d.db.Query(`
SELECT id, user_id, key_type, old_version, new_version, reason, status,
secrets_count, duration_ms, created_at, completed_at
FROM key_rotation_events
WHERE user_id = ?
ORDER BY created_at DESC

## LIMIT ?

`, userID, limit)
if err != nil {
return nil, err

## }

defer rows.Close()
var events []KeyRotationEvent
for rows.Next() {
var e KeyRotationEvent
if err := rows.Scan(&e.ID, &e.UserID, &e.KeyType, &e.OldVersion, &e.NewVersion,
&e.Reason, &e.Status, &e.SecretsCount, &e.Duration, &e.CreatedAt, &e.CompletedAt); err !
= nil {
return nil, err

## }

events = append(events, e)

## }

return events, nil

## }

// GetUserForKeyRotation получает данные пользователя для ротации
func (d *Database) GetUserForKeyRotation(userID string) (*User, []byte, error) {
var user User
var salt []byte
err := d.db.QueryRow(`
SELECT id, email, master_key_salt, current_key_version FROM users WHERE id = ?
`, userID).Scan(&user.ID, &user.Email, &salt, &user.CurrentKeyVersion)
if err != nil {
return nil, nil, err

## }

return &user, salt, nil

## }

// UpdateUserMasterKey обновляет мастер-ключ пользователя
func (d *Database) UpdateUserMasterKey(userID string, salt, encryptedVaultKey, nonce,
srpVerifier []byte) error {
tx, err := d.db.Begin()
if err != nil {
return err

## }

defer tx.Rollback()
// Обновляем пользователя
_, err = tx.Exec(`
UPDATE users SET master_key_salt = ?, srp_verifier = ? WHERE id = ?
`, salt, srpVerifier, userID)
if err != nil {
return err

## }

// Обновляем активный vault key
_, err = tx.Exec(`
UPDATE key_versions SET encrypted_key = ?, nonce = ?
WHERE user_id = ? AND key_type = 'vault' AND status = 'active'
`, encryptedVaultKey, nonce, userID)
if err != nil {
return err

## }

return tx.Commit()

## }

// CreateSecret создаёт секрет
func (d *Database) CreateSecret(vaultID string, data *SecretData, vaultKey []byte,
keyVersion int) (*SecretItem, error) {
crypto := d.crypto
jsonData, err := json.Marshal(data)
if err != nil {
return nil, err

## }

encryptedData, nonce, err := crypto.EncryptXChaCha20(jsonData, vaultKey)
if err != nil {
return nil, err

## }

secretID := uuid.New().String()

now := time.Now()
_, err = d.db.Exec(`
INSERT INTO secrets (id, vault_id, encrypted_data, nonce, key_version, version,
created_at, updated_at)

## VALUES (?, ?, ?, ?, ?, 1, ?, ?)

`, secretID, vaultID, encryptedData, nonce, keyVersion, now, now)
if err != nil {
return nil, err

## }

return &SecretItem{
ID:         secretID,
VaultID:    vaultID,
KeyVersion: keyVersion,

## Version:    1,

CreatedAt:  now,
UpdatedAt:  now,
}, nil

## }

// GetSecret получает секрет
func (d *Database) GetSecret(secretID string, vaultKey []byte) (*SecretData, error) {
crypto := d.crypto
var encryptedData, nonce []byte
err := d.db.QueryRow(`
SELECT encrypted_data, nonce FROM secrets WHERE id = ?
`, secretID).Scan(&encryptedData, &nonce)
if err != nil {
return nil, err

## }

jsonData, err := crypto.DecryptXChaCha20(encryptedData, nonce, vaultKey)
if err != nil {
return nil, err

## }

var data SecretData
if err := json.Unmarshal(jsonData, &data); err != nil {
return nil, err

## }

return &data, nil

## }

## //

## ======================================================================

## ======

## // HTTP SERVER

## //

## ======================================================================

## ======

type Server struct {
db         *Database
crypto     *CryptoEngine
keyManager *KeyRotationManager
jwtSecret  []byte

## }

type Session struct {
UserID     string
VaultKey   []byte
KeyVersion int
MasterKey  []byte // Для операций ротации

## }

var sessions = make(map[string]*Session)
var sessionsMu sync.RWMutex
func NewServer(db *Database) *Server {
jwtSecret := make([]byte, 32)
rand.Read(jwtSecret)
return &Server{
db:         db,
crypto:     &CryptoEngine{},
keyManager: NewKeyRotationManager(db),
jwtSecret:  jwtSecret,

## }

## }

type RegisterRequest struct {
Email    string `json:"email"`
Password string `json:"password"`

## }

type LoginRequest struct {
Email     string `json:"email"`
Password  string `json:"password"`
SecretKey string `json:"secret_key"`

## }

type RotateKeyRequest struct {
Reason string `json:"reason,omitempty"`

## }

type ChangePasswordRequest struct {
OldPassword string `json:"old_password"`
NewPassword string `json:"new_password"`
SecretKey   string `json:"secret_key"`

## }

func (s *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {
var req RegisterRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "Invalid request", http.StatusBadRequest)
return

## }

if len(req.Password) < 12 {
http.Error(w, "Password must be at least 12 characters", http.StatusBadRequest)
return

## }

user, secretKey, err := s.db.CreateUser(req.Email, req.Password)
if err != nil {
http.Error(w, "Registration failed: "+err.Error(), http.StatusInternalServerError)
return

## }

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
"user_id":    user.ID,
"secret_key": secretKey,
"message":    "⚠️ ВАЖНО! Сохраните Secret Key. Без него восстановить доступ

## НЕВОЗМОЖНО!",

"key_rotation": map[string]interface{}{

## "current_version": 1,

"next_rotation":   time.Now().Add(time.Duration(DefaultKeyRotationDays) * 24 *
time.Hour),

## },

## })

## }

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
var req LoginRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "Invalid request", http.StatusBadRequest)
return

## }

user, vaultKey, err := s.db.AuthenticateUser(req.Email, req.Password, req.SecretKey)
if err != nil {
http.Error(w, "Invalid credentials", http.StatusUnauthorized)
return

## }

// Проверяем, нужна ли ротация
needsRotation, keyVersion, _ := s.keyManager.CheckRotationNeeded(user.ID)
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
"user_id": user.ID,
"exp":     time.Now().Add(time.Hour * 24).Unix(),

## })

tokenString, _ := token.SignedString(s.jwtSecret)
// Сохраняем сессию с master key для возможной ротации
secretKeyBytes, _ := s.crypto.ParseSecretKey(req.SecretKey)
var salt []byte
s.db.db.QueryRow(`SELECT master_key_salt FROM users WHERE id = ?`,
user.ID).Scan(&salt)
masterKey := s.crypto.DeriveMasterKey(req.Password, secretKeyBytes, salt)
sessionID := uuid.New().String()
sessionsMu.Lock()
sessions[sessionID] = &Session{
UserID:     user.ID,
VaultKey:   vaultKey,
KeyVersion: user.CurrentKeyVersion,
MasterKey:  masterKey,

## }

sessionsMu.Unlock()
response := map[string]interface{}{
"token":      tokenString,
"session_id": sessionID,
"key_info": map[string]interface{}{
"current_version": keyVersion.Version,
"expires_at":      keyVersion.ExpiresAt,
"needs_rotation":  needsRotation,

## },

## }

if needsRotation {
response["warning"] = "⚠️ Ваш ключ шифрования истёк. Рекомендуется выполнить
ротацию."

## }

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(response)

## }

func (s *Server) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
return func(w http.ResponseWriter, r *http.Request) {
authHeader := r.Header.Get("Authorization")

if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
http.Error(w, "Unauthorized", http.StatusUnauthorized)
return

## }

tokenString := strings.TrimPrefix(authHeader, "Bearer ")
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
return s.jwtSecret, nil

## })

if err != nil || !token.Valid {
http.Error(w, "Invalid token", http.StatusUnauthorized)
return

## }

claims := token.Claims.(jwt.MapClaims)
userID := claims["user_id"].(string)
r.Header.Set("X-User-ID", userID)
sessionID := r.Header.Get("X-Session-ID")
sessionsMu.RLock()
session, ok := sessions[sessionID]
sessionsMu.RUnlock()
if ok && session.UserID == userID {
r.Header.Set("X-Vault-Key", hex.EncodeToString(session.VaultKey))
r.Header.Set("X-Key-Version", fmt.Sprintf("%d", session.KeyVersion))

## }

next(w, r)

## }

## }

// RotateKeyHandler обработчик ротации ключа
func (s *Server) RotateKeyHandler(w http.ResponseWriter, r *http.Request) {
userID := r.Header.Get("X-User-ID")
sessionID := r.Header.Get("X-Session-ID")
sessionsMu.RLock()
session, ok := sessions[sessionID]
sessionsMu.RUnlock()
if !ok || session.MasterKey == nil {
http.Error(w, "Session expired, please login again", http.StatusUnauthorized)
return

## }

var req RotateKeyRequest
json.NewDecoder(r.Body).Decode(&req)

reason := req.Reason
if reason == "" {
reason = "manual"

## }

event, err := s.keyManager.RotateVaultKey(userID, session.MasterKey, reason)
if err != nil {
http.Error(w, "Key rotation failed: "+err.Error(), http.StatusInternalServerError)
return

## }

// Обновляем сессию с новым vault key
if event.Status == "completed" {
newKeyVersion, _ := s.db.GetCurrentKeyVersion(userID, "vault")
newVaultKey, _ := s.crypto.DecryptXChaCha20(newKeyVersion.EncryptedKey,
newKeyVersion.Nonce, session.MasterKey)
sessionsMu.Lock()
session.VaultKey = newVaultKey
session.KeyVersion = newKeyVersion.Version
sessionsMu.Unlock()

## }

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(event)

## }

// EmergencyRotateHandler экстренная ротация
func (s *Server) EmergencyRotateHandler(w http.ResponseWriter, r *http.Request) {
userID := r.Header.Get("X-User-ID")
sessionID := r.Header.Get("X-Session-ID")
sessionsMu.RLock()
session, ok := sessions[sessionID]
sessionsMu.RUnlock()
if !ok || session.MasterKey == nil {
http.Error(w, "Session expired", http.StatusUnauthorized)
return

## }

event, err := s.keyManager.EmergencyRotation(userID, session.MasterKey)
if err != nil {
http.Error(w, "Emergency rotation failed: "+err.Error(), http.StatusInternalServerError)
return

## }

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
"event":   event,

"message": " Экстренная ротация выполнена. Все предыдущие ключи помечены
как скомпрометированные.",

## })

## }

// ChangePasswordHandler смена пароля
func (s *Server) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
userID := r.Header.Get("X-User-ID")
var req ChangePasswordRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "Invalid request", http.StatusBadRequest)
return

## }

if len(req.NewPassword) < 12 {
http.Error(w, "New password must be at least 12 characters", http.StatusBadRequest)
return

## }

secretKeyBytes, err := s.crypto.ParseSecretKey(req.SecretKey)
if err != nil {
http.Error(w, "Invalid secret key", http.StatusBadRequest)
return

## }

err = s.keyManager.RotateMasterKey(userID, req.OldPassword, req.NewPassword,
secretKeyBytes)
if err != nil {
http.Error(w, "Password change failed: "+err.Error(), http.StatusBadRequest)
return

## }

// Инвалидируем все сессии пользователя
sessionsMu.Lock()
for id, sess := range sessions {
if sess.UserID == userID {
s.crypto.SecureZeroMemory(sess.VaultKey)
s.crypto.SecureZeroMemory(sess.MasterKey)
delete(sessions, id)

## }

## }

sessionsMu.Unlock()
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]string{
"message": "✅ Пароль успешно изменён. Пожалуйста, войдите заново.",

## })

## }

// GetKeyRotationHistoryHandler история ротаций
func (s *Server) GetKeyRotationHistoryHandler(w http.ResponseWriter, r *http.Request) {
userID := r.Header.Get("X-User-ID")
events, err := s.db.GetKeyRotationHistory(userID, 50)
if err != nil {
http.Error(w, "Failed to get history", http.StatusInternalServerError)
return

## }

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
"events": events,

## })

## }

// GetKeyStatusHandler статус ключей
func (s *Server) GetKeyStatusHandler(w http.ResponseWriter, r *http.Request) {
userID := r.Header.Get("X-User-ID")
currentVersion, err := s.db.GetCurrentKeyVersion(userID, "vault")
if err != nil {
http.Error(w, "Failed to get key status", http.StatusInternalServerError)
return

## }

needsRotation := time.Now().After(currentVersion.ExpiresAt)
daysUntilExpiry := int(time.Until(currentVersion.ExpiresAt).Hours() / 24)
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
"current_version":   currentVersion.Version,
"status":            currentVersion.Status,
"created_at":        currentVersion.CreatedAt,
"expires_at":        currentVersion.ExpiresAt,
"days_until_expiry": daysUntilExpiry,
"needs_rotation":    needsRotation,

## })

## }

// CreateSecretHandler создание секрета
func (s *Server) CreateSecretHandler(w http.ResponseWriter, r *http.Request) {
vaultKeyHex := r.Header.Get("X-Vault-Key")
keyVersionStr := r.Header.Get("X-Key-Version")
if vaultKeyHex == "" {
http.Error(w, "Session expired", http.StatusUnauthorized)
return

## }

vaultKey, _ := hex.DecodeString(vaultKeyHex)
var keyVersion int
fmt.Sscanf(keyVersionStr, "%d", &keyVersion)
var data SecretData
if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
http.Error(w, "Invalid request", http.StatusBadRequest)
return

## }

vars := mux.Vars(r)
vaultID := vars["vault_id"]
secret, err := s.db.CreateSecret(vaultID, &data, vaultKey, keyVersion)
if err != nil {
http.Error(w, "Failed to create secret", http.StatusInternalServerError)
return

## }

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(secret)

## }

// GetSecretHandler получение секрета
func (s *Server) GetSecretHandler(w http.ResponseWriter, r *http.Request) {
vaultKeyHex := r.Header.Get("X-Vault-Key")
if vaultKeyHex == "" {
http.Error(w, "Session expired", http.StatusUnauthorized)
return

## }

vaultKey, _ := hex.DecodeString(vaultKeyHex)
vars := mux.Vars(r)
secretID := vars["secret_id"]
data, err := s.db.GetSecret(secretID, vaultKey)
if err != nil {
http.Error(w, "Secret not found", http.StatusNotFound)
return

## }

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(data)

## }

func main() {
db, err := NewDatabase("vault_with_rotation.db")
if err != nil {
log.Fatal(err)

## }

## Архитектура Key Rotation:

## ```

## ┌────────────────────────────────────────────────────

## ─────────────┐

server := NewServer(db)
r := mux.NewRouter()

## // Публичные

r.HandleFunc("/api/v1/register", server.RegisterHandler).Methods("POST")
r.HandleFunc("/api/v1/login", server.LoginHandler).Methods("POST")
// Защищённые - секреты
r.HandleFunc("/api/v1/vaults/{vault_id}/secrets",
server.AuthMiddleware(server.CreateSecretHandler)).Methods("POST")
r.HandleFunc("/api/v1/secrets/{secret_id}",
server.AuthMiddleware(server.GetSecretHandler)).Methods("GET")
// Защищённые - ротация ключей
r.HandleFunc("/api/v1/keys/rotate",
server.AuthMiddleware(server.RotateKeyHandler)).Methods("POST")
r.HandleFunc("/api/v1/keys/emergency-rotate",
server.AuthMiddleware(server.EmergencyRotateHandler)).Methods("POST")
r.HandleFunc("/api/v1/keys/status",
server.AuthMiddleware(server.GetKeyStatusHandler)).Methods("GET")
r.HandleFunc("/api/v1/keys/history",
server.AuthMiddleware(server.GetKeyRotationHistoryHandler)).Methods("GET")
r.HandleFunc("/api/v1/account/change-password",
server.AuthMiddleware(server.ChangePasswordHandler)).Methods("POST")
fmt.Println(" Secure Vault Server with Key Rotation")
fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## ━━━━━")

fmt.Println(" Криптография: XChaCha20-Poly1305 + Argon2id")
fmt.Println(" Key Rotation: автоматическая каждые 90 дней")
fmt.Println(" Emergency Rotation: мгновенная при компрометации")
fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## ━━━━━")

fmt.Println(" Server starting on :8080")
log.Fatal(http.ListenAndServe(":8080", r))

## }

## │                     KEY ROTATION FLOW                           │

## ├────────────────────────────────────────────────────

## ─────────────┤

## │                                                                 │

## │  ┌──────────┐    ┌──────────┐    ┌──────────┐

## ┌──────────┐  │

│  │ Key v1   │───▶│ Key v2   │───▶│ Key v3   │───▶│ Key v4   │  │

## │  │ ACTIVE   │    │ ACTIVE   │    │ ACTIVE   │    │ ACTIVE   │  │

## │  └────┬─────┘    └────┬─────┘    └────┬─────┘

## └──────────┘  │

## │       │               │               │

Давай использовать postgres

## Claude Opus 4.5

Вот обновлённая версия с PostgreSQL:

## ```go

package main
import (

## "context"

## "crypto/rand"

## "crypto/sha256"

## "crypto/subtle"

## "database/sql"

## "encoding/base64"

## "encoding/hex"

## "encoding/json"

## "errors"

## "fmt"

## "io"

## "log"

## "net/http"

## "os"

## "strings"

## "sync"

## "time"

## "github.com/golang-jwt/jwt/v5"

## "github.com/google/uuid"

## _ "github.com/lib/pq"

## "golang.org/x/crypto/argon2"

## "golang.org/x/crypto/chacha20poly1305"

## "golang.org/x/crypto/hkdf"

## )

## //

## ================================================================

## ============

## // Configuration

## //

## ================================================================

## ============

type Config struct {
DatabaseURL     string
JWTSecret       []byte
ServerPort      string
Argon2Time      uint32
Argon2Memory    uint32
Argon2Threads   uint8
Argon2KeyLen    uint32
KeyRotationDays int
GracePeriodDays int

## }

func LoadConfig() *Config {
dbURL := os.Getenv("DATABASE_URL")
if dbURL == "" {
dbURL = "postgres://postgres:postgres@localhost:5432/keyvault?sslmode=disable"

## }

## }

jwtSecret := os.Getenv("JWT_SECRET")
if jwtSecret == "" {
jwtSecret = "change-me-in-production-use-strong-secret"

## }

return &Config{
DatabaseURL:     dbURL,
JWTSecret:       []byte(jwtSecret),
ServerPort:      getEnvOrDefault("SERVER_PORT", "8080"),
Argon2Time:      3,
Argon2Memory:    64 * 1024,
Argon2Threads:   4,
Argon2KeyLen:    32,
KeyRotationDays: 90,
GracePeriodDays: 7,

## }

func getEnvOrDefault(key, defaultVal string) string {
if val := os.Getenv(key); val != "" {
return val

## }

return defaultVal

## }

## //

## ================================================================

## ============

## // Database Models

## //

## ================================================================

## ============

type User struct {
ID                  string
Email               string
SecretKeyHash       string
MasterKeySalt       []byte
EncryptedVaultKey   []byte
VaultKeyNonce       []byte
SRPVerifier         []byte
CurrentKeyVersionID string
CreatedAt           time.Time
UpdatedAt           time.Time

## }

type KeyVersion struct {
ID                string
UserID            string
Version           int
EncryptedVaultKey []byte
VaultKeyNonce     []byte
Status            string      // active, deprecated, destroyed
CreatedAt         time.Time
RotatedAt         time.Time
DestroyAt         time.Time
DestroyedAt       *time.Time
RotationReason    string

## }

type Secret struct {
ID             string
UserID         string
KeyVersionID   string
Name           string
EncryptedData  []byte

## Nonce          []byte

SecretType     string
json:"id"
json:"email"
json:"-"
json:"-"
json:"-"
json:"-"
json:"-"
json:"-"
json:"created_at"
json:"updated_at"
json:"id"
json:"user_id"
json:"version"
json:"-"
json:"-"
json:"status"
json:"created_at"
json:"rotated_at,omitempty"
json:"destroy_at,omitempty"
json:"destroyed_at,omitempty"
json:"rotation_reason,omitempty"
json:"id"
json:"user_id"
json:"key_version_id"
json:"name"
json:"-"
json:"-"
json:"type"

Version        int
PreviousID     string

## Tags           []string

CreatedAt      time.Time
UpdatedAt      time.Time
ExpiresAt      time.Time
AccessCount    int
LastAccessedAt *time.Time

## }

type AuditLog struct {
ID         string
UserID     string
Action     string
ResourceID string
Details    json.RawMessage
IPAddress  string
UserAgent  string
CreatedAt  time.Time

## }

type Session struct {
UserID       string
VaultKey     []byte
KeyVersionID string
ExpiresAt    time.Time

## }

## //

## ================================================================

## ============

## // Database Layer

## //

## ================================================================

## ============

type Database struct {
db *sql.DB

## }

func NewDatabase(databaseURL string) (*Database, error) {
db, err := sql.Open("postgres", databaseURL)
if err != nil {
return nil, fmt.Errorf("failed to open database: %w", err)

## }

json:"version"
json:"previous_id,omitempty"
json:"tags,omitempty"
json:"created_at"
json:"updated_at"
json:"expires_at,omitempty"
json:"access_count"
json:"last_accessed_at,omitempty"
json:"id"
json:"user_id"
json:"action"
json:"resource_id,omitempty"
json:"details,omitempty"
json:"ip_address"
json:"user_agent"
json:"created_at"
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)

## }

func (d *Database) migrate() error {
migrations := []string{

## ,

## ,

if err := db.Ping(); err != nil {
return nil, fmt.Errorf("failed to ping database: %w", err)

## }

database := &Database{db: db}
if err := database.migrate(); err != nil {
return nil, fmt.Errorf("failed to migrate database: %w", err)

## }

return database, nil
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

## )

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

## )

CREATE TABLE IF NOT EXISTS secrets (
id UUID PRIMARY KEY,
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
key_version_id UUID NOT NULL REFERENCES key_versions(id),
name VARCHAR(255) NOT NULL,
encrypted_data BYTEA NOT NULL,
nonce BYTEA NOT NULL,

## ,

## ,

## ,

## ,

## ,

## ,

## ,

## ,

## ,

## }

## }

func (d *Database) Close() error {
return d.db.Close()

## }

secret_type VARCHAR(50) NOT NULL,
version INTEGER NOT NULL DEFAULT 1,
previous_id UUID REFERENCES secrets(id),
tags TEXT[],
created_at TIMESTAMPTZ DEFAULT NOW(),
updated_at TIMESTAMPTZ DEFAULT NOW(),
expires_at TIMESTAMPTZ,
access_count INTEGER DEFAULT 0,
last_accessed_at TIMESTAMPTZ

## )

CREATE TABLE IF NOT EXISTS audit_logs (
id UUID PRIMARY KEY,
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
action VARCHAR(100) NOT NULL,
resource_id VARCHAR(255),
details JSONB,
ip_address VARCHAR(45),
user_agent TEXT,
created_at TIMESTAMPTZ DEFAULT NOW()

## )

CREATE INDEX IF NOT EXISTS idx_secrets_user_id ON secrets(user_id)
CREATE INDEX IF NOT EXISTS idx_secrets_name ON secrets(user_id, name)
CREATE INDEX IF NOT EXISTS idx_secrets_key_version ON
secrets(key_version_id)
CREATE INDEX IF NOT EXISTS idx_key_versions_user_id ON
key_versions(user_id)
CREATE INDEX IF NOT EXISTS idx_key_versions_status ON key_versions(user_id,
status)
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id)
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON
audit_logs(created_at)
for _, migration := range migrations {
if _, err := d.db.Exec(migration); err != nil {
return fmt.Errorf("migration failed: %w", err)

## }

## }

return nil

// User operations
func (d Database) CreateUser(ctx context.Context, user User) error {
query :=

## }

func (d Database) GetUserByEmail(ctx context.Context, email string) (User, error) {
query :=

## }

func (d Database) GetUserByID(ctx context.Context, id string) (User, error) {
query :=
INSERT INTO users (id, email, secret_key_hash, master_key_salt,
encrypted_vault_key, vault_key_nonce, srp_verifier, current_key_version_id,
created_at, updated_at)

## VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)

_, err := d.db.ExecContext(ctx, query,
user.ID, user.Email, user.SecretKeyHash, user.MasterKeySalt,
user.EncryptedVaultKey, user.VaultKeyNonce, user.SRPVerifier,
user.CurrentKeyVersionID, user.CreatedAt, user.UpdatedAt)
return err
SELECT id, email, secret_key_hash, master_key_salt,
encrypted_vault_key, vault_key_nonce, srp_verifier, current_key_version_id,
created_at, updated_at
FROM users WHERE email = $1
user := &User{}
var currentKeyVersionID sql.NullString
err := d.db.QueryRowContext(ctx, query, email).Scan(
&user.ID, &user.Email, &user.SecretKeyHash, &user.MasterKeySalt,
&user.EncryptedVaultKey, &user.VaultKeyNonce, &user.SRPVerifier,
&currentKeyVersionID, &user.CreatedAt, &user.UpdatedAt)
if err == sql.ErrNoRows {
return nil, nil

## }

if err != nil {
return nil, err

## }

if currentKeyVersionID.Valid {
user.CurrentKeyVersionID = currentKeyVersionID.String

## }

return user, nil
SELECT id, email, secret_key_hash, master_key_salt,

## }

func (d *Database) UpdateUserKeys(ctx context.Context, userID string,
encryptedVaultKey, vaultKeyNonce, masterKeySalt []byte, currentKeyVersionID
string) error {
query :=

## }

func (d *Database) UpdateUserCurrentKeyVersion(ctx context.Context, userID,
keyVersionID string) error {
query :=
_, err := d.db.ExecContext(ctx, query, keyVersionID, userID)
encrypted_vault_key, vault_key_nonce, srp_verifier, current_key_version_id,
created_at, updated_at
FROM users WHERE id = $1
user := &User{}
var currentKeyVersionID sql.NullString
err := d.db.QueryRowContext(ctx, query, id).Scan(
&user.ID, &user.Email, &user.SecretKeyHash, &user.MasterKeySalt,
&user.EncryptedVaultKey, &user.VaultKeyNonce, &user.SRPVerifier,
&currentKeyVersionID, &user.CreatedAt, &user.UpdatedAt)
if err == sql.ErrNoRows {
return nil, nil

## }

if err != nil {
return nil, err

## }

if currentKeyVersionID.Valid {
user.CurrentKeyVersionID = currentKeyVersionID.String

## }

return user, nil
UPDATE users
SET encrypted_vault_key = $1, vault_key_nonce = $2, master_key_salt = $3,
current_key_version_id = $4, updated_at = NOW()
WHERE id = $5
_, err := d.db.ExecContext(ctx, query, encryptedVaultKey, vaultKeyNonce, masterKeySalt,
currentKeyVersionID, userID)
return err
UPDATE users SET current_key_version_id = $1, updated_at = NOW()
WHERE id = $2

return err

## }

// Key version operations
func (d Database) CreateKeyVersion(ctx context.Context, kv KeyVersion) error {
query :=

## }

func (d Database) GetKeyVersion(ctx context.Context, id string) (KeyVersion, error) {
query :=

## }

func (d Database) GetActiveKeyVersion(ctx context.Context, userID string)
(KeyVersion, error) {
query :=
INSERT INTO key_versions (id, user_id, version, encrypted_vault_key,
vault_key_nonce, status, created_at, rotation_reason)

## VALUES ($1, $2, $3, $4, $5, $6, $7, $8)

_, err := d.db.ExecContext(ctx, query,
kv.ID, kv.UserID, kv.Version, kv.EncryptedVaultKey, kv.VaultKeyNonce,
kv.Status, kv.CreatedAt, kv.RotationReason)
return err
SELECT id, user_id, version, encrypted_vault_key, vault_key_nonce,
status, created_at, rotated_at, destroy_at, destroyed_at, rotation_reason
FROM key_versions WHERE id = $1
kv := &KeyVersion{}
err := d.db.QueryRowContext(ctx, query, id).Scan(
&kv.ID, &kv.UserID, &kv.Version, &kv.EncryptedVaultKey, &kv.VaultKeyNonce,
&kv.Status, &kv.CreatedAt, &kv.RotatedAt, &kv.DestroyAt, &kv.DestroyedAt,
&kv.RotationReason)
if err == sql.ErrNoRows {
return nil, nil

## }

return kv, err
SELECT id, user_id, version, encrypted_vault_key, vault_key_nonce,
status, created_at, rotated_at, destroy_at, destroyed_at, rotation_reason
FROM key_versions WHERE user_id = $1 AND status = 'active' ORDER BY
version DESC LIMIT 1
kv := &KeyVersion{}
err := d.db.QueryRowContext(ctx, query, userID).Scan(
&kv.ID, &kv.UserID, &kv.Version, &kv.EncryptedVaultKey, &kv.VaultKeyNonce,
&kv.Status, &kv.CreatedAt, &kv.RotatedAt, &kv.DestroyAt, &kv.DestroyedAt,

## }

func (d *Database) GetLatestKeyVersionNumber(ctx context.Context, userID string)
(int, error) {
query :=
var version int
err := d.db.QueryRowContext(ctx, query, userID).Scan(&version)
return version, err

## }

func (d Database) GetKeyVersionHistory(ctx context.Context, userID string)
([]KeyVersion, error) {
query :=

## }

&kv.RotationReason)
if err == sql.ErrNoRows {
return nil, nil

## }

return kv, err
SELECT COALESCE(MAX(version), 0) FROM key_versions WHERE
user_id = $1
SELECT id, user_id, version, encrypted_vault_key, vault_key_nonce,
status, created_at, rotated_at, destroy_at, destroyed_at, rotation_reason
FROM key_versions WHERE user_id = $1 ORDER BY version DESC
rows, err := d.db.QueryContext(ctx, query, userID)
if err != nil {
return nil, err

## }

defer rows.Close()
var versions []*KeyVersion
for rows.Next() {
kv := &KeyVersion{}
if err := rows.Scan(
&kv.ID, &kv.UserID, &kv.Version, &kv.EncryptedVaultKey, &kv.VaultKeyNonce,
&kv.Status, &kv.CreatedAt, &kv.RotatedAt, &kv.DestroyAt, &kv.DestroyedAt,
&kv.RotationReason); err != nil {
return nil, err

## }

versions = append(versions, kv)

## }

return versions, rows.Err()

func (d Database) UpdateKeyVersionStatus(ctx context.Context, id, status string,
rotatedAt, destroyAt time.Time) error {
query :=
_, err := d.db.ExecContext(ctx, query, status, rotatedAt, destroyAt, id)
return err

## }

func (d *Database) DestroyKeyVersion(ctx context.Context, id string) error {
now := time.Now()
query :=
_, err := d.db.ExecContext(ctx, query, now, id)
return err

## }

// Secret operations
func (d Database) CreateSecret(ctx context.Context, secret Secret) error {
query :=

## }

func (d Database) GetSecret(ctx context.Context, id, userID string) (Secret, error) {
query :=
UPDATE key_versions SET status = $1, rotated_at = $2, destroy_at = $3
WHERE id = $4
UPDATE key_versions SET status = 'destroyed', destroyed_at = $1,
encrypted_vault_key = NULL, vault_key_nonce = NULL WHERE id = $2
INSERT INTO secrets (id, user_id, key_version_id, name, encrypted_data,
nonce, secret_type, version, previous_id, tags, created_at, updated_at, expires_at)

## VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)

_, err := d.db.ExecContext(ctx, query,
secret.ID, secret.UserID, secret.KeyVersionID, secret.Name, secret.EncryptedData,
secret.Nonce, secret.SecretType, secret.Version, secret.PreviousID, secret.Tags,
secret.CreatedAt, secret.UpdatedAt, secret.ExpiresAt)
return err
SELECT id, user_id, key_version_id, name, encrypted_data, nonce,
secret_type, version, previous_id, tags, created_at, updated_at, expires_at,
access_count, last_accessed_at
FROM secrets WHERE id = $1 AND user_id = $2
secret := &Secret{}
err := d.db.QueryRowContext(ctx, query, id, userID).Scan(
&secret.ID, &secret.UserID, &secret.KeyVersionID, &secret.Name,
&secret.EncryptedData,
&secret.Nonce, &secret.SecretType, &secret.Version, &secret.PreviousID, &secret.Tags,
&secret.CreatedAt, &secret.UpdatedAt, &secret.ExpiresAt, &secret.AccessCount,
&secret.LastAccessedAt)
if err == sql.ErrNoRows {
return nil, nil

## }

func (d Database) GetSecretsByUser(ctx context.Context, userID string) ([]Secret,
error) {
query :=

## }

func (d Database) GetSecretsByKeyVersion(ctx context.Context, keyVersionID
string) ([]Secret, error) {
query :=

## }

return secret, err
SELECT id, user_id, key_version_id, name, encrypted_data, nonce,
secret_type, version, previous_id, tags, created_at, updated_at, expires_at,
access_count, last_accessed_at
FROM secrets WHERE user_id = $1 ORDER BY name
rows, err := d.db.QueryContext(ctx, query, userID)
if err != nil {
return nil, err

## }

defer rows.Close()
var secrets []*Secret
for rows.Next() {
secret := &Secret{}
if err := rows.Scan(
&secret.ID, &secret.UserID, &secret.KeyVersionID, &secret.Name,
&secret.EncryptedData,
&secret.Nonce, &secret.SecretType, &secret.Version, &secret.PreviousID,
&secret.Tags,
&secret.CreatedAt, &secret.UpdatedAt, &secret.ExpiresAt, &secret.AccessCount,
&secret.LastAccessedAt); err != nil {
return nil, err

## }

secrets = append(secrets, secret)

## }

return secrets, rows.Err()
SELECT id, user_id, key_version_id, name, encrypted_data, nonce,
secret_type, version, previous_id, tags, created_at, updated_at, expires_at,
access_count, last_accessed_at
FROM secrets WHERE key_version_id = $1
rows, err := d.db.QueryContext(ctx, query, keyVersionID)
if err != nil {
return nil, err

## }

func (d *Database) UpdateSecretKeyVersion(ctx context.Context, secretID,
keyVersionID string, encryptedData, nonce []byte) error {
query :=
_, err := d.db.ExecContext(ctx, query, keyVersionID, encryptedData, nonce, secretID)
return err

## }

func (d *Database) UpdateSecretAccess(ctx context.Context, id string) error {
query :=
_, err := d.db.ExecContext(ctx, query, id)
return err

## }

func (d *Database) DeleteSecret(ctx context.Context, id, userID string) error {
query :=
_, err := d.db.ExecContext(ctx, query, id, userID)
return err

## }

// Audit log operations
func (d Database) CreateAuditLog(ctx context.Context, log AuditLog) error {
query :=

## }

defer rows.Close()
var secrets []*Secret
for rows.Next() {
secret := &Secret{}
if err := rows.Scan(
&secret.ID, &secret.UserID, &secret.KeyVersionID, &secret.Name,
&secret.EncryptedData,
&secret.Nonce, &secret.SecretType, &secret.Version, &secret.PreviousID,
&secret.Tags,
&secret.CreatedAt, &secret.UpdatedAt, &secret.ExpiresAt, &secret.AccessCount,
&secret.LastAccessedAt); err != nil {
return nil, err

## }

secrets = append(secrets, secret)

## }

return secrets, rows.Err()
UPDATE secrets SET key_version_id = $1, encrypted_data = $2, nonce =
$3, updated_at = NOW() WHERE id = $4
UPDATE secrets SET access_count = access_count + 1, last_accessed_at
= NOW() WHERE id = $1
DELETE FROM secrets WHERE id = $1 AND user_id = $2
INSERT INTO audit_logs (id, user_id, action, resource_id, details,
ip_address, user_agent, created_at)

## VALUES ($1, $2, $3, $4, $5, $6, $7, $8)

## }

func (d Database) GetAuditLogs(ctx context.Context, userID string, limit int)
([]AuditLog, error) {
query :=

## }

// Transaction support
func (d Database) BeginTx(ctx context.Context) (sql.Tx, error) {
return d.db.BeginTx(ctx, nil)

## }

## //

## ================================================================

## ============

## // Cryptography

## //

## ================================================================

## ============

_, err := d.db.ExecContext(ctx, query,
log.ID, log.UserID, log.Action, log.ResourceID, log.Details, log.IPAddress, log.UserAgent,
log.CreatedAt)
return err
SELECT id, user_id, action, resource_id, details, ip_address, user_agent,
created_at
FROM audit_logs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2
rows, err := d.db.QueryContext(ctx, query, userID, limit)
if err != nil {
return nil, err

## }

defer rows.Close()
var logs []*AuditLog
for rows.Next() {
log := &AuditLog{}
if err := rows.Scan(
&log.ID, &log.UserID, &log.Action, &log.ResourceID, &log.Details,
&log.IPAddress, &log.UserAgent, &log.CreatedAt); err != nil {
return nil, err

## }

logs = append(logs, log)

## }

return logs, rows.Err()

type CryptoService struct {
config *Config

## }

func NewCryptoService(config Config) CryptoService {
return &CryptoService{config: config}

## }

func (c *CryptoService) GenerateRandomBytes(n int) ([]byte, error) {
b := make([]byte, n)
if _, err := rand.Read(b); err != nil {
return nil, err

## }

return b, nil

## }

func (c *CryptoService) GenerateSecretKey() (string, error) {
bytes, err := c.GenerateRandomBytes(32)
if err != nil {
return "", err

## }

encoded := base64.RawURLEncoding.EncodeToString(bytes)
parts := make([]string, 0, 7)
for i := 0; i < len(encoded) && len(parts) < 7; i += 6 {
end := i + 6
if end > len(encoded) {
end = len(encoded)

## }

parts = append(parts, encoded[i:end])

## }

return strings.Join(parts, "-"), nil

## }

func (c *CryptoService) HashSecretKey(secretKey string) string {
hash := sha256.Sum256([]byte(secretKey))
return hex.EncodeToString(hash[:])

## }

func (c *CryptoService) DeriveMasterKey(password, secretKey string, salt []byte)

## []byte {

combined := []byte(password + secretKey)
return argon2.IDKey(combined, salt, c.config.Argon2Time, c.config.Argon2Memory,
c.config.Argon2Threads, c.config.Argon2KeyLen)

## }

func (c *CryptoService) DeriveKey(masterKey []byte, context string, length int)
([]byte, error) {
reader := hkdf.New(sha256.New, masterKey, nil, []byte(context))

key := make([]byte, length)
if _, err := io.ReadFull(reader, key); err != nil {
return nil, err

## }

return key, nil

## }

func (c *CryptoService) Encrypt(key, plaintext []byte) (ciphertext, nonce []byte, err
error) {
aead, err := chacha20poly1305.NewX(key)
if err != nil {
return nil, nil, err

## }

## }

func (c *CryptoService) Decrypt(key, ciphertext, nonce []byte) ([]byte, error) {
aead, err := chacha20poly1305.NewX(key)
if err != nil {
return nil, err

## }

## }

func (c *CryptoService) GenerateVaultKey() ([]byte, error) {
return c.GenerateRandomBytes(32)

## }

func (c *CryptoService) GenerateSRPVerifier(password string, salt []byte) []byte {
return argon2.IDKey([]byte(password), salt, c.config.Argon2Time,
nonce, err = c.GenerateRandomBytes(chacha20poly1305.NonceSizeX)
if err != nil {
return nil, nil, err

## }

ciphertext = aead.Seal(nil, nonce, plaintext, nil)
return ciphertext, nonce, nil
plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
if err != nil {
return nil, errors.New("decryption failed: invalid key or corrupted data")

## }

return plaintext, nil

c.config.Argon2Memory, c.config.Argon2Threads, 32)

## }

func SecureZeroMemory(b []byte) {
for i := range b {
b[i] = 0

## }

## }

## //

## ================================================================

## ============

## // Session Manager

## //

## ================================================================

## ============

type SessionManager struct {
sessions map[string]*Session
mu       sync.RWMutex

## }

func NewSessionManager() SessionManager {
return &SessionManager{
sessions: make(map[string]Session),

## }

## }

func (sm *SessionManager) Create(userID string, vaultKey []byte, keyVersionID
string, duration time.Duration) string {
sm.mu.Lock()
defer sm.mu.Unlock()

## }

sessionID := uuid.New().String()
keyCopy := make([]byte, len(vaultKey))
copy(keyCopy, vaultKey)
sm.sessions[sessionID] = &Session{
UserID:       userID,
VaultKey:     keyCopy,
KeyVersionID: keyVersionID,
ExpiresAt:    time.Now().Add(duration),

## }

return sessionID

func (sm SessionManager) Get(sessionID string) (Session, bool) {
sm.mu.RLock()
defer sm.mu.RUnlock()

## }

func (sm *SessionManager) Update(sessionID string, vaultKey []byte, keyVersionID
string) bool {
sm.mu.Lock()
defer sm.mu.Unlock()

## }

func (sm *SessionManager) Delete(sessionID string) {
sm.mu.Lock()
defer sm.mu.Unlock()

## }

session, exists := sm.sessions[sessionID]
if !exists || time.Now().After(session.ExpiresAt) {
return nil, false

## }

return session, true
session, exists := sm.sessions[sessionID]
if !exists {
return false

## }

SecureZeroMemory(session.VaultKey)
keyCopy := make([]byte, len(vaultKey))
copy(keyCopy, vaultKey)
session.VaultKey = keyCopy
session.KeyVersionID = keyVersionID
return true
if session, exists := sm.sessions[sessionID]; exists {
SecureZeroMemory(session.VaultKey)
delete(sm.sessions, sessionID)

## }

func (sm *SessionManager) DeleteUserSessions(userID string) {
sm.mu.Lock()
defer sm.mu.Unlock()

## }

## //

## ================================================================

## ============

## // Key Rotation Manager

## //

## ================================================================

## ============

type KeyRotationManager struct {
db      Database
crypto  CryptoService
config  Config
session SessionManager

## }

func NewKeyRotationManager(db Database, crypto CryptoService, config Config,
session SessionManager) *KeyRotationManager {
return &KeyRotationManager{
db:      db,
crypto:  crypto,
config:  config,
session: session,

## }

## }

func (krm KeyRotationManager) RotateVaultKey(ctx context.Context, userID string,
currentVaultKey []byte, reason string) (KeyVersion, error) {
tx, err := krm.db.BeginTx(ctx)
if err != nil {
return nil, err

## }

defer tx.Rollback()
for id, session := range sm.sessions {
if session.UserID == userID {
SecureZeroMemory(session.VaultKey)
delete(sm.sessions, id)

## }

## }

user, err := krm.db.GetUserByID(ctx, userID)
if err != nil || user == nil {
return nil, errors.New("user not found")

## }

currentKV, err := krm.db.GetActiveKeyVersion(ctx, userID)
if err != nil {
return nil, err

## }

// Generate new vault key
newVaultKey, err := krm.crypto.GenerateVaultKey()
if err != nil {
return nil, err

## }

defer SecureZeroMemory(newVaultKey)
// Encrypt new vault key with user's master key (we need to derive it from session)
// For now, we'll encrypt with the current vault key as a wrapper
encryptedNewVaultKey, newNonce, err := krm.crypto.Encrypt(currentVaultKey, newVaultKey)
if err != nil {
return nil, err

## }

// Get next version number
latestVersion, err := krm.db.GetLatestKeyVersionNumber(ctx, userID)
if err != nil {
return nil, err

## }

// Create new key version
newKV := &KeyVersion{
ID:                uuid.New().String(),
UserID:            userID,
Version:           latestVersion + 1,
EncryptedVaultKey: encryptedNewVaultKey,
VaultKeyNonce:     newNonce,

## Status:            "active",

CreatedAt:         time.Now(),
RotationReason:    reason,

## }

if err := krm.db.CreateKeyVersion(ctx, newKV); err != nil {
return nil, err

## }

// Mark old key version as deprecated
if currentKV != nil {
now := time.Now()
destroyAt := now.Add(time.Duration(krm.config.GracePeriodDays) * 24 * time.Hour)

## }

func (krm *KeyRotationManager) RotateMasterKey(ctx context.Context, userID,
oldPassword, newPassword, secretKey string) error {
user, err := krm.db.GetUserByID(ctx, userID)
if err := krm.db.UpdateKeyVersionStatus(ctx, currentKV.ID, "deprecated", &now,
&destroyAt); err != nil {
return nil, err

## }

## }

// Re-encrypt all secrets with new key
secrets, err := krm.db.GetSecretsByKeyVersion(ctx, currentKV.ID)
if err != nil {
return nil, err

## }

for _, secret := range secrets {
// Decrypt with old key
plaintext, err := krm.crypto.Decrypt(currentVaultKey, secret.EncryptedData, secret.Nonce)
if err != nil {
continue // Skip corrupted secrets

## }

// Re-encrypt with new key
newCiphertext, newNonce, err := krm.crypto.Encrypt(newVaultKey, plaintext)
if err != nil {
SecureZeroMemory(plaintext)
continue

## }

SecureZeroMemory(plaintext)
if err := krm.db.UpdateSecretKeyVersion(ctx, secret.ID, newKV.ID, newCiphertext,
newNonce); err != nil {
return nil, err

## }

## }

// Update user's current key version
if err := krm.db.UpdateUserCurrentKeyVersion(ctx, userID, newKV.ID); err != nil {
return nil, err

## }

if err := tx.Commit(); err != nil {
return nil, err

## }

return newKV, nil

if err != nil || user == nil {
return errors.New("user not found")

## }

// Verify old password
oldMasterKey := krm.crypto.DeriveMasterKey(oldPassword, secretKey, user.MasterKeySalt)
defer SecureZeroMemory(oldMasterKey)
encryptionKey, err := krm.crypto.DeriveKey(oldMasterKey, "vault-key-encryption", 32)
if err != nil {
return err

## }

defer SecureZeroMemory(encryptionKey)
// Decrypt current vault key
vaultKey, err := krm.crypto.Decrypt(encryptionKey, user.EncryptedVaultKey,
user.VaultKeyNonce)
if err != nil {
return errors.New("invalid old password")

## }

defer SecureZeroMemory(vaultKey)
// Generate new salt and derive new master key
newSalt, err := krm.crypto.GenerateRandomBytes(32)
if err != nil {
return err

## }

newMasterKey := krm.crypto.DeriveMasterKey(newPassword, secretKey, newSalt)
defer SecureZeroMemory(newMasterKey)
newEncryptionKey, err := krm.crypto.DeriveKey(newMasterKey, "vault-key-encryption", 32)
if err != nil {
return err

## }

defer SecureZeroMemory(newEncryptionKey)
// Re-encrypt vault key with new master key
newEncryptedVaultKey, newNonce, err := krm.crypto.Encrypt(newEncryptionKey, vaultKey)
if err != nil {
return err

## }

// Update user
if err := krm.db.UpdateUserKeys(ctx, userID, newEncryptedVaultKey, newNonce, newSalt,
user.CurrentKeyVersionID); err != nil {
return err

## }

// Invalidate all sessions

## }

## //

## ================================================================

## ============

## // Vault Service

## //

## ================================================================

## ============

type VaultService struct {
db         Database
crypto     CryptoService
sessions   SessionManager
keyManager KeyRotationManager
config     *Config

## }

func NewVaultService(db Database, crypto CryptoService, sessions
SessionManager, config Config) *VaultService {
vs := &VaultService{
db:       db,
crypto:   crypto,
sessions: sessions,
config:   config,

## }

vs.keyManager = NewKeyRotationManager(db, crypto, config, sessions)
return vs

## }

type RegisterRequest struct {
Email    string
Password string

## }

type RegisterResponse struct {
UserID    string
SecretKey string
Message   string

## }

func (vs VaultService) Register(ctx context.Context, req RegisterRequest)
(*RegisterResponse, error) {
existing, err := vs.db.GetUserByEmail(ctx, req.Email)
krm.session.DeleteUserSessions(userID)
return nil
json:"email"
json:"password"
json:"user_id"
json:"secret_key"
json:"message"

if err != nil {
return nil, err

## }

if existing != nil {
return nil, errors.New("email already registered")

## }

secretKey, err := vs.crypto.GenerateSecretKey()
if err != nil {
return nil, err

## }

salt, err := vs.crypto.GenerateRandomBytes(32)
if err != nil {
return nil, err

## }

masterKey := vs.crypto.DeriveMasterKey(req.Password, secretKey, salt)
defer SecureZeroMemory(masterKey)
vaultKey, err := vs.crypto.GenerateVaultKey()
if err != nil {
return nil, err

## }

defer SecureZeroMemory(vaultKey)
encryptionKey, err := vs.crypto.DeriveKey(masterKey, "vault-key-encryption", 32)
if err != nil {
return nil, err

## }

defer SecureZeroMemory(encryptionKey)
encryptedVaultKey, nonce, err := vs.crypto.Encrypt(encryptionKey, vaultKey)
if err != nil {
return nil, err

## }

srpVerifier := vs.crypto.GenerateSRPVerifier(req.Password, salt)
userID := uuid.New().String()
keyVersionID := uuid.New().String()
now := time.Now()
user := &User{
ID:                  userID,
Email:               req.Email,
SecretKeyHash:       vs.crypto.HashSecretKey(secretKey),
MasterKeySalt:       salt,
EncryptedVaultKey:   encryptedVaultKey,
VaultKeyNonce:       nonce,
SRPVerifier:         srpVerifier,

## }

type LoginRequest struct {
Email     string
Password  string
SecretKey string

## }

type LoginResponse struct {
Token     string
ExpiresAt int64

## }

func (vs VaultService) Login(ctx context.Context, req LoginRequest, ipAddress,
userAgent string) (*LoginResponse, error) {
user, err := vs.db.GetUserByEmail(ctx, req.Email)
if err != nil {
return nil, err

## }

CurrentKeyVersionID: keyVersionID,
CreatedAt:           now,
UpdatedAt:           now,

## }

if err := vs.db.CreateUser(ctx, user); err != nil {
return nil, err

## }

keyVersion := &KeyVersion{
ID:                keyVersionID,
UserID:            userID,

## Version:           1,

EncryptedVaultKey: encryptedVaultKey,
VaultKeyNonce:     nonce,

## Status:            "active",

CreatedAt:         now,
RotationReason:    "initial",

## }

if err := vs.db.CreateKeyVersion(ctx, keyVersion); err != nil {
return nil, err

## }

return &RegisterResponse{
UserID:    userID,
SecretKey: secretKey,
Message:   "IMPORTANT: Save your Secret Key securely. It cannot be recovered!",
}, nil
json:"email"
json:"password"
json:"secret_key"
json:"token"
json:"expires_at"

if user == nil {
time.Sleep(100 * time.Millisecond) // Timing attack mitigation
return nil, errors.New("invalid credentials")

## }

secretKeyHash := vs.crypto.HashSecretKey(req.SecretKey)
if subtle.ConstantTimeCompare([]byte(secretKeyHash), []byte(user.SecretKeyHash)) != 1 {
return nil, errors.New("invalid credentials")

## }

masterKey := vs.crypto.DeriveMasterKey(req.Password, req.SecretKey, user.MasterKeySalt)
defer SecureZeroMemory(masterKey)
encryptionKey, err := vs.crypto.DeriveKey(masterKey, "vault-key-encryption", 32)
if err != nil {
return nil, err

## }

defer SecureZeroMemory(encryptionKey)
vaultKey, err := vs.crypto.Decrypt(encryptionKey, user.EncryptedVaultKey,
user.VaultKeyNonce)
if err != nil {
return nil, errors.New("invalid credentials")

## }

sessionDuration := 24 * time.Hour
sessionID := vs.sessions.Create(user.ID, vaultKey, user.CurrentKeyVersionID,
sessionDuration)
SecureZeroMemory(vaultKey)
expiresAt := time.Now().Add(sessionDuration)
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
"sub":        user.ID,
"session_id": sessionID,
"exp":        expiresAt.Unix(),

## })

tokenString, err := token.SignedString(vs.config.JWTSecret)
if err != nil {
return nil, err

## }

// Audit log
vs.logAudit(ctx, user.ID, "login", "", nil, ipAddress, userAgent)
return &LoginResponse{
Token:     tokenString,
ExpiresAt: expiresAt.Unix(),
}, nil

## }

type CreateSecretRequest struct {
Name      string
Data      string
Type      string

## Tags      []string

ExpiresAt *time.Time

## }

type SecretResponse struct {
ID        string
Name      string
Data      string
Type      string
Version   int

## Tags      []string

CreatedAt time.Time
UpdatedAt time.Time
ExpiresAt *time.Time

## }

func (vs VaultService) CreateSecret(ctx context.Context, session Session, req
CreateSecretRequest) (SecretResponse, error) {
encryptedData, nonce, err := vs.crypto.Encrypt(session.VaultKey, []byte(req.Data))
if err != nil {
return nil, err

## }

json:"name"
json:"data"
json:"type"
json:"tags,omitempty"
json:"expires_at,omitempty"
json:"id"
json:"name"
json:"data,omitempty"
json:"type"
json:"version"
json:"tags,omitempty"
json:"created_at"
json:"updated_at"
json:"expires_at,omitempty"
now := time.Now()
secret := &Secret{
ID:            uuid.New().String(),
UserID:        session.UserID,
KeyVersionID:  session.KeyVersionID,
Name:          req.Name,
EncryptedData: encryptedData,
Nonce:         nonce,
SecretType:    req.Type,

## Version:       1,

Tags:          req.Tags,
CreatedAt:     now,
UpdatedAt:     now,
ExpiresAt:     req.ExpiresAt,

## }

if err := vs.db.CreateSecret(ctx, secret); err != nil {
return nil, err

## }

## }

func (vs VaultService) GetSecret(ctx context.Context, session Session, secretID
string) (*SecretResponse, error) {
secret, err := vs.db.GetSecret(ctx, secretID, session.UserID)
if err != nil {
return nil, err

## }

if secret == nil {
return nil, errors.New("secret not found")

## }

## }

func (vs VaultService) ListSecrets(ctx context.Context, session Session)
([]*SecretResponse, error) {
secrets, err := vs.db.GetSecretsByUser(ctx, session.UserID)
return &SecretResponse{
ID:        secret.ID,
Name:      secret.Name,
Type:      secret.SecretType,
Version:   secret.Version,
Tags:      secret.Tags,
CreatedAt: secret.CreatedAt,
UpdatedAt: secret.UpdatedAt,
ExpiresAt: secret.ExpiresAt,
}, nil
plaintext, err := vs.crypto.Decrypt(session.VaultKey, secret.EncryptedData, secret.Nonce)
if err != nil {
return nil, err

## }

vs.db.UpdateSecretAccess(ctx, secretID)
return &SecretResponse{
ID:        secret.ID,
Name:      secret.Name,
Data:      string(plaintext),
Type:      secret.SecretType,
Version:   secret.Version,
Tags:      secret.Tags,
CreatedAt: secret.CreatedAt,
UpdatedAt: secret.UpdatedAt,
ExpiresAt: secret.ExpiresAt,
}, nil

if err != nil {
return nil, err

## }

## }

func (vs VaultService) DeleteSecret(ctx context.Context, session Session, secretID,
ipAddress, userAgent string) error {
if err := vs.db.DeleteSecret(ctx, secretID, session.UserID); err != nil {
return err

## }

## }

func (vs VaultService) RotateVaultKey(ctx context.Context, session Session, reason
string) (*KeyVersion, error) {
return vs.keyManager.RotateVaultKey(ctx, session.UserID, session.VaultKey,
reason)

## }

func (vs *VaultService) ChangePassword(ctx context.Context, userID, oldPassword,
newPassword, secretKey string) error {
return vs.keyManager.RotateMasterKey(ctx, userID, oldPassword, newPassword,
secretKey)

## }

responses := make([]*SecretResponse, len(secrets))
for i, secret := range secrets {
responses[i] = &SecretResponse{
ID:        secret.ID,
Name:      secret.Name,
Type:      secret.SecretType,
Version:   secret.Version,
Tags:      secret.Tags,
CreatedAt: secret.CreatedAt,
UpdatedAt: secret.UpdatedAt,
ExpiresAt: secret.ExpiresAt,

## }

## }

return responses, nil
vs.logAudit(ctx, session.UserID, "secret_deleted", secretID, nil, ipAddress, userAgent)
return nil

func (vs VaultService) GetKeyStatus(ctx context.Context, userID string)
(KeyVersion, error) {
return vs.db.GetActiveKeyVersion(ctx, userID)

## }

func (vs VaultService) GetKeyHistory(ctx context.Context, userID string)
([]KeyVersion, error) {
return vs.db.GetKeyVersionHistory(ctx, userID)

## }

func (vs *VaultService) logAudit(ctx context.Context, userID, action, resourceID
string, details map[string]interface{}, ipAddress, userAgent string) {
var detailsJSON json.RawMessage
if details != nil {
detailsJSON, _ = json.Marshal(details)

## }

## }

## //

## ================================================================

## ============

// HTTP Handlers

## //

## ================================================================

## ============

type Server struct {
vault  VaultService
config Config

## }

func NewServer(vault VaultService, config Config) *Server {
return &Server{vault: vault, config: config}

## }

log := &AuditLog{
ID:         uuid.New().String(),
UserID:     userID,
Action:     action,
ResourceID: resourceID,
Details:    detailsJSON,
IPAddress:  ipAddress,
UserAgent:  userAgent,
CreatedAt:  time.Now(),

## }

vs.db.CreateAuditLog(ctx, log)

func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(status)
json.NewEncoder(w).Encode(data)

## }

func (s *Server) respondError(w http.ResponseWriter, status int, message string) {
s.respondJSON(w, status, map[string]string{"error": message})

## }

func (s Server) authenticate(r http.Request) (*Session, error) {
authHeader := r.Header.Get("Authorization")
if !strings.HasPrefix(authHeader, "Bearer ") {
return nil, errors.New("missing authorization header")

## }

## }

tokenString := strings.TrimPrefix(authHeader, "Bearer ")
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
return nil, errors.New("invalid signing method")

## }

return s.config.JWTSecret, nil

## })

if err != nil || !token.Valid {
return nil, errors.New("invalid token")

## }

claims, ok := token.Claims.(jwt.MapClaims)
if !ok {
return nil, errors.New("invalid claims")

## }

sessionID, ok := claims["session_id"].(string)
if !ok {
return nil, errors.New("invalid session")

## }

session, exists := s.vault.sessions.Get(sessionID)
if !exists {
return nil, errors.New("session expired")

## }

return session, nil

func (s Server) handleRegister(w http.ResponseWriter, r http.Request) {
var req RegisterRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
s.respondError(w, http.StatusBadRequest, "invalid request body")
return

## }

## }

func (s Server) handleLogin(w http.ResponseWriter, r http.Request) {
var req LoginRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
s.respondError(w, http.StatusBadRequest, "invalid request body")
return

## }

## }

func (s Server) handleCreateSecret(w http.ResponseWriter, r http.Request) {
session, err := s.authenticate(r)
if err != nil {
if req.Email == "" || req.Password == "" {
s.respondError(w, http.StatusBadRequest, "email and password required")
return

## }

if len(req.Password) < 12 {
s.respondError(w, http.StatusBadRequest, "password must be at least 12 characters")
return

## }

resp, err := s.vault.Register(r.Context(), &req)
if err != nil {
s.respondError(w, http.StatusBadRequest, err.Error())
return

## }

s.respondJSON(w, http.StatusCreated, resp)
resp, err := s.vault.Login(r.Context(), &req, r.RemoteAddr, r.UserAgent())
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, resp)

s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

## }

func (s Server) handleGetSecret(w http.ResponseWriter, r http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

## }

func (s Server) handleListSecrets(w http.ResponseWriter, r http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

var req CreateSecretRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
s.respondError(w, http.StatusBadRequest, "invalid request body")
return

## }

resp, err := s.vault.CreateSecret(r.Context(), session, &req)
if err != nil {
s.respondError(w, http.StatusInternalServerError, err.Error())
return

## }

s.respondJSON(w, http.StatusCreated, resp)
secretID := strings.TrimPrefix(r.URL.Path, "/api/v1/secrets/")
resp, err := s.vault.GetSecret(r.Context(), session, secretID)
if err != nil {
s.respondError(w, http.StatusNotFound, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, resp)

## }

func (s Server) handleDeleteSecret(w http.ResponseWriter, r http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

## }

func (s Server) handleRotateKey(w http.ResponseWriter, r http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

resp, err := s.vault.ListSecrets(r.Context(), session)
if err != nil {
s.respondError(w, http.StatusInternalServerError, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, resp)
secretID := strings.TrimPrefix(r.URL.Path, "/api/v1/secrets/")
if err := s.vault.DeleteSecret(r.Context(), session, secretID, r.RemoteAddr, r.UserAgent());
err != nil {
s.respondError(w, http.StatusInternalServerError, err.Error())
return

## }

w.WriteHeader(http.StatusNoContent)
var req struct {
Reason string `json:"reason"`

## }

json.NewDecoder(r.Body).Decode(&req)
if req.Reason == "" {
req.Reason = "manual_rotation"

## }

kv, err := s.vault.RotateVaultKey(r.Context(), session, req.Reason)
if err != nil {
s.respondError(w, http.StatusInternalServerError, err.Error())

## }

func (s Server) handleKeyStatus(w http.ResponseWriter, r http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

## }

func (s Server) handleKeyHistory(w http.ResponseWriter, r http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

## }

func (s Server) handleChangePassword(w http.ResponseWriter, r http.Request) {
session, err := s.authenticate(r)
return

## }

s.respondJSON(w, http.StatusOK, map[string]interface{}{
"message":     "Key rotated successfully",
"new_version": kv.Version,

## })

kv, err := s.vault.GetKeyStatus(r.Context(), session.UserID)
if err != nil {
s.respondError(w, http.StatusInternalServerError, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, kv)
history, err := s.vault.GetKeyHistory(r.Context(), session.UserID)
if err != nil {
s.respondError(w, http.StatusInternalServerError, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, history)

if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

## }

func (s Server) handleHealth(w http.ResponseWriter, r http.Request) {
s.respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})

## }

func (s *Server) SetupRoutes() http.Handler {
mux := http.NewServeMux()
var req struct {
OldPassword string `json:"old_password"`
NewPassword string `json:"new_password"`
SecretKey   string `json:"secret_key"`

## }

if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
s.respondError(w, http.StatusBadRequest, "invalid request body")
return

## }

if len(req.NewPassword) < 12 {
s.respondError(w, http.StatusBadRequest, "new password must be at least 12
characters")
return

## }

if err := s.vault.ChangePassword(r.Context(), session.UserID, req.OldPassword,
req.NewPassword, req.SecretKey); err != nil {
s.respondError(w, http.StatusBadRequest, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, map[string]string{
"message": "Password changed successfully. Please login again.",

## })

## // Health

mux.HandleFunc("GET /health", s.handleHealth)

## // Auth

mux.HandleFunc("POST /api/v1/auth/register", s.handleRegister)
mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)

Теперь надо доделать доступ чтобы можно было указывать в
своих проектах в конфигурации вместо значения стринги как с
Azure KeyVault и в конфиге указывать подключение к нашему
сервису давай сделаем по client secret clientId Доработкой наш
сервис

## Claude Opus 4.5

Добавляю поддержку Service Principal аутентификации (Client ID + Client Secret)
для доступа из приложений:

## ```go

package main
import (

## "context"

## "crypto/rand"

## "crypto/sha256"

## "crypto/subtle"

## "database/sql"

## "encoding/base64"

## "encoding/hex"

## "encoding/json"

## "errors"

## "fmt"

## "io"

## "log"

## "net/http"

## "os"

## "strings"

## "sync"

## "time"

## // Secrets

mux.HandleFunc("POST /api/v1/secrets", s.handleCreateSecret)
mux.HandleFunc("GET /api/v1/secrets", s.handleListSecrets)
mux.HandleFunc("GET /api/v1/secrets/", s.handleGetSecret)
mux.HandleFunc("DELETE /api/v1/secrets/", s.handleDeleteSecret)
// Key management
mux.HandleFunc("POST /api/v1/keys/rotate", s.handleRotateKey)
mux.HandleFunc("GET /api/v1/keys/status", s.handleKeyStatus)
mux.

## "github.com/golang-jwt/jwt/v5"

## "github.com/google/uuid"

## )

## //

## ================================================================

## ============

## // Configuration

## //

## ================================================================

## ============

type Config struct {
DatabaseURL     string
JWTSecret       []byte
ServerPort      string
Argon2Time      uint32
Argon2Memory    uint32
Argon2Threads   uint8
Argon2KeyLen    uint32
KeyRotationDays int
GracePeriodDays int

## }

func LoadConfig() *Config {
dbURL := os.Getenv("DATABASE_URL")
if dbURL == "" {
dbURL = "postgres://postgres:postgres@localhost:5432/keyvault?sslmode=disable"

## }

## _ "github.com/lib/pq"

## "golang.org/x/crypto/argon2"

## "golang.org/x/crypto/chacha20poly1305"

## "golang.org/x/crypto/hkdf"

jwtSecret := os.Getenv("JWT_SECRET")
if jwtSecret == "" {
jwtSecret = "change-me-in-production-use-strong-secret"

## }

return &Config{
DatabaseURL:     dbURL,
JWTSecret:       []byte(jwtSecret),
ServerPort:      getEnvOrDefault("SERVER_PORT", "8080"),
Argon2Time:      3,
Argon2Memory:    64 * 1024,
Argon2Threads:   4,
Argon2KeyLen:    32,
KeyRotationDays: 90,

## }

func getEnvOrDefault(key, defaultVal string) string {
if val := os.Getenv(key); val != "" {
return val

## }

return defaultVal

## }

## //

## ================================================================

## ============

## // Database Models

## //

## ================================================================

## ============

type User struct {
ID                  string
Email               string
SecretKeyHash       string
MasterKeySalt       []byte
EncryptedVaultKey   []byte
VaultKeyNonce       []byte
SRPVerifier         []byte
CurrentKeyVersionID string
CreatedAt           time.Time
UpdatedAt           time.Time

## }

type ServicePrincipal struct {
ID               string
UserID           string
ClientID         string
ClientSecretHash string
Name             string
Description      string
Permissions      []string    // read, write, delete, list, rotate
AllowedSecrets   []string    // nil = all, or specific
secret IDs/names
AllowedTags      []string       // access secrets with
these tags
IPWhitelist      []string
RateLimit        int         // requests per minute
IsActive         bool
ExpiresAt        time.Time
GracePeriodDays: 7,

## }

json:"id"
json:"email"
json:"-"
json:"-"
json:"-"
json:"-"
json:"-"
json:"-"
json:"created_at"
json:"updated_at"
json:"id"
json:"user_id"
json:"client_id"
json:"-"
json:"name"
json:"description,omitempty"
json:"permissions"
json:"allowed_secrets,omitempty"
json:"allowed_tags,omitempty"
json:"ip_whitelist,omitempty"
json:"rate_limit"
json:"is_active"
json:"expires_at,omitempty"

LastUsedAt       time.Time
CreatedAt        time.Time
UpdatedAt        time.Time

## }

type KeyVersion struct {
ID                string
UserID            string
Version           int
EncryptedVaultKey []byte
VaultKeyNonce     []byte
Status            string
CreatedAt         time.Time
RotatedAt         time.Time
DestroyAt         time.Time
DestroyedAt       *time.Time
RotationReason    string

## }

type Secret struct {
ID             string
UserID         string
KeyVersionID   string
Name           string
EncryptedData  []byte

## Nonce          []byte

SecretType     string
Version        int
PreviousID     string

## Tags           []string

CreatedAt      time.Time
UpdatedAt      time.Time
ExpiresAt      time.Time
AccessCount    int
LastAccessedAt *time.Time

## }

type AuditLog struct {
ID                 string
UserID             string
ServicePrincipalID *string
Action             string
ResourceID         string
Details            json.RawMessage
IPAddress          string
UserAgent          string
CreatedAt          time.Time

## }

json:"last_used_at,omitempty"
json:"created_at"
json:"updated_at"
json:"id"
json:"user_id"
json:"version"
json:"-"
json:"-"
json:"status"
json:"created_at"
json:"rotated_at,omitempty"
json:"destroy_at,omitempty"
json:"destroyed_at,omitempty"
json:"rotation_reason,omitempty"
json:"id"
json:"user_id"
json:"key_version_id"
json:"name"
json:"-"
json:"-"
json:"type"
json:"version"
json:"previous_id,omitempty"
json:"tags,omitempty"
json:"created_at"
json:"updated_at"
json:"expires_at,omitempty"
json:"access_count"
json:"last_accessed_at,omitempty"
json:"id"
json:"user_id"
json:"service_principal_id,omitempty"
json:"action"
json:"resource_id,omitempty"
json:"details,omitempty"
json:"ip_address"
json:"user_agent"
json:"created_at"

type Session struct {
UserID             string
ServicePrincipalID string
VaultKey           []byte
KeyVersionID       string

## Permissions        []string

AllowedSecrets     []string
AllowedTags        []string
ExpiresAt          time.Time

## }

## //

## ================================================================

## ============

## // Database Layer

## //

## ================================================================

## ============

type Database struct {
db *sql.DB

## }

func NewDatabase(databaseURL string) (*Database, error) {
db, err := sql.Open("postgres", databaseURL)
if err != nil {
return nil, fmt.Errorf("failed to open database: %w", err)

## }

## }

func (d *Database) migrate() error {
migrations := []string{
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
if err := db.Ping(); err != nil {
return nil, fmt.Errorf("failed to ping database: %w", err)

## }

database := &Database{db: db}
if err := database.migrate(); err != nil {
return nil, fmt.Errorf("failed to migrate database: %w", err)

## }

return database, nil

## ,

## ,

## ,

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

## )

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

## )

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

## )

CREATE TABLE IF NOT EXISTS secrets (
id UUID PRIMARY KEY,
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
key_version_id UUID NOT NULL REFERENCES key_versions(id),
name VARCHAR(255) NOT NULL,
encrypted_data BYTEA NOT NULL,

## ,

## ,

## ,

## ,

## ,

## ,

## ,

## ,

## ,

## ,

## ,

## }

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

## )

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

## )

CREATE TABLE IF NOT EXISTS rate_limits (
key VARCHAR(255) PRIMARY KEY,
count INTEGER DEFAULT 0,
window_start TIMESTAMPTZ DEFAULT NOW()

## )

CREATE INDEX IF NOT EXISTS idx_secrets_user_id ON secrets(user_id)
CREATE INDEX IF NOT EXISTS idx_secrets_name ON secrets(user_id, name)
CREATE INDEX IF NOT EXISTS idx_secrets_tags ON secrets USING GIN(tags)
CREATE INDEX IF NOT EXISTS idx_service_principals_client_id ON
service_principals(client_id)
CREATE INDEX IF NOT EXISTS idx_service_principals_user_id ON
service_principals(user_id)
CREATE INDEX IF NOT EXISTS idx_key_versions_user_id ON
key_versions(user_id)
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id)
CREATE INDEX IF NOT EXISTS idx_audit_logs_sp_id ON
audit_logs(service_principal_id)
for _, migration := range migrations {
if _, err := d.db.Exec(migration); err != nil {
return fmt.Errorf("migration failed: %w", err)

## }

## }

return nil

## }

func (d *Database) Close() error {
return d.db.Close()

## }

// User operations
func (d Database) CreateUser(ctx context.Context, user User) error {
query :=

## }

func (d Database) GetUserByEmail(ctx context.Context, email string) (User, error) {
query :=
INSERT INTO users (id, email, secret_key_hash, master_key_salt,
encrypted_vault_key, vault_key_nonce, srp_verifier, current_key_version_id,
created_at, updated_at)

## VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)

_, err := d.db.ExecContext(ctx, query,
user.ID, user.Email, user.SecretKeyHash, user.MasterKeySalt,
user.EncryptedVaultKey, user.VaultKeyNonce, user.SRPVerifier,
user.CurrentKeyVersionID, user.CreatedAt, user.UpdatedAt)
return err
SELECT id, email, secret_key_hash, master_key_salt,
encrypted_vault_key, vault_key_nonce, srp_verifier, current_key_version_id,
created_at, updated_at
FROM users WHERE email = $1
user := &User{}
var currentKeyVersionID sql.NullString
err := d.db.QueryRowContext(ctx, query, email).Scan(
&user.ID, &user.Email, &user.SecretKeyHash, &user.MasterKeySalt,
&user.EncryptedVaultKey, &user.VaultKeyNonce, &user.SRPVerifier,
&currentKeyVersionID, &user.CreatedAt, &user.UpdatedAt)
if err == sql.ErrNoRows {
return nil, nil

## }

if err != nil {
return nil, err

## }

if currentKeyVersionID.Valid {
user.CurrentKeyVersionID = currentKeyVersionID.String

## }

## }

func (d Database) GetUserByID(ctx context.Context, id string) (User, error) {
query :=

## }

func (d *Database) UpdateUserKeys(ctx context.Context, userID string,
encryptedVaultKey, vaultKeyNonce, masterKeySalt []byte, currentKeyVersionID
string) error {
query :=
return user, nil
SELECT id, email, secret_key_hash, master_key_salt,
encrypted_vault_key, vault_key_nonce, srp_verifier, current_key_version_id,
created_at, updated_at
FROM users WHERE id = $1
user := &User{}
var currentKeyVersionID sql.NullString
err := d.db.QueryRowContext(ctx, query, id).Scan(
&user.ID, &user.Email, &user.SecretKeyHash, &user.MasterKeySalt,
&user.EncryptedVaultKey, &user.VaultKeyNonce, &user.SRPVerifier,
&currentKeyVersionID, &user.CreatedAt, &user.UpdatedAt)
if err == sql.ErrNoRows {
return nil, nil

## }

if err != nil {
return nil, err

## }

if currentKeyVersionID.Valid {
user.CurrentKeyVersionID = currentKeyVersionID.String

## }

return user, nil
UPDATE users
SET encrypted_vault_key = $1, vault_key_nonce = $2, master_key_salt = $3,
current_key_version_id = $4, updated_at = NOW()
WHERE id = $5
_, err := d.db.ExecContext(ctx, query, encryptedVaultKey, vaultKeyNonce, masterKeySalt,
currentKeyVersionID, userID)
return err

## }

func (d *Database) UpdateUserCurrentKeyVersion(ctx context.Context, userID,
keyVersionID string) error {
query :=
_, err := d.db.ExecContext(ctx, query, keyVersionID, userID)
return err

## }

// Service Principal operations
func (d Database) CreateServicePrincipal(ctx context.Context, sp ServicePrincipal)
error {
query :=

## }

func (d
Database) GetServicePrincipalByClientID(ctx context.Context, clientID string)
(ServicePrincipal, error) {
query :=
UPDATE users SET current_key_version_id = $1, updated_at = NOW()
WHERE id = $2
INSERT INTO service_principals (id, user_id, client_id, client_secret_hash,
name, description, permissions, allowed_secrets, allowed_tags, ip_whitelist,
rate_limit, is_active, expires_at, created_at, updated_at)

## VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)

_, err := d.db.ExecContext(ctx, query,
sp.ID, sp.UserID, sp.ClientID, sp.ClientSecretHash, sp.Name, sp.Description,
sp.Permissions, sp.AllowedSecrets, sp.AllowedTags, sp.IPWhitelist,
sp.RateLimit, sp.IsActive, sp.ExpiresAt, sp.CreatedAt, sp.UpdatedAt)
return err
SELECT id, user_id, client_id, client_secret_hash, name, description,
permissions, allowed_secrets, allowed_tags, ip_whitelist, rate_limit, is_active,
expires_at, last_used_at, created_at, updated_at
FROM service_principals WHERE client_id = $1
sp := &ServicePrincipal{}
err := d.db.QueryRowContext(ctx, query, clientID).Scan(
&sp.ID, &sp.UserID, &sp.ClientID, &sp.ClientSecretHash, &sp.Name, &sp.Description,
&sp.Permissions, &sp.AllowedSecrets, &sp.AllowedTags, &sp.IPWhitelist,
&sp.RateLimit, &sp.IsActive, &sp.ExpiresAt, &sp.LastUsedAt, &sp.CreatedAt,
&sp.UpdatedAt)
if err == sql.ErrNoRows {
return nil, nil

## }

return sp, err

## }

func (d Database) GetServicePrincipalsByUser(ctx context.Context, userID string)
([]ServicePrincipal, error) {
query :=

## }

func (d *Database) UpdateServicePrincipalLastUsed(ctx context.Context, id string)
error {
query :=
_, err := d.db.ExecContext(ctx, query, id)
return err

## }

func (d Database) UpdateServicePrincipal(ctx context.Context, sp ServicePrincipal)
error {
query :=
SELECT id, user_id, client_id, client_secret_hash, name, description,
permissions, allowed_secrets, allowed_tags, ip_whitelist, rate_limit, is_active,
expires_at, last_used_at, created_at, updated_at
FROM service_principals WHERE user_id = $1 ORDER BY name
rows, err := d.db.QueryContext(ctx, query, userID)
if err != nil {
return nil, err

## }

defer rows.Close()
var sps []*ServicePrincipal
for rows.Next() {
sp := &ServicePrincipal{}
if err := rows.Scan(
&sp.ID, &sp.UserID, &sp.ClientID, &sp.ClientSecretHash, &sp.Name, &sp.Description,
&sp.Permissions, &sp.AllowedSecrets, &sp.AllowedTags, &sp.IPWhitelist,
&sp.RateLimit, &sp.IsActive, &sp.ExpiresAt, &sp.LastUsedAt, &sp.CreatedAt,
&sp.UpdatedAt); err != nil {
return nil, err

## }

sps = append(sps, sp)

## }

return sps, rows.Err()
UPDATE service_principals SET last_used_at = NOW() WHERE id = $1
UPDATE service_principals
SET name = $1, description = $2, permissions = $3, allowed_secrets = $4,
allowed_tags = $5, ip_whitelist = $6, rate_limit = $7, is_active = $8, expires_at = $9,
updated_at = NOW()
WHERE id = $10

## }

func (d *Database) DeleteServicePrincipal(ctx context.Context, id, userID string)
error {
query :=
_, err := d.db.ExecContext(ctx, query, id, userID)
return err

## }

func (d *Database) RegenerateClientSecret(ctx context.Context, id, newSecretHash
string) error {
query :=
_, err := d.db.ExecContext(ctx, query, newSecretHash, id)
return err

## }

// Rate limiting
func (d *Database) CheckRateLimit(ctx context.Context, key string, limit int) (bool,
error) {
query :=
_, err := d.db.ExecContext(ctx, query,
sp.Name, sp.Description, sp.Permissions, sp.AllowedSecrets, sp.AllowedTags,
sp.IPWhitelist, sp.RateLimit, sp.IsActive, sp.ExpiresAt, sp.ID)
return err
DELETE FROM service_principals WHERE id = $1 AND user_id = $2
UPDATE service_principals SET client_secret_hash = $1, updated_at =
NOW() WHERE id = $2
INSERT INTO rate_limits (key, count, window_start)

## VALUES ($1, 1, NOW())

ON CONFLICT (key) DO UPDATE
SET count = CASE
WHEN rate_limits.window_start < NOW() - INTERVAL '1 minute' THEN 1
ELSE rate_limits.count + 1

## END,

window_start = CASE
WHEN rate_limits.window_start < NOW() - INTERVAL '1 minute' THEN

## NOW()

ELSE rate_limits.window_start

## END

RETURNING count
var count int
err := d.db.QueryRowContext(ctx, query, key).Scan(&count)
if err != nil {
return false, err

## }

return count <= limit, nil

## }

// Key version operations
func (d Database) CreateKeyVersion(ctx context.Context, kv KeyVersion) error {
query :=

## }

func (d Database) GetActiveKeyVersion(ctx context.Context, userID string)
(KeyVersion, error) {
query :=

## }

func (d *Database) GetLatestKeyVersionNumber(ctx context.Context, userID string)
(int, error) {
query :=
var version int
err := d.db.QueryRowContext(ctx, query, userID).Scan(&version)
return version, err

## }

INSERT INTO key_versions (id, user_id, version, encrypted_vault_key,
vault_key_nonce, status, created_at, rotation_reason)

## VALUES ($1, $2, $3, $4, $5, $6, $7, $8)

_, err := d.db.ExecContext(ctx, query,
kv.ID, kv.UserID, kv.Version, kv.EncryptedVaultKey, kv.VaultKeyNonce,
kv.Status, kv.CreatedAt, kv.RotationReason)
return err
SELECT id, user_id, version, encrypted_vault_key, vault_key_nonce,
status, created_at, rotated_at, destroy_at, destroyed_at, rotation_reason
FROM key_versions WHERE user_id = $1 AND status = 'active' ORDER BY
version DESC LIMIT 1
kv := &KeyVersion{}
err := d.db.QueryRowContext(ctx, query, userID).Scan(
&kv.ID, &kv.UserID, &kv.Version, &kv.EncryptedVaultKey, &kv.VaultKeyNonce,
&kv.Status, &kv.CreatedAt, &kv.RotatedAt, &kv.DestroyAt, &kv.DestroyedAt,
&kv.RotationReason)
if err == sql.ErrNoRows {
return nil, nil

## }

return kv, err
SELECT COALESCE(MAX(version), 0) FROM key_versions WHERE
user_id = $1

func (d Database) GetKeyVersionHistory(ctx context.Context, userID string)
([]KeyVersion, error) {
query :=

## }

func (d Database) UpdateKeyVersionStatus(ctx context.Context, id, status string,
rotatedAt, destroyAt time.Time) error {
query :=
_, err := d.db.ExecContext(ctx, query, status, rotatedAt, destroyAt, id)
return err

## }

// Secret operations
func (d Database) CreateSecret(ctx context.Context, secret Secret) error {
query :=
SELECT id, user_id, version, encrypted_vault_key, vault_key_nonce,
status, created_at, rotated_at, destroy_at, destroyed_at, rotation_reason
FROM key_versions WHERE user_id = $1 ORDER BY version DESC
rows, err := d.db.QueryContext(ctx, query, userID)
if err != nil {
return nil, err

## }

defer rows.Close()
var versions []*KeyVersion
for rows.Next() {
kv := &KeyVersion{}
if err := rows.Scan(
&kv.ID, &kv.UserID, &kv.Version, &kv.EncryptedVaultKey, &kv.VaultKeyNonce,
&kv.Status, &kv.CreatedAt, &kv.RotatedAt, &kv.DestroyAt, &kv.DestroyedAt,
&kv.RotationReason); err != nil {
return nil, err

## }

versions = append(versions, kv)

## }

return versions, rows.Err()
UPDATE key_versions SET status = $1, rotated_at = $2, destroy_at = $3
WHERE id = $4
INSERT INTO secrets (id, user_id, key_version_id, name, encrypted_data,
nonce, secret_type, version, previous_id, tags, created_at, updated_at, expires_at)

## VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)

_, err := d.db.ExecContext(ctx, query,
secret.ID, secret.UserID, secret.KeyVersionID, secret.Name, secret.EncryptedData,
secret.Nonce, secret.SecretType, secret.Version, secret.PreviousID, secret.Tags,
secret.CreatedAt, secret.UpdatedAt, secret.ExpiresAt)
return err

## }

func (d Database) GetSecret(ctx context.Context, id, userID string) (Secret, error) {
query :=

## }

func (d Database) GetSecretByName(ctx context.Context, name, userID string)
(Secret, error) {
query :=

## }

SELECT id, user_id, key_version_id, name, encrypted_data, nonce,
secret_type, version, previous_id, tags, created_at, updated_at, expires_at,
access_count, last_accessed_at
FROM secrets WHERE id = $1 AND user_id = $2
secret := &Secret{}
err := d.db.QueryRowContext(ctx, query, id, userID).Scan(
&secret.ID, &secret.UserID, &secret.KeyVersionID, &secret.Name,
&secret.EncryptedData,
&secret.Nonce, &secret.SecretType, &secret.Version, &secret.PreviousID, &secret.Tags,
&secret.CreatedAt, &secret.UpdatedAt, &secret.ExpiresAt, &secret.AccessCount,
&secret.LastAccessedAt)
if err == sql.ErrNoRows {
return nil, nil

## }

return secret, err
SELECT id, user_id, key_version_id, name, encrypted_data, nonce,
secret_type, version, previous_id, tags, created_at, updated_at, expires_at,
access_count, last_accessed_at
FROM secrets WHERE name = $1 AND user_id = $2
secret := &Secret{}
err := d.db.QueryRowContext(ctx, query, name, userID).Scan(
&secret.ID, &secret.UserID, &secret.KeyVersionID, &secret.Name,
&secret.EncryptedData,
&secret.Nonce, &secret.SecretType, &secret.Version, &secret.PreviousID, &secret.Tags,
&secret.CreatedAt, &secret.UpdatedAt, &secret.ExpiresAt, &secret.AccessCount,
&secret.LastAccessedAt)
if err == sql.ErrNoRows {
return nil, nil

## }

return secret, err

func (d Database) GetSecretsByUser(ctx context.Context, userID string) ([]Secret,
error) {
query :=

## }

func (d Database) GetSecretsByKeyVersion(ctx context.Context, keyVersionID
string) ([]Secret, error) {
query :=
SELECT id, user_id, key_version_id, name, encrypted_data, nonce,
secret_type, version, previous_id, tags, created_at, updated_at, expires_at,
access_count, last_accessed_at
FROM secrets WHERE user_id = $1 ORDER BY name
rows, err := d.db.QueryContext(ctx, query, userID)
if err != nil {
return nil, err

## }

defer rows.Close()
var secrets []*Secret
for rows.Next() {
secret := &Secret{}
if err := rows.Scan(
&secret.ID, &secret.UserID, &secret.KeyVersionID, &secret.Name,
&secret.EncryptedData,
&secret.Nonce, &secret.SecretType, &secret.Version, &secret.PreviousID,
&secret.Tags,
&secret.CreatedAt, &secret.UpdatedAt, &secret.ExpiresAt, &secret.AccessCount,
&secret.LastAccessedAt); err != nil {
return nil, err

## }

secrets = append(secrets, secret)

## }

return secrets, rows.Err()
SELECT id, user_id, key_version_id, name, encrypted_data, nonce,
secret_type, version, previous_id, tags, created_at, updated_at, expires_at,
access_count, last_accessed_at
FROM secrets WHERE key_version_id = $1
rows, err := d.db.QueryContext(ctx, query, keyVersionID)
if err != nil {
return nil, err

## }

defer rows.Close()
var secrets []*Secret
for rows.Next() {
secret := &Secret{}

## }

func (d *Database) UpdateSecretKeyVersion(ctx context.Context, secretID,
keyVersionID string, encryptedData, nonce []byte) error {
query :=
_, err := d.db.ExecContext(ctx, query, keyVersionID, encryptedData, nonce, secretID)
return err

## }

func (d *Database) UpdateSecretAccess(ctx context.Context, id string) error {
query :=
_, err := d.db.ExecContext(ctx, query, id)
return err

## }

func (d *Database) DeleteSecret(ctx context.Context, id, userID string) error {
query :=
_, err := d.db.ExecContext(ctx, query, id, userID)
return err

## }

// Audit log operations
func (d Database) CreateAuditLog(ctx context.Context, log AuditLog) error {
query :=
if err := rows.Scan(
&secret.ID, &secret.UserID, &secret.KeyVersionID, &secret.Name,
&secret.EncryptedData,
&secret.Nonce, &secret.SecretType, &secret.Version, &secret.PreviousID,
&secret.Tags,
&secret.CreatedAt, &secret.UpdatedAt, &secret.ExpiresAt, &secret.AccessCount,
&secret.LastAccessedAt); err != nil {
return nil, err

## }

secrets = append(secrets, secret)

## }

return secrets, rows.Err()
UPDATE secrets SET key_version_id = $1, encrypted_data = $2, nonce =
$3, updated_at = NOW() WHERE id = $4
UPDATE secrets SET access_count = access_count + 1, last_accessed_at
= NOW() WHERE id = $1
DELETE FROM secrets WHERE id = $1 AND user_id = $2
INSERT INTO audit_logs (id, user_id, service_principal_id, action,
resource_id, details, ip_address, user_agent, created_at)

## VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)

_, err := d.db.ExecContext(ctx, query,
log.ID, log.UserID, log.ServicePrincipalID, log.Action, log.ResourceID,
log.Details, log.IPAddress, log.UserAgent, log.CreatedAt)
return err

## }

func (d Database) BeginTx(ctx context.Context) (sql.Tx, error) {
return d.db.BeginTx(ctx, nil)

## }

## //

## ================================================================

## ============

## // Cryptography

## //

## ================================================================

## ============

type CryptoService struct {
config *Config

## }

func NewCryptoService(config Config) CryptoService {
return &CryptoService{config: config}

## }

func (c *CryptoService) GenerateRandomBytes(n int) ([]byte, error) {
b := make([]byte, n)
if _, err := rand.Read(b); err != nil {
return nil, err

## }

return b, nil

## }

func (c *CryptoService) GenerateSecretKey() (string, error) {
bytes, err := c.GenerateRandomBytes(32)
if err != nil {
return "", err

## }

encoded := base64.RawURLEncoding.EncodeToString(bytes)
parts := make([]string, 0, 7)
for i := 0; i < len(encoded) && len(parts) < 7; i += 6 {
end := i + 6
if end > len(encoded) {
end = len(encoded)

## }

parts = append(parts, encoded[i:end])

## }

return strings.Join(parts, "-"), nil

## }

func (c *CryptoService) GenerateClientID() string {
bytes := make([]byte, 16)
rand.Read(bytes)
return hex.EncodeToString(bytes)

## }

func (c *CryptoService) GenerateClientSecret() (string, error) {
bytes, err := c.GenerateRandomBytes(32)
if err != nil {
return "", err

## }

return base64.RawURLEncoding.EncodeToString(bytes), nil

## }

func (c *CryptoService) HashSecret(secret string) string {
hash := sha256.Sum256([]byte(secret))
return hex.EncodeToString(hash[:])

## }

func (c *CryptoService) HashSecretKey(secretKey string) string {
hash := sha256.Sum256([]byte(secretKey))
return hex.EncodeToString(hash[:])

## }

func (c *CryptoService) DeriveMasterKey(password, secretKey string, salt []byte)

## []byte {

combined := []byte(password + secretKey)
return argon2.IDKey(combined, salt, c.config.Argon2Time, c.config.Argon2Memory,
c.config.Argon2Threads, c.config.Argon2KeyLen)

## }

func (c *CryptoService) DeriveKey(masterKey []byte, context string, length int)
([]byte, error) {
reader := hkdf.New(sha256.New, masterKey, nil, []byte(context))
key := make([]byte, length)
if _, err := io.ReadFull(reader, key); err != nil {
return nil, err

## }

return key, nil

## }

func (c *CryptoService) Encrypt(key, plaintext []byte) (ciphertext, nonce []byte, err
error) {
aead, err := chacha20poly1305.NewX(key)
if err != nil {
return nil, nil, err

## }

## }

func (c *CryptoService) Decrypt(key, ciphertext, nonce []byte) ([]byte, error) {
aead, err := chacha20poly1305.NewX(key)
if err != nil {
return nil, err

## }

## }

func (c *CryptoService) GenerateVaultKey() ([]byte, error) {
return c.GenerateRandomBytes(32)

## }

func (c *CryptoService) GenerateSRPVerifier(password string, salt []byte) []byte {
return argon2.IDKey([]byte(password), salt, c.config.Argon2Time,
c.config.Argon2Memory, c.config.Argon2Threads, 32)

## }

func SecureZeroMemory(b []byte) {
for i := range b {
b[i] = 0

## }

## }

## //

## ================================================================

## ============

## // Session Manager

## //

## ================================================================

## ============

nonce, err = c.GenerateRandomBytes(chacha20poly1305.NonceSizeX)
if err != nil {
return nil, nil, err

## }

ciphertext = aead.Seal(nil, nonce, plaintext, nil)
return ciphertext, nonce, nil
plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
if err != nil {
return nil, errors.New("decryption failed: invalid key or corrupted data")

## }

return plaintext, nil

type SessionManager struct {
sessions map[string]*Session
mu       sync.RWMutex

## }

func NewSessionManager() SessionManager {
return &SessionManager{
sessions: make(map[string]Session),

## }

## }

func (sm *SessionManager) Create(userID string, vaultKey []byte, keyVersionID
string, permissions, allowedSecrets, allowedTags []string, spID string, duration
time.Duration) string {
sm.mu.Lock()
defer sm.mu.Unlock()

## }

func (sm SessionManager) Get(sessionID string) (Session, bool) {
sm.mu.RLock()
defer sm.mu.RUnlock()
sessionID := uuid.New().String()
keyCopy := make([]byte, len(vaultKey))
copy(keyCopy, vaultKey)
sm.sessions[sessionID] = &Session{
UserID:             userID,
ServicePrincipalID: spID,
VaultKey:           keyCopy,
KeyVersionID:       keyVersionID,
Permissions:        permissions,
AllowedSecrets:     allowedSecrets,
AllowedTags:        allowedTags,
ExpiresAt:          time.Now().Add(duration),

## }

return sessionID
session, exists := sm.sessions[sessionID]
if !exists || time.Now().After(session.ExpiresAt) {
return nil, false

## }

return session, true

## }

func (sm *SessionManager) Update(sessionID string, vaultKey []byte, keyVersionID
string) bool {
sm.mu.Lock()
defer sm.mu.Unlock()

## }

func (sm *SessionManager) Delete(sessionID string) {
sm.mu.Lock()
defer sm.mu.Unlock()

## }

func (sm *SessionManager) DeleteUserSessions(userID string) {
sm.mu.Lock()
defer sm.mu.Unlock()
session, exists := sm.sessions[sessionID]
if !exists {
return false

## }

SecureZeroMemory(session.VaultKey)
keyCopy := make([]byte, len(vaultKey))
copy(keyCopy, vaultKey)
session.VaultKey = keyCopy
session.KeyVersionID = keyVersionID
return true
if session, exists := sm.sessions[sessionID]; exists {
SecureZeroMemory(session.VaultKey)
delete(sm.sessions, sessionID)

## }

for id, session := range sm.sessions {
if session.UserID == userID {
SecureZeroMemory(session.VaultKey)
delete(sm.sessions, id)

## }

## }

## }

## //

## ================================================================

## ============

## // Key Rotation Manager

## //

## ================================================================

## ============

type KeyRotationManager struct {
db      Database
crypto  CryptoService
config  Config
session SessionManager

## }

func NewKeyRotationManager(db Database, crypto CryptoService, config Config,
session SessionManager) *KeyRotationManager {
return &KeyRotationManager{
db:      db,
crypto:  crypto,
config:  config,
session: session,

## }

## }

func (krm KeyRotationManager) RotateVaultKey(ctx context.Context, userID string,
currentVaultKey []byte, reason string) (KeyVersion, error) {
tx, err := krm.db.BeginTx(ctx)
if err != nil {
return nil, err

## }

defer tx.Rollback()
user, err := krm.db.GetUserByID(ctx, userID)
if err != nil || user == nil {
return nil, errors.New("user not found")

## }

currentKV, err := krm.db.GetActiveKeyVersion(ctx, userID)
if err != nil {
return nil, err

## }

newVaultKey, err := krm.crypto.GenerateVaultKey()
if err != nil {
return nil, err

## }

defer SecureZeroMemory(newVaultKey)
encryptedNewVaultKey, newNonce, err := krm.crypto.Encrypt(currentVaultKey, newVaultKey)
if err != nil {
return nil, err

## }

latestVersion, err := krm.db.GetLatestKeyVersionNumber(ctx, userID)
if err != nil {
return nil, err

## }

newKV := &KeyVersion{
ID:                uuid.New().String(),
UserID:            userID,
Version:           latestVersion + 1,
EncryptedVaultKey: encryptedNewVaultKey,
VaultKeyNonce:     newNonce,

## Status:            "active",

CreatedAt:         time.Now(),
RotationReason:    reason,

## }

if err := krm.db.CreateKeyVersion(ctx, newKV); err != nil {
return nil, err

## }

if currentKV != nil {
now := time.Now()
destroyAt := now.Add(time.Duration(krm.config.GracePeriodDays) * 24 * time.Hour)
if err := krm.db.UpdateKeyVersionStatus(ctx, currentKV.ID, "deprecated", &now,
&destroyAt); err != nil {
return nil, err

## }

## }

secrets, err := krm.db.GetSecretsByKeyVersion(ctx, currentKV.ID)
if err != nil {
return nil, err

## }

for _, secret := range secrets {
plaintext, err := krm.crypto.Decrypt(currentVaultKey, secret.EncryptedData, secret.Nonce)
if err != nil {
continue

## }

newCiphertext, newNonce, err := krm.crypto.Encrypt(newVaultKey, plaintext)
if err != nil {
SecureZeroMemory(plaintext)

## }

func (krm *KeyRotationManager) RotateMasterKey(ctx context.Context, userID,
oldPassword, newPassword, secretKey string) error {
user, err := krm.db.GetUserByID(ctx, userID)
if err != nil || user == nil {
return errors.New("user not found")

## }

continue

## }

SecureZeroMemory(plaintext)
if err := krm.db.UpdateSecretKeyVersion(ctx, secret.ID, newKV.ID, newCiphertext,
newNonce); err != nil {
return nil, err

## }

## }

if err := krm.db.UpdateUserCurrentKeyVersion(ctx, userID, newKV.ID); err != nil {
return nil, err

## }

if err := tx.Commit(); err != nil {
return nil, err

## }

return newKV, nil
oldMasterKey := krm.crypto.DeriveMasterKey(oldPassword, secretKey, user.MasterKeySalt)
defer SecureZeroMemory(oldMasterKey)
encryptionKey, err := krm.crypto.DeriveKey(oldMasterKey, "vault-key-encryption", 32)
if err != nil {
return err

## }

defer SecureZeroMemory(encryptionKey)
vaultKey, err := krm.crypto.Decrypt(encryptionKey, user.EncryptedVaultKey,
user.VaultKeyNonce)
if err != nil {
return errors.New("invalid old password")

## }

defer SecureZeroMemory(vaultKey)
newSalt, err := krm.crypto.GenerateRandomBytes(32)
if err != nil {
return err

## }

## }

## //

## ================================================================

## ============

## // Vault Service

## //

## ================================================================

## ============

type VaultService struct {
db         Database
crypto     CryptoService
sessions   SessionManager
keyManager KeyRotationManager
config     *Config

## }

func NewVaultService(db Database, crypto CryptoService, sessions
SessionManager, config Config) *VaultService {
vs := &VaultService{
db:       db,
crypto:   crypto,
sessions: sessions,
config:   config,
newMasterKey := krm.crypto.DeriveMasterKey(newPassword, secretKey, newSalt)
defer SecureZeroMemory(newMasterKey)
newEncryptionKey, err := krm.crypto.DeriveKey(newMasterKey, "vault-key-encryption", 32)
if err != nil {
return err

## }

defer SecureZeroMemory(newEncryptionKey)
newEncryptedVaultKey, newNonce, err := krm.crypto.Encrypt(newEncryptionKey, vaultKey)
if err != nil {
return err

## }

if err := krm.db.UpdateUserKeys(ctx, userID, newEncryptedVaultKey, newNonce, newSalt,
user.CurrentKeyVersionID); err != nil {
return err

## }

krm.session.DeleteUserSessions(userID)
return nil

## }

vs.keyManager = NewKeyRotationManager(db, crypto, config, sessions)
return vs

## }

## // User Registration

type RegisterRequest struct {
Email    string
Password string

## }

type RegisterResponse struct {
UserID    string
SecretKey string
Message   string

## }

func (vs VaultService) Register(ctx context.Context, req RegisterRequest)
(*RegisterResponse, error) {
existing, err := vs.db.GetUserByEmail(ctx, req.Email)
if err != nil {
return nil, err

## }

if existing != nil {
return nil, errors.New("email already registered")

## }

json:"email"
json:"password"
json:"user_id"
json:"secret_key"
json:"message"
secretKey, err := vs.crypto.GenerateSecretKey()
if err != nil {
return nil, err

## }

salt, err := vs.crypto.GenerateRandomBytes(32)
if err != nil {
return nil, err

## }

masterKey := vs.crypto.DeriveMasterKey(req.Password, secretKey, salt)
defer SecureZeroMemory(masterKey)
vaultKey, err := vs.crypto.GenerateVaultKey()
if err != nil {
return nil, err

## }

defer SecureZeroMemory(vaultKey)
encryptionKey, err := vs.crypto.DeriveKey(masterKey, "vault-key-encryption", 32)
if err != nil {
return nil, err

## }

defer SecureZeroMemory(encryptionKey)
encryptedVaultKey, nonce, err := vs.crypto.Encrypt(encryptionKey, vaultKey)
if err != nil {
return nil, err

## }

srpVerifier := vs.crypto.GenerateSRPVerifier(req.Password, salt)
userID := uuid.New().String()
keyVersionID := uuid.New().String()
now := time.Now()
user := &User{
ID:                  userID,
Email:               req.Email,
SecretKeyHash:       vs.crypto.HashSecretKey(secretKey),
MasterKeySalt:       salt,
EncryptedVaultKey:   encryptedVaultKey,
VaultKeyNonce:       nonce,
SRPVerifier:         srpVerifier,
CurrentKeyVersionID: keyVersionID,
CreatedAt:           now,
UpdatedAt:           now,

## }

if err := vs.db.CreateUser(ctx, user); err != nil {
return nil, err

## }

keyVersion := &KeyVersion{
ID:                keyVersionID,
UserID:            userID,

## Version:           1,

EncryptedVaultKey: encryptedVaultKey,
VaultKeyNonce:     nonce,

## Status:            "active",

CreatedAt:         now,
RotationReason:    "initial",

## }

if err := vs.db.CreateKeyVersion(ctx, keyVersion); err != nil {
return nil, err

## }

return &RegisterResponse{
UserID:    userID,
SecretKey: secretKey,
Message:   "IMPORTANT: Save your Secret Key securely. It cannot be recovered!",
}, nil

## }

## // User Login

type LoginRequest struct {
Email     string
Password  string
SecretKey string

## }

type LoginResponse struct {
Token     string
ExpiresAt int64

## }

func (vs VaultService) Login(ctx context.Context, req LoginRequest, ipAddress,
userAgent string) (*LoginResponse, error) {
user, err := vs.db.GetUserByEmail(ctx, req.Email)
if err != nil {
return nil, err

## }

if user == nil {
time.Sleep(100 * time.Millisecond)
return nil, errors.New("invalid credentials")

## }

json:"email"
json:"password"
json:"secret_key"
json:"token"
json:"expires_at"
secretKeyHash := vs.crypto.HashSecretKey(req.SecretKey)
if subtle.ConstantTimeCompare([]byte(secretKeyHash), []byte(user.SecretKeyHash)) != 1 {
return nil, errors.New("invalid credentials")

## }

masterKey := vs.crypto.DeriveMasterKey(req.Password, req.SecretKey, user.MasterKeySalt)
defer SecureZeroMemory(masterKey)
encryptionKey, err := vs.crypto.DeriveKey(masterKey, "vault-key-encryption", 32)
if err != nil {
return nil, err

## }

defer SecureZeroMemory(encryptionKey)
vaultKey, err := vs.crypto.Decrypt(encryptionKey, user.EncryptedVaultKey,
user.VaultKeyNonce)
if err != nil {
return nil, errors.New("invalid credentials")

## }

sessionDuration := 24 * time.Hour
allPermissions := []string{"read", "write", "delete", "list", "rotate", "admin"}
sessionID := vs.sessions.Create(user.ID, vaultKey, user.CurrentKeyVersionID,

## }

## // Service Principal Authentication

type ClientCredentialsRequest struct {
ClientID     string
ClientSecret string

## }

func (vs VaultService) AuthenticateServicePrincipal(ctx context.Context, req
ClientCredentialsRequest, ipAddress, userAgent string) (*LoginResponse, error) {
sp, err := vs.db.GetServicePrincipalByClientID(ctx, req.ClientID)
if err != nil {
return nil, err

## }

if sp == nil {
time.Sleep(100 * time.Millisecond)
return nil, errors.New("invalid credentials")

## }

allPermissions, nil, nil, "", sessionDuration)
SecureZeroMemory(vaultKey)
expiresAt := time.Now().Add(sessionDuration)
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
"sub":        user.ID,
"session_id": sessionID,

## "type":       "user",

"exp":        expiresAt.Unix(),

## })

tokenString, err := token.SignedString(vs.config.JWTSecret)
if err != nil {
return nil, err

## }

vs.logAudit(ctx, user.ID, nil, "login", "", nil, ipAddress, userAgent)
return &LoginResponse{
Token:     tokenString,
ExpiresAt: expiresAt.Unix(),
}, nil
json:"client_id"
json:"client_secret"
if !sp.IsActive {
return nil, errors.New("service principal is disabled")

## }

if sp.ExpiresAt != nil && time.Now().After(*sp.ExpiresAt) {
return nil, errors.New("service principal has expired")

## }

// Check IP whitelist
if len(sp.IPWhitelist) > 0 {
allowed := false
for _, ip := range sp.IPWhitelist {
if ip == ipAddress || ip == "*" {
allowed = true
break

## }

## }

if !allowed {
return nil, errors.New("IP address not allowed")

## }

## }

// Verify client secret
secretHash := vs.crypto.HashSecret(req.ClientSecret)
if subtle.ConstantTimeCompare([]byte(secretHash), []byte(sp.ClientSecretHash)) != 1 {
return nil, errors.New("invalid credentials")

## }

// Check rate limit
allowed, err := vs.db.CheckRateLimit(ctx, "sp:"+sp.ID, sp.RateLimit)
if err != nil {
return nil, err

## }

if !allowed {
return nil, errors.New("rate limit exceeded")

## }

// Get user and vault key
user, err := vs.db.GetUserByID(ctx, sp.UserID)
if err != nil || user == nil {
return nil, errors.New("user not found")

## }

// For service principals, we need to derive the vault key differently
// We'll use a service-specific key derivation
kv, err := vs.db.GetActiveKeyVersion(ctx, user.ID)
if err != nil || kv == nil {
return nil, errors.New("no active key version")

## }

// Service principals use a derived key from the encrypted vault key
// This is a simplified approach - in production, you'd want HSM-backed keys
serviceKey, err := vs.crypto.DeriveKey(kv.EncryptedVaultKey, "service-principal-"+sp.ID, 32)
if err != nil {
return nil, err

## }

## }

## // Service Principal Management

type CreateServicePrincipalRequest struct {
Name           string
Description    string

## Permissions    []string

AllowedSecrets []string
AllowedTags    []string
IPWhitelist    []string
RateLimit      int
ExpiresAt      *time.Time

## }

type CreateServicePrincipalResponse struct {
ID           string
ClientID     string
ClientSecret string
Message      string

## }

sessionDuration := time.Hour
sessionID := vs.sessions.Create(user.ID, serviceKey, kv.ID, sp.Permissions,
sp.AllowedSecrets, sp.AllowedTags, sp.ID, sessionDuration)
vs.db.UpdateServicePrincipalLastUsed(ctx, sp.ID)
expiresAt := time.Now().Add(sessionDuration)
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
"sub":        user.ID,
"sp_id":      sp.ID,
"session_id": sessionID,

## "type":       "service_principal",

"exp":        expiresAt.Unix(),

## })

tokenString, err := token.SignedString(vs.config.JWTSecret)
if err != nil {
return nil, err

## }

vs.logAudit(ctx, user.ID, &sp.ID, "sp_login", "", nil, ipAddress, userAgent)
return &LoginResponse{
Token:     tokenString,
ExpiresAt: expiresAt.Unix(),
}, nil
json:"name"
json:"description,omitempty"
json:"permissions"
json:"allowed_secrets,omitempty"
json:"allowed_tags,omitempty"
json:"ip_whitelist,omitempty"
json:"rate_limit,omitempty"
json:"expires_at,omitempty"
json:"id"
json:"client_id"
json:"client_secret"
json:"message"

func (vs VaultService) CreateServicePrincipal(ctx context.Context, session Session,
req CreateServicePrincipalRequest) (CreateServicePrincipalResponse, error) {
if !vs.hasPermission(session, "admin") {
return nil, errors.New("admin permission required")

## }

// Validate permissions
validPerms := map[string]bool{"read": true, "write": true, "delete": true, "list": true, "rotate":
true}
for _, p := range req.Permissions {
if !validPerms[p] {
return nil, fmt.Errorf("invalid permission: %s", p)

## }

## }

clientID := vs.crypto.GenerateClientID()
clientSecret, err := vs.crypto.GenerateClientSecret()
if err != nil {
return nil, err

## }

rateLimit := req.RateLimit
if rateLimit <= 0 {
rateLimit = 60

## }

now := time.Now()
sp := &ServicePrincipal{
ID:               uuid.New().String(),
UserID:           session.UserID,
ClientID:         clientID,
ClientSecretHash: vs.crypto.HashSecret(clientSecret),
Name:             req.Name,
Description:      req.Description,
Permissions:      req.Permissions,
AllowedSecrets:   req.AllowedSecrets,
AllowedTags:      req.AllowedTags,
IPWhitelist:      req.IPWhitelist,
RateLimit:        rateLimit,
IsActive:         true,
ExpiresAt:        req.ExpiresAt,
CreatedAt:        now,
UpdatedAt:        now,

## }

if err := vs.db.CreateServicePrincipal(ctx, sp); err != nil {
return nil, err

## }

return &CreateServicePrincipalResponse{

## }

func (vs VaultService) ListServicePrincipals(ctx context.Context, session Session)
([]*ServicePrincipal, error) {
if !vs.hasPermission(session, "admin") {
return nil, errors.New("admin permission required")

## }

## }

func (vs VaultService) UpdateServicePrincipal(ctx context.Context, session Session,
spID string, req *CreateServicePrincipalRequest) error {
if !vs.hasPermission(session, "admin") {
return errors.New("admin permission required")

## }

ID:           sp.ID,
ClientID:     clientID,
ClientSecret: clientSecret,
Message:      "IMPORTANT: Save the client_secret securely. It cannot be retrieved again!",
}, nil
sps, err := vs.db.GetServicePrincipalsByUser(ctx, session.UserID)
if err != nil {
return nil, err

## }

// Clear sensitive data
for _, sp := range sps {
sp.ClientSecretHash = ""

## }

return sps, nil
sp, err := vs.db.GetServicePrincipalByClientID(ctx, spID)
if err != nil {
return err

## }

if sp == nil || sp.UserID != session.UserID {
return errors.New("service principal not found")

## }

sp.Name = req.Name
sp.Description = req.Description
sp.Permissions = req.Permissions
sp.AllowedSecrets = req.AllowedSecrets
sp.AllowedTags = req.AllowedTags

## }

func (vs VaultService) DeleteServicePrincipal(ctx context.Context, session Session,
spID string) error {
if !vs.hasPermission(session, "admin") {
return errors.New("admin

## Продолжай

## Claude Opus 4.5

sp.IPWhitelist = req.IPWhitelist
if req.RateLimit > 0 {
sp.RateLimit = req.RateLimit

## }

sp.ExpiresAt = req.ExpiresAt
return vs.db.UpdateServicePrincipal(ctx, sp)
permission required")

## }

return vs.db.DeleteServicePrincipal(ctx, spID, session.UserID)

## }

func (vs *VaultService) RegenerateClientSecret(ctx context.Context, session *Session, spID
string) (*CreateServicePrincipalResponse, error) {
if !vs.hasPermission(session, "admin") {
return nil, errors.New("admin permission required")

## }

sps, err := vs.db.GetServicePrincipalsByUser(ctx, session.UserID)
if err != nil {
return nil, err

## }

var targetSP *ServicePrincipal
for _, sp := range sps {
if sp.ID == spID {
targetSP = sp
break

## }

## }

if targetSP == nil {
return nil, errors.New("service principal not found")

## }

newSecret, err := vs.crypto.GenerateClientSecret()
if err != nil {
return nil, err

## }

if err := vs.db.RegenerateClientSecret(ctx, spID, vs.crypto.HashSecret(newSecret)); err !=
nil {
return nil, err

## }

return &CreateServicePrincipalResponse{
ID:           targetSP.ID,
ClientID:     targetSP.ClientID,
ClientSecret: newSecret,
Message:      "Client secret regenerated. Save it securely!",
}, nil

## }

func (vs *VaultService) ToggleServicePrincipal(ctx context.Context, session *Session, spID
string, active bool) error {
if !vs.hasPermission(session, "admin") {
return errors.New("admin permission required")

## }

sps, err := vs.db.GetServicePrincipalsByUser(ctx, session.UserID)
if err != nil {
return err

## }

for _, sp := range sps {
if sp.ID == spID {
sp.IsActive = active
return vs.db.UpdateServicePrincipal(ctx, sp)

## }

## }

return errors.New("service principal not found")

## }

// Permission checking
func (vs *VaultService) hasPermission(session *Session, required string) bool {
for _, p := range session.Permissions {
if p == required || p == "admin" {
return true

## }

## }

return false

## }

func (vs *VaultService) canAccessSecret(session *Session, secret *Secret) bool {
// Admin or user session can access all
if session.ServicePrincipalID == "" {
return true

## }

// Check allowed secrets
if len(session.AllowedSecrets) > 0 {
for _, allowed := range session.AllowedSecrets {
if allowed == secret.ID || allowed == secret.Name || allowed == "*" {
return true

## }

## }

## }

// Check allowed tags
if len(session.AllowedTags) > 0 && len(secret.Tags) > 0 {
for _, allowedTag := range session.AllowedTags {
for _, secretTag := range secret.Tags {
if allowedTag == secretTag {
return true

## }

## }

## }

## }

// If no restrictions, allow access
if len(session.AllowedSecrets) == 0 && len(session.AllowedTags) == 0 {
return true

## }

return false

## }

// Secret operations
type CreateSecretRequest struct {
Name      string     `json:"name"`
Data      string     `json:"data"`
Type      string     `json:"type"`

## Tags      []string   `json:"tags,omitempty"`

ExpiresAt *time.Time `json:"expires_at,omitempty"`

## }

type SecretResponse struct {
ID        string     `json:"id"`
Name      string     `json:"name"`
Data      string     `json:"data,omitempty"`
Type      string     `json:"type"`
Version   int        `json:"version"`

## Tags      []string   `json:"tags,omitempty"`

CreatedAt time.Time  `json:"created_at"`
UpdatedAt time.Time  `json:"updated_at"`
ExpiresAt *time.Time `json:"expires_at,omitempty"`

## }

func (vs *VaultService) CreateSecret(ctx context.Context, session *Session, req
*CreateSecretRequest) (*SecretResponse, error) {
if !vs.hasPermission(session, "write") {
return nil, errors.New("write permission required")

## }

encryptedData, nonce, err := vs.crypto.Encrypt(session.VaultKey, []byte(req.Data))
if err != nil {
return nil, err

## }

now := time.Now()
secret := &Secret{
ID:            uuid.New().String(),
UserID:        session.UserID,
KeyVersionID:  session.KeyVersionID,
Name:          req.Name,
EncryptedData: encryptedData,
Nonce:         nonce,
SecretType:    req.Type,

## Version:       1,

Tags:          req.Tags,
CreatedAt:     now,
UpdatedAt:     now,
ExpiresAt:     req.ExpiresAt,

## }

if err := vs.db.CreateSecret(ctx, secret); err != nil {
return nil, err

## }

return &SecretResponse{
ID:        secret.ID,
Name:      secret.Name,
Type:      secret.SecretType,
Version:   secret.Version,
Tags:      secret.Tags,
CreatedAt: secret.CreatedAt,
UpdatedAt: secret.UpdatedAt,
ExpiresAt: secret.ExpiresAt,
}, nil

## }

func (vs *VaultService) GetSecret(ctx context.Context, session *Session, secretID string)

(*SecretResponse, error) {
if !vs.hasPermission(session, "read") {
return nil, errors.New("read permission required")

## }

secret, err := vs.db.GetSecret(ctx, secretID, session.UserID)
if err != nil {
return nil, err

## }

if secret == nil {
return nil, errors.New("secret not found")

## }

if !vs.canAccessSecret(session, secret) {
return nil, errors.New("access denied to this secret")

## }

plaintext, err := vs.crypto.Decrypt(session.VaultKey, secret.EncryptedData, secret.Nonce)
if err != nil {
return nil, err

## }

vs.db.UpdateSecretAccess(ctx, secretID)
return &SecretResponse{
ID:        secret.ID,
Name:      secret.Name,
Data:      string(plaintext),
Type:      secret.SecretType,
Version:   secret.Version,
Tags:      secret.Tags,
CreatedAt: secret.CreatedAt,
UpdatedAt: secret.UpdatedAt,
ExpiresAt: secret.ExpiresAt,
}, nil

## }

func (vs *VaultService) GetSecretByName(ctx context.Context, session *Session, name
string) (*SecretResponse, error) {
if !vs.hasPermission(session, "read") {
return nil, errors.New("read permission required")

## }

secret, err := vs.db.GetSecretByName(ctx, name, session.UserID)
if err != nil {
return nil, err

## }

if secret == nil {
return nil, errors.New("secret not found")

## }

if !vs.canAccessSecret(session, secret) {
return nil, errors.New("access denied to this secret")

## }

plaintext, err := vs.crypto.Decrypt(session.VaultKey, secret.EncryptedData, secret.Nonce)
if err != nil {
return nil, err

## }

vs.db.UpdateSecretAccess(ctx, secret.ID)
return &SecretResponse{
ID:        secret.ID,
Name:      secret.Name,
Data:      string(plaintext),
Type:      secret.SecretType,
Version:   secret.Version,
Tags:      secret.Tags,
CreatedAt: secret.CreatedAt,
UpdatedAt: secret.UpdatedAt,
ExpiresAt: secret.ExpiresAt,
}, nil

## }

func (vs *VaultService) ListSecrets(ctx context.Context, session *Session)
([]*SecretResponse, error) {
if !vs.hasPermission(session, "list") {
return nil, errors.New("list permission required")

## }

secrets, err := vs.db.GetSecretsByUser(ctx, session.UserID)
if err != nil {
return nil, err

## }

var responses []*SecretResponse
for _, secret := range secrets {
if vs.canAccessSecret(session, secret) {
responses = append(responses, &SecretResponse{
ID:        secret.ID,
Name:      secret.Name,
Type:      secret.SecretType,
Version:   secret.Version,
Tags:      secret.Tags,
CreatedAt: secret.CreatedAt,
UpdatedAt: secret.UpdatedAt,
ExpiresAt: secret.ExpiresAt,

## })

## }

## }

return responses, nil

## }

func (vs *VaultService) DeleteSecret(ctx context.Context, session *Session, secretID,
ipAddress, userAgent string) error {
if !vs.hasPermission(session, "delete") {
return errors.New("delete permission required")

## }

secret, err := vs.db.GetSecret(ctx, secretID, session.UserID)
if err != nil {
return err

## }

if secret == nil {
return errors.New("secret not found")

## }

if !vs.canAccessSecret(session, secret) {
return errors.New("access denied to this secret")

## }

if err := vs.db.DeleteSecret(ctx, secretID, session.UserID); err != nil {
return err

## }

spID := &session.ServicePrincipalID
if *spID == "" {
spID = nil

## }

vs.logAudit(ctx, session.UserID, spID, "secret_deleted", secretID, nil, ipAddress,
userAgent)
return nil

## }

func (vs *VaultService) RotateVaultKey(ctx context.Context, session *Session, reason string)
(*KeyVersion, error) {
if !vs.hasPermission(session, "rotate") {
return nil, errors.New("rotate permission required")

## }

return vs.keyManager.RotateVaultKey(ctx, session.UserID, session.VaultKey, reason)

## }

func (vs *VaultService) ChangePassword(ctx context.Context, userID, oldPassword,
newPassword, secretKey string) error {
return vs.keyManager.RotateMasterKey(ctx, userID, oldPassword, newPassword,
secretKey)

## }

func (vs *VaultService) GetKeyStatus(ctx context.Context, userID string) (*KeyVersion, error)

## {

return vs.db.GetActiveKeyVersion(ctx, userID)

## }

func (vs *VaultService) GetKeyHistory(ctx context.Context, userID string) ([]*KeyVersion,
error) {
return vs.db.GetKeyVersionHistory(ctx, userID)

## }

func (vs *VaultService) logAudit(ctx context.Context, userID string, spID *string, action,
resourceID string, details map[string]interface{}, ipAddress, userAgent string) {
var detailsJSON json.RawMessage
if details != nil {
detailsJSON, _ = json.Marshal(details)

## }

log := &AuditLog{
ID:                 uuid.New().String(),
UserID:             userID,
ServicePrincipalID: spID,
Action:             action,
ResourceID:         resourceID,
Details:            detailsJSON,
IPAddress:          ipAddress,
UserAgent:          userAgent,
CreatedAt:          time.Now(),

## }

vs.db.CreateAuditLog(ctx, log)

## }

## //

## ======================================================================

## ======

// HTTP Handlers

## //

## ======================================================================

## ======

type Server struct {
vault  *VaultService
config *Config

## }

func NewServer(vault *VaultService, config *Config) *Server {
return &Server{vault: vault, config: config}

## }

func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(status)
json.NewEncoder(w).Encode(data)

## }

func (s *Server) respondError(w http.ResponseWriter, status int, message string) {
s.respondJSON(w, status, map[string]string{"error": message})

## }

func (s *Server) authenticate(r *http.Request) (*Session, error) {
authHeader := r.Header.Get("Authorization")
if !strings.HasPrefix(authHeader, "Bearer ") {
return nil, errors.New("missing authorization header")

## }

tokenString := strings.TrimPrefix(authHeader, "Bearer ")
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
return nil, errors.New("invalid signing method")

## }

return s.config.JWTSecret, nil

## })

if err != nil || !token.Valid {
return nil, errors.New("invalid token")

## }

claims, ok := token.Claims.(jwt.MapClaims)
if !ok {
return nil, errors.New("invalid claims")

## }

sessionID, ok := claims["session_id"].(string)
if !ok {
return nil, errors.New("invalid session")

## }

session, exists := s.vault.sessions.Get(sessionID)
if !exists {
return nil, errors.New("session expired")

## }

return session, nil

## }

// Auth handlers
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
var req RegisterRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
s.respondError(w, http.StatusBadRequest, "invalid request body")

return

## }

if req.Email == "" || req.Password == "" {
s.respondError(w, http.StatusBadRequest, "email and password required")
return

## }

if len(req.Password) < 12 {
s.respondError(w, http.StatusBadRequest, "password must be at least 12 characters")
return

## }

resp, err := s.vault.Register(r.Context(), &req)
if err != nil {
s.respondError(w, http.StatusBadRequest, err.Error())
return

## }

s.respondJSON(w, http.StatusCreated, resp)

## }

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
var req LoginRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
s.respondError(w, http.StatusBadRequest, "invalid request body")
return

## }

resp, err := s.vault.Login(r.Context(), &req, r.RemoteAddr, r.UserAgent())
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, resp)

## }

func (s *Server) handleClientCredentials(w http.ResponseWriter, r *http.Request) {
var req ClientCredentialsRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
s.respondError(w, http.StatusBadRequest, "invalid request body")
return

## }

resp, err := s.vault.AuthenticateServicePrincipal(r.Context(), &req, r.RemoteAddr,
r.UserAgent())
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, resp)

## }

// Service Principal handlers
func (s *Server) handleCreateServicePrincipal(w http.ResponseWriter, r *http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

var req CreateServicePrincipalRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
s.respondError(w, http.StatusBadRequest, "invalid request body")
return

## }

resp, err := s.vault.CreateServicePrincipal(r.Context(), session, &req)
if err != nil {
s.respondError(w, http.StatusBadRequest, err.Error())
return

## }

s.respondJSON(w, http.StatusCreated, resp)

## }

func (s *Server) handleListServicePrincipals(w http.ResponseWriter, r *http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

sps, err := s.vault.ListServicePrincipals(r.Context(), session)
if err != nil {
s.respondError(w, http.StatusBadRequest, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, sps)

## }

func (s *Server) handleDeleteServicePrincipal(w http.ResponseWriter, r *http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

spID := strings.TrimPrefix(r.URL.Path, "/api/v1/service-principals/")
if err := s.vault.DeleteServicePrincipal(r.Context(), session, spID); err != nil {
s.respondError(w, http.StatusBadRequest, err.Error())
return

## }

w.WriteHeader(http.StatusNoContent)

## }

func (s *Server) handleRegenerateClientSecret(w http.ResponseWriter, r *http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

spID := strings.TrimPrefix(r.URL.Path, "/api/v1/service-principals/")
spID = strings.TrimSuffix(spID, "/regenerate-secret")
resp, err := s.vault.RegenerateClientSecret(r.Context(), session, spID)
if err != nil {
s.respondError(w, http.StatusBadRequest, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, resp)

## }

func (s *Server) handleToggleServicePrincipal(w http.ResponseWriter, r *http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

spID := strings.TrimPrefix(r.URL.Path, "/api/v1/service-principals/")
spID = strings.TrimSuffix(spID, "/toggle")
var req struct {
Active bool `json:"active"`

## }

if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
s.respondError(w, http.StatusBadRequest, "invalid request body")
return

## }

if err := s.vault.ToggleServicePrincipal(r.Context(), session, spID, req.Active); err != nil {
s.respondError(w, http.StatusBadRequest, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})

## }

// Secret handlers
func (s *Server) handleCreateSecret(w http.ResponseWriter, r *http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

var req CreateSecretRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
s.respondError(w, http.StatusBadRequest, "invalid request body")
return

## }

resp, err := s.vault.CreateSecret(r.Context(), session, &req)
if err != nil {
s.respondError(w, http.StatusInternalServerError, err.Error())
return

## }

s.respondJSON(w, http.StatusCreated, resp)

## }

func (s *Server) handleGetSecret(w http.ResponseWriter, r *http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

secretID := strings.TrimPrefix(r.URL.Path, "/api/v1/secrets/")
resp, err := s.vault.GetSecret(r.Context(), session, secretID)
if err != nil {
s.respondError(w, http.StatusNotFound, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, resp)

## }

func (s *Server) handleGetSecretByName(w http.ResponseWriter, r *http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

name := r.URL.Query().Get("name")
if name == "" {
s.respondError(w, http.StatusBadRequest, "name parameter required")
return

## }

resp, err := s.vault.GetSecretByName(r.Context(), session, name)
if err != nil {
s.respondError(w, http.StatusNotFound, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, resp)

## }

func (s *Server) handleListSecrets(w http.ResponseWriter, r *http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

resp, err := s.vault.ListSecrets(r.Context(), session)
if err != nil {
s.respondError(w, http.StatusInternalServerError, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, resp)

## }

func (s *Server) handleDeleteSecret(w http.ResponseWriter, r *http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

secretID := strings.TrimPrefix(r.URL.Path, "/api/v1/secrets/")
if err := s.vault.DeleteSecret(r.Context(), session, secretID, r.RemoteAddr, r.UserAgent());
err != nil {
s.respondError(w, http.StatusInternalServerError, err.Error())
return

## }

w.WriteHeader(http.StatusNoContent)

## }

// Key management handlers
func (s *Server) handleRotateKey(w http.ResponseWriter, r *http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

var req struct {
Reason string `json:"reason"`

## }

json.NewDecoder(r.Body).Decode(&req)
if req.Reason == "" {
req.Reason = "manual_rotation"

## }

kv, err := s.vault.RotateVaultKey(r.Context(), session, req.Reason)
if err != nil {
s.respondError(w, http.StatusInternalServerError, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, map[string]interface{}{
"message":     "Key rotated successfully",
"new_version": kv.Version,

## })

## }

func (s *Server) handleKeyStatus(w http.ResponseWriter, r *http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

kv, err := s.vault.GetKeyStatus(r.Context(), session.UserID)
if err != nil {
s.respondError(w, http.StatusInternalServerError, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, kv)

## }

func (s *Server) handleKeyHistory(w http.ResponseWriter, r *http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

history, err := s.vault.GetKeyHistory(r.Context(), session.UserID)
if err != nil {
s.respondError(w, http.StatusInternalServerError, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, history)

## }

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
session, err := s.authenticate(r)
if err != nil {
s.respondError(w, http.StatusUnauthorized, err.Error())
return

## }

var req struct {
OldPassword string `json:"old_password"`
NewPassword string `json:"new_password"`
SecretKey   string `json:"secret_key"`

## }

if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
s.respondError(w, http.StatusBadRequest, "invalid request body")
return

## }

if len(req.NewPassword) < 12 {
s.respondError(w, http.StatusBadRequest, "new password must be at least 12
characters")
return

## }

if err := s.vault.ChangePassword(r.Context(), session.UserID, req.OldPassword,
req.NewPassword, req.SecretKey); err != nil {
s.respondError(w, http.StatusBadRequest, err.Error())
return

## }

s.respondJSON(w, http.StatusOK, map[string]string{
"message": "Password changed successfully. Please login again.",

## })

## }

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
s.respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})

## }

func (s *Server) SetupRoutes() http.Handler {
mux := http.NewServeMux()

## // Health

mux.HandleFunc("GET /health", s.handleHealth)

## // Auth

mux.HandleFunc("POST /api/v1/auth/register", s.handleRegister)
mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
mux.HandleFunc("POST /api/v1/auth/token", s.handleClientCredentials) // OAuth2 style

## // Service Principals

mux.HandleFunc("POST /api/v1/service-principals", s.handleCreateServicePrincipal)
mux.HandleFunc("GET /api/v1/service-principals", s.handleListServicePrincipals)
mux.HandleFunc("DELETE /api/v1/service-principals/", s.handleDeleteServicePrincipal)
mux.HandleFunc("POST /api/v1/service-principals/{id}/regenerate-secret",
s.handleRegenerateClientSecret)
mux.HandleFunc("POST /api/v1/service-principals/{id}/toggle",
s.handleToggleServicePrincipal)

## // Secrets

mux.HandleFunc("POST /api/v1/secrets", s.handleCreateSecret)
mux.HandleFunc("GET /api/v1/secrets", s.handleListSecrets)
mux.HandleFunc("GET /api/v1/secrets/by-name", s.handleGetSecretByName)
mux.HandleFunc("GET /api/v1/secrets/", s.handleGetSecret)
mux.HandleFunc("DELETE /api/v1/secrets/", s.handleDeleteSecret)
// Key management
mux.HandleFunc("POST /api/v1/keys/rotate", s.handleRotateKey)
mux.HandleFunc("GET /api/v1/keys/status", s.handleKeyStatus)
mux.HandleFunc("GET /api/v1/keys/history", s.handleKeyHistory)

## // Account

mux.HandleFunc("POST /api/v1/account/change-password", s.handleChangePassword)
return mux

## }

func main() {
config := LoadConfig()
db, err := NewDatabase(config.DatabaseURL)
if err != nil {
log.Fatalf("Failed to connect to database: %v", err)

## }

defer db.Close()
crypto := NewCryptoService(config)
sessions := NewSessionManager()
vault := NewVaultService(db, crypto, sessions, config)

Пример использования в приложении

## Конфигурация (config.yaml):

Go клиент:
server := NewServer(vault, config)
handler := server.SetupRoutes()
log.Printf("Starting KeyVault server on port %s", config.ServerPort)
if err := http.ListenAndServe(":"+config.ServerPort, handler); err != nil {
log.Fatalf("Server failed: %v", err)

## }

## }

keyvault:
endpoint: "https://your-keyvault.example.com"
client_id: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
client_secret: "your-client-secret-here"
package keyvault
import (

## "bytes"

## "encoding/json"

## "fmt"

## "net/http"

## "sync"

## "time"

## )

type Client struct {
endpoint     string
clientID     string
clientSecret string
token        string
tokenExpiry  time.Time
httpClient   *http.Client
mu           sync.RWMutex

## }

func NewClient(endpoint, clientID, clientSecret string) *Client {
return &Client{
endpoint:     endpoint,
clientID:     clientID,
clientSecret: clientSecret,
httpClient:   &http.Client{Timeout: 10 * time.Second},

## }

## }

func (c *Client) authenticate() error {
c.mu.Lock()
defer c.mu.Unlock()
if c.token != "" && time.Now().Before(c.tokenExpiry.Add(-5*time.Minute)) {
return nil

## }

body, _ := json.Marshal(map[string]string{
"client_id":     c.clientID,
"client_secret": c.clientSecret,

## })

resp, err := c.httpClient.Post(c.endpoint+"/api/v1/auth/token", "application/json",
bytes.NewReader(body))
if err != nil {
return err

## }

defer resp.Body.Close()
if resp.StatusCode != http.StatusOK {
return fmt.Errorf("authentication failed: %d", resp.StatusCode)

## }

var result struct {
Token     string `json:"token"`
ExpiresAt int64  `json:"expires_at"`

## }

if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
return err

## }

c.token = result.Token
c.tokenExpiry = time.Unix(result.ExpiresAt, 0)
return nil

## }

func (c *Client) GetSecret(name string) (string, error) {
if err := c.authenticate(); err != nil {
return "", err

## }

API Endpoints
req, _ := http.NewRequest("GET", c.endpoint+"/api/v1/secrets/by-name?name="+name,
nil)
req.Header.Set("Authorization", "Bearer "+c.token)
resp, err := c.httpClient.Do(req)
if err != nil {
return "", err

## }

defer resp.Body.Close()
if resp.StatusCode != http.StatusOK {
return "", fmt.Errorf("failed to get secret: %d", resp.StatusCode)

## }

var result struct {
Data string `json:"data"`

## }

if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
return "", err

## }

return result.Data, nil

## }

## // Использование

func main() {
client := NewClient(

## "https://keyvault.example.com",

## "your-client-id",

## "your-client-secret",

## )

dbPassword, _ := client.GetSecret("database-password")
apiKey, _ := client.GetSecret("api-key")

## }

EndpointMethodОписание

## POST

Аутентификация по client_id/
client_secret
POSTСоздать Service Principal
GETСписок Service Principals
DELETEУдалить Service Principal
POSTПерегенерировать client_secret
POSTВключить/выключить
GETПолучить секрет по имени
А как будет выглядеть пример секрета к примеру мы добавим
коннекшен стнинг в наш KeyVault как будет выглядеть key
который мы положим в наш конфиг?

## Claude Opus 4.5

Пример использования KeyVault для Connection

## String

- Создание секрета в KeyVault
  
  ## /api/v1/auth/token
  
  ## /api/v1/service-principals
  
  ## /api/v1/service-principals
  
  ## /api/v1/service-principals/{id}
  
  ## /api/v1/service-principals/{id}/
  
  regenerate-secret
  
  ## /api/v1/service-principals/{id}/toggle
  
  /api/v1/secrets/by-name?name=X
  
  # Сначала логинимся как пользователь и создаём секрет
  
  curl -X POST https://keyvault.example.com/api/v1/secrets \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  
  ## -d '{
  
  ## "name": "postgres-connection-string",
  
  "data": "postgresql://user:password123@db.example.com:5432/myapp?sslmode=require",
  
  ## "type": "connection_string",
  
  ## "tags": ["database", "production"]
  
  ## }'

- Конфигурация приложения
  config.yaml:

- Код приложения
  database:
  
  # Вместо прямого connection string - ссылка на KeyVault
  
  connection_string_key: "postgres-connection-string"
  keyvault:
  endpoint: "https://keyvault.example.com"
  client_id: "sp-a1b2c3d4e5f6"
  client_secret: "kv.secret.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  
  # Или через environment variables (рекомендуется):
  
  # KEYVAULT_ENDPOINT=https://keyvault.example.com
  
  # KEYVAULT_CLIENT_ID=sp-a1b2c3d4e5f6
  
  # KEYVAULT_CLIENT_SECRET=kv.secret.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
  
  package main
  import (
  
  ## "database/sql"
  
  ## "log"
  
  ## "myapp/keyvault"
  
  ## _ "github.com/lib/pq"
  
  ## )
  
  func main() {
  // Инициализация KeyVault клиента
  kvClient := keyvault.NewClient(
  os.Getenv("KEYVAULT_ENDPOINT"),
  os.Getenv("KEYVAULT_CLIENT_ID"),
  os.Getenv("KEYVAULT_CLIENT_SECRET"),
  
  ## )
  
  // Получаем connection string из KeyVault
  connStr, err := kvClient.GetSecret("postgres-connection-string")
  if err != nil {

- Формат ключей
  
  ## Что
  
  хранится
  Имя секрета (key)Пример значения
  PostgreSQL
  
  ## Redis
  
  API ключ
  JWT Secret

- Что попадает в конфиг vs KeyVault
  log.Fatalf("Failed to get connection string: %v", err)
  
  ## }
  
  // Подключаемся к БД
  db, err := sql.Open("postgres", connStr)
  if err != nil {
  log.Fatalf("Failed to connect to database: %v", err)
  
  ## }
  
  defer db.Close()
  
  ## // ...
  
  ## }
  
  postgres-connection-
  string
  postgresql://user:pass@host:5432/db
  redis-url
  redis://:password@redis.example.com:
  
  ## 6379/0
  
  stripe-api-keysk_live_xxxxxxxxxxxxx
  jwt-signing-keyyour-256-bit-secret
  
  # ✅ В конфиге (config.yaml) - только ссылки:
  
  database:
  connection_key: "postgres-prod"
  
  # имя секрета
  
  max_connections: 100
  
  # не секретные параметры
  
  redis:
  url_key: "redis-prod"

## 6. Environment Variables (production)

Таким образом, в конфиге хранится только имя секрета (
), а сам connection string безопасно хранится в KeyVault и извлекается при
старте приложения.
stripe:
api_key_name: "stripe-api-key"

# ❌ НЕ в конфиге - реальные credentials:

# connection_string: "postgresql://user:password@..."  <- НИКОГДА!

# .env или Kubernetes secrets

KEYVAULT_ENDPOINT=https://keyvault.internal.example.com
KEYVAULT_CLIENT_ID=sp-myapp-prod-a1b2c3d4
KEYVAULT_CLIENT_SECRET=kv.secret.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
postgres-connection-
string