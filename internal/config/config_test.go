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
	first.Canvas.Widgets[0].X = 999
	if second.Canvas.Widgets[0].X == first.Canvas.Widgets[0].X {
		t.Fatal("store returned a mutable canvas widget slice")
	}
}

func TestStoreGeneratesPersistentAutomationToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), ConfigName)
	store, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	first := store.Get().Automation.Token
	if len(first) < 24 {
		t.Fatalf("automation token is unexpectedly short: %q", first)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config permissions = %o, want 600", info.Mode().Perm())
	}
	reloaded, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if second := reloaded.Get().Automation.Token; second != first {
		t.Fatalf("automation token changed after reload: %q != %q", second, first)
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
	if cfg.Dashboard.Layout != "auto" || cfg.Dashboard.FontScale != 1 {
		t.Fatalf("dashboard defaults = layout %q scale %.2f, want auto and 1", cfg.Dashboard.Layout, cfg.Dashboard.FontScale)
	}
}

func TestValidateRejectsUnsupportedDashboardLanguage(t *testing.T) {
	cfg := Default()
	cfg.Dashboard.Language = "fr-FR"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected unsupported dashboard language to fail validation")
	}
}

func TestValidateRejectsInvalidDashboardAppearance(t *testing.T) {
	for _, mutate := range []func(*Config){
		func(cfg *Config) { cfg.Dashboard.Layout = "floating" },
		func(cfg *Config) { cfg.Dashboard.FontScale = 2.1 },
		func(cfg *Config) { cfg.Dashboard.FontScale = 0.7 },
	} {
		cfg := Default()
		mutate(&cfg)
		if err := cfg.Validate(); err == nil {
			t.Fatal("expected invalid dashboard appearance to fail validation")
		}
	}
}

func TestValidateRejectsInvalidCanvasWidget(t *testing.T) {
	cfg := Default()
	cfg.Canvas.Widgets[0].X = 1900
	cfg.Canvas.Widgets[0].Width = 200
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected out-of-bounds canvas widget to fail validation")
	}
}

func TestOlderConfigGetsDefaultCanvas(t *testing.T) {
	cfg := Default()
	cfg.Canvas = Canvas{}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.Canvas.BackgroundType == "" || len(cfg.Canvas.Widgets) == 0 {
		t.Fatal("missing canvas was not populated with defaults")
	}
}
