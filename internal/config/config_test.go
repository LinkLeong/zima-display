package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCreatesConfigForRequestedDataDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), ConfigName)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DataDir != filepath.Dir(path) {
		t.Fatalf("data directory = %q, want %q", cfg.DataDir, filepath.Dir(path))
	}
}

func TestValidateRejectsUnsafeMediaRoot(t *testing.T) {
	cfg := Default()
	cfg.MediaRoots = []string{"/etc"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected unsafe media root to fail validation")
	}
}

func TestStoreReturnsIndependentCopy(t *testing.T) {
	path := filepath.Join(t.TempDir(), ConfigName)
	store, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	first := store.Get()
	first.MediaRoots[0] = "/media/changed"
	second := store.Get()
	if second.MediaRoots[0] == first.MediaRoots[0] {
		t.Fatal("store returned a mutable media roots slice")
	}
}

func TestLoadOlderConfigDefaultsDashboardLanguage(t *testing.T) {
	path := filepath.Join(t.TempDir(), ConfigName)
	data := `{
  "data_dir": "/DATA/AppData/zima-display",
  "media_roots": ["/DATA"],
  "default_mode": "dashboard",
  "auto_start_display": true,
  "volume": 70,
  "renderer": {"backend": "drm", "resolution": "1920x1080", "audio_device": "auto"},
  "dashboard": {"title": "Zima Display", "refresh_interval_ms": 1000}
}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Dashboard.Language != "zh-CN" {
		t.Fatalf("dashboard language = %q, want zh-CN", cfg.Dashboard.Language)
	}
}

func TestValidateRejectsUnsupportedDashboardLanguage(t *testing.T) {
	cfg := Default()
	cfg.Dashboard.Language = "fr-FR"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected unsupported dashboard language to fail validation")
	}
}
