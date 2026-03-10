package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromEnvLoadsRootDotEnv(t *testing.T) {
	tempRoot := t.TempDir()
	serverDir := filepath.Join(tempRoot, "server")
	if err := os.MkdirAll(serverDir, 0o755); err != nil {
		t.Fatalf("mkdir server dir: %v", err)
	}

	dotEnv := filepath.Join(tempRoot, ".env")
	if err := os.WriteFile(dotEnv, []byte("INTERVIEW_CRAWLER_MYSQL_DSN=file-dsn\nINTERVIEW_CRAWLER_HTTP_ADDR=:9090\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	restoreCwd := chdirForTest(t, serverDir)
	defer restoreCwd()
	unsetEnvForTest(t, "INTERVIEW_CRAWLER_MYSQL_DSN")
	unsetEnvForTest(t, "INTERVIEW_CRAWLER_HTTP_ADDR")

	cfg := LoadFromEnv()

	if cfg.MySQLDSN != "file-dsn" {
		t.Fatalf("expected MySQL DSN from .env, got %q", cfg.MySQLDSN)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("expected HTTP address from .env, got %q", cfg.HTTPAddr)
	}
}

func TestLoadFromEnvDoesNotOverrideExistingEnv(t *testing.T) {
	tempRoot := t.TempDir()
	serverDir := filepath.Join(tempRoot, "server")
	if err := os.MkdirAll(serverDir, 0o755); err != nil {
		t.Fatalf("mkdir server dir: %v", err)
	}

	dotEnv := filepath.Join(tempRoot, ".env")
	if err := os.WriteFile(dotEnv, []byte("INTERVIEW_CRAWLER_MYSQL_DSN=file-dsn\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	restoreCwd := chdirForTest(t, serverDir)
	defer restoreCwd()
	setEnvForTest(t, "INTERVIEW_CRAWLER_MYSQL_DSN", "process-dsn")

	cfg := LoadFromEnv()

	if cfg.MySQLDSN != "process-dsn" {
		t.Fatalf("expected process env to win, got %q", cfg.MySQLDSN)
	}
}

func chdirForTest(t *testing.T, dir string) func() {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	return func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}
}

func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()
	value, exists := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset env %s: %v", key, err)
	}
	t.Cleanup(func() {
		if exists {
			_ = os.Setenv(key, value)
			return
		}
		_ = os.Unsetenv(key)
	})
}

func setEnvForTest(t *testing.T, key, value string) {
	t.Helper()
	previous, exists := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("set env %s: %v", key, err)
	}
	t.Cleanup(func() {
		if exists {
			_ = os.Setenv(key, previous)
			return
		}
		_ = os.Unsetenv(key)
	})
}
