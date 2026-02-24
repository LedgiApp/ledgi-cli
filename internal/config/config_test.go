package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	// Use a temp dir instead of real home.
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	cfg := &Config{APIKey: "ledgi_sk_test123", APIURL: "http://localhost:8000"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists with correct permissions.
	path := filepath.Join(tmp, ".ledgi", "config.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("config file not found: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected perms 0600, got %o", info.Mode().Perm())
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.APIKey != cfg.APIKey {
		t.Errorf("APIKey = %q, want %q", loaded.APIKey, cfg.APIKey)
	}
	if loaded.APIURL != cfg.APIURL {
		t.Errorf("APIURL = %q, want %q", loaded.APIURL, cfg.APIURL)
	}
}

func TestLoadMissing(t *testing.T) {
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should not fail for missing file: %v", err)
	}
	if cfg.APIKey != "" {
		t.Errorf("expected empty APIKey, got %q", cfg.APIKey)
	}
}

func TestResolvePriority(t *testing.T) {
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	// Save a config file.
	Save(&Config{APIKey: "ledgi_sk_fromfile", APIURL: "http://file-url"})

	// Config file should be used when no flag or env.
	os.Unsetenv("LEDGI_API_KEY")
	os.Unsetenv("LEDGI_API_URL")
	key, url := Resolve("", "")
	if key != "ledgi_sk_fromfile" {
		t.Errorf("expected file key, got %q", key)
	}
	if url != "http://file-url" {
		t.Errorf("expected file url, got %q", url)
	}

	// Env overrides file.
	os.Setenv("LEDGI_API_KEY", "ledgi_sk_fromenv")
	os.Setenv("LEDGI_API_URL", "http://env-url")
	defer os.Unsetenv("LEDGI_API_KEY")
	defer os.Unsetenv("LEDGI_API_URL")

	key, url = Resolve("", "")
	if key != "ledgi_sk_fromenv" {
		t.Errorf("expected env key, got %q", key)
	}
	if url != "http://env-url" {
		t.Errorf("expected env url, got %q", url)
	}

	// Flag overrides env.
	key, url = Resolve("ledgi_sk_fromflag", "http://flag-url")
	if key != "ledgi_sk_fromflag" {
		t.Errorf("expected flag key, got %q", key)
	}
	if url != "http://flag-url" {
		t.Errorf("expected flag url, got %q", url)
	}
}

func TestResolveDefaultURL(t *testing.T) {
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)
	os.Unsetenv("LEDGI_API_KEY")
	os.Unsetenv("LEDGI_API_URL")

	_, url := Resolve("", "")
	if url != DefaultAPIURL {
		t.Errorf("expected default URL %q, got %q", DefaultAPIURL, url)
	}
}
