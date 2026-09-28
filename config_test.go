package app

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/go-sdk/core/config"
	"github.com/go-sdk/database/rdx"
)

func TestLoadSettingsDecodesConfig(t *testing.T) {
	setTestConfig(t, `
app:
  name: " example "
database:
  driver: " sqlite "
  dsn: " test.db "
redis:
  enabled: true
  mode: sentinel
  addresses:
    - 127.0.0.1:26379
    - 127.0.0.2:26379
  master_name: " redis-master "
  username: redis-user
  password: redis-password
  sentinel_username: sentinel-user
  sentinel_password: sentinel-password
  database: 2
  client_name: " example-client "
  dial_timeout: 3s
  read_timeout: 4s
  write_timeout: 5s
  pool_timeout: 6s
  pool_size: 20
  min_idle_conns: 2
  max_idle_conns: 10
  max_active_conns: 30
  conn_max_idle_time: 7m
  conn_max_lifetime: 8m
  tls:
    enabled: true
    server_name: redis.example.com
server:
  address: " :9000 "
  reflection: true
auth:
  jwt_secret: secret
  jwt_private_key: private-key
`)
	value, err := loadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if value.App.Name != "example" || value.Database.Driver != "sqlite" || value.Database.DSN != "test.db" {
		t.Fatalf("unexpected settings: %#v", value)
	}
	if value.Server.Address != ":9000" || !value.Server.Reflection ||
		value.Auth.JWTSecret != "secret" || value.Auth.JWTPrivateKey != "private-key" {
		t.Fatalf("unexpected settings: %#v", value)
	}
	expectedRedis := rdx.Config{
		Mode:             rdx.ModeSentinel,
		Addresses:        []string{"127.0.0.1:26379", "127.0.0.2:26379"},
		MasterName:       "redis-master",
		Username:         "redis-user",
		Password:         "redis-password",
		SentinelUsername: "sentinel-user",
		SentinelPassword: "sentinel-password",
		Database:         2,
		ClientName:       "example-client",
		DialTimeout:      3 * time.Second,
		ReadTimeout:      4 * time.Second,
		WriteTimeout:     5 * time.Second,
		PoolTimeout:      6 * time.Second,
		PoolSize:         20,
		MinIdleConns:     2,
		MaxIdleConns:     10,
		MaxActiveConns:   30,
		ConnMaxIdleTime:  7 * time.Minute,
		ConnMaxLifetime:  8 * time.Minute,
		TLS: rdx.TLSConfig{
			Enabled:    true,
			ServerName: "redis.example.com",
		},
	}
	if !value.Redis.Enabled || !reflect.DeepEqual(value.Redis.Config, expectedRedis) {
		t.Fatal("unexpected redis settings")
	}
}

func TestLoadSettingsKeepsDefaults(t *testing.T) {
	setTestConfig(t, `
database:
  dsn: test.db
`)
	value, err := loadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if value.App.Name == "" || value.Database.Driver != "postgres" || value.Server.Address != ":8080" {
		t.Fatalf("unexpected default settings: %#v", value)
	}
}

func TestServerOptionsUsesJWTPrivateKey(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	value := settings{
		App:    appSettings{Name: "example"},
		Server: serverSettings{Address: ":8080"},
		Auth: authSettings{
			JWTSecret:     "secret",
			JWTPrivateKey: string(privateKeyPEM),
		},
	}
	options, err := serverOptions(value, registrationSnapshot{})
	if err != nil {
		t.Fatal(err)
	}
	if len(options) != 3 {
		t.Fatalf("unexpected server option count: %d", len(options))
	}
}

func TestServerOptionsDoesNotFallBackFromInvalidJWTPrivateKey(t *testing.T) {
	value := settings{
		App:    appSettings{Name: "example"},
		Server: serverSettings{Address: ":8080"},
		Auth: authSettings{
			JWTSecret:     "secret",
			JWTPrivateKey: "invalid",
		},
	}
	if _, err := serverOptions(value, registrationSnapshot{}); err == nil {
		t.Fatal("expected invalid jwt private key error")
	}
}

func setTestConfig(t *testing.T, content string) {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(filename, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	value := config.New(config.WithFile(filename))
	if err := value.Load(); err != nil {
		t.Fatal(err)
	}
	config.SetDefault(value)
}
