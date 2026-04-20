package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAcceptsTomlInCompatibilityMode(t *testing.T) {
	t.Setenv("BACKEND_REBUILD_CONFIG_COMPAT_MODE", "true")
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(`
[server]
addr = "127.0.0.1:19082"

[mysql]
dsn = "root:pass@tcp(localhost:3306)/feasoj"

[jwt]
secret = "toml-secret"

[avatar_storage]
dir = "/tmp/avatars"
max_bytes = 12345
`), 0644); err != nil {
		t.Fatalf("write toml config failed: %v", err)
	}
	t.Setenv("BACKEND_REBUILD_CONFIG", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !strings.HasPrefix(cfg.ConfigSource, "toml-compat:") {
		t.Fatalf("expected toml-compat source, got %s", cfg.ConfigSource)
	}
	if cfg.Addr != "127.0.0.1:19082" {
		t.Fatalf("unexpected addr: %s", cfg.Addr)
	}
	if cfg.JWTSecret != "toml-secret" {
		t.Fatalf("unexpected jwt secret: %s", cfg.JWTSecret)
	}
	if cfg.AvatarUploadDir != "/tmp/avatars" {
		t.Fatalf("unexpected avatar dir: %s", cfg.AvatarUploadDir)
	}
	if cfg.AvatarUploadMaxBytes != 12345 {
		t.Fatalf("unexpected avatar max bytes: %d", cfg.AvatarUploadMaxBytes)
	}
}

func TestLoadRejectsTomlWhenCompatibilityDisabled(t *testing.T) {
	t.Setenv("BACKEND_REBUILD_CONFIG_COMPAT_MODE", "false")
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("[server]\naddr='127.0.0.1:19082'\n"), 0644); err != nil {
		t.Fatalf("write toml config failed: %v", err)
	}
	t.Setenv("BACKEND_REBUILD_CONFIG", path)

	_, err := Load()
	if err == nil {
		t.Fatal("expected Load to fail when compatibility mode is disabled")
	}
	if !strings.Contains(err.Error(), "BACKEND_REBUILD_CONFIG_COMPAT_MODE=false") {
		t.Fatalf("unexpected error: %v", err)
	}
}
