package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-sdk/core/config"
)

func TestLoadSettingsDecodesConfig(t *testing.T) {
	setTestConfig(t, `
app:
  name: " example "
database:
  driver: " sqlite "
  dsn: " test.db "
server:
  address: " :9000 "
  reflection: true
auth:
  jwt_secret: secret
`)
	value, err := loadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if value.App.Name != "example" || value.Database.Driver != "sqlite" || value.Database.DSN != "test.db" {
		t.Fatalf("unexpected settings: %#v", value)
	}
	if value.Server.Address != ":9000" || !value.Server.Reflection || value.Auth.JWTSecret != "secret" {
		t.Fatalf("unexpected settings: %#v", value)
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
