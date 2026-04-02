package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	originalHost := os.Getenv("ZENODB_HOST")
	originalPort := os.Getenv("ZENODB_PORT")
	originalName := os.Getenv("ZENODB_NAME")
	os.Unsetenv("ZENODB_HOST")
	os.Unsetenv("ZENODB_PORT")
	os.Unsetenv("ZENODB_NAME")
	defer func() {
		if originalHost != "" {
			os.Setenv("ZENODB_HOST", originalHost)
		}
		if originalPort != "" {
			os.Setenv("ZENODB_PORT", originalPort)
		}
		if originalName != "" {
			os.Setenv("ZENODB_NAME", originalName)
		}
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.DBHost != "127.0.0.1" {
		t.Errorf("expected default DBHost 127.0.0.1, got %s", cfg.DBHost)
	}
	if cfg.DBPort != "3306" {
		t.Errorf("expected default DBPort 3306, got %s", cfg.DBPort)
	}
	if cfg.DBName != "zeno" {
		t.Errorf("expected default DBName zeno, got %s", cfg.DBName)
	}
	if cfg.Theme != "catppuccin-macchiato" {
		t.Errorf("expected default Theme catppuccin-macchiato, got %s", cfg.Theme)
	}
}

func TestEnvVarOverride(t *testing.T) {
	os.Setenv("ZENODB_HOST", "customhost")
	os.Setenv("ZENODB_PORT", "5432")
	os.Setenv("ZENODB_NAME", "customdb")
	os.Setenv("ZENODB_USER", "customuser")
	os.Setenv("ZENODB_PASS", "custompass")
	defer func() {
		os.Unsetenv("ZENODB_HOST")
		os.Unsetenv("ZENODB_PORT")
		os.Unsetenv("ZENODB_NAME")
		os.Unsetenv("ZENODB_USER")
		os.Unsetenv("ZENODB_PASS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.DBHost != "customhost" {
		t.Errorf("expected DBHost customhost, got %s", cfg.DBHost)
	}
	if cfg.DBPort != "5432" {
		t.Errorf("expected DBPort 5432, got %s", cfg.DBPort)
	}
	if cfg.DBName != "customdb" {
		t.Errorf("expected DBName customdb, got %s", cfg.DBName)
	}
	if cfg.DBUser != "customuser" {
		t.Errorf("expected DBUser customuser, got %s", cfg.DBUser)
	}
	if cfg.DBPass != "custompass" {
		t.Errorf("expected DBPass custompass, got %s", cfg.DBPass)
	}
}

func TestMissingCredentials(t *testing.T) {
	os.Unsetenv("ZENODB_USER")
	os.Unsetenv("ZENODB_PASS")

	cfgDir := filepath.Join(os.TempDir(), "zeno-test-noenv")
	os.MkdirAll(cfgDir, 0755)
	defer os.RemoveAll(cfgDir)

	_, err := Load()
	if err == nil {
		t.Error("expected error for missing credentials, got nil")
	}
}

func TestConfigYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zeno-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "zeno")
	os.MkdirAll(configDir, 0755)

	yamlContent := `database:
  host: yamldb
  port: "9999"
  name: yamldbname
ui:
  theme: monokai
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	envContent := `ZENODB_USER=yamluser
ZENODB_PASS=yamlpass`
	if err := os.WriteFile(filepath.Join(configDir, ".env"), []byte(envContent), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", tmpDir)

	os.Unsetenv("ZENODB_HOST")
	os.Unsetenv("ZENODB_PORT")
	os.Unsetenv("ZENODB_NAME")
	os.Unsetenv("ZENODB_USER")
	os.Unsetenv("ZENODB_PASS")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.DBHost != "yamldb" {
		t.Errorf("expected DBHost yamldb, got %s", cfg.DBHost)
	}
	if cfg.DBPort != "9999" {
		t.Errorf("expected DBPort 9999, got %s", cfg.DBPort)
	}
	if cfg.DBName != "yamldbname" {
		t.Errorf("expected DBName yamldbname, got %s", cfg.DBName)
	}
	if cfg.Theme != "monokai" {
		t.Errorf("expected Theme monokai, got %s", cfg.Theme)
	}
	if cfg.DBUser != "yamluser" {
		t.Errorf("expected DBUser yamluser, got %s", cfg.DBUser)
	}
	if cfg.DBPass != "yamlpass" {
		t.Errorf("expected DBPass yamlpass, got %s", cfg.DBPass)
	}
}
