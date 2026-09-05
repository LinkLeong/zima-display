package config

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	DefaultDataDir = "/DATA/AppData/zima-display"
	ConfigName     = "config.json"
)

type Renderer struct {
	Backend     string `json:"backend"`
	Resolution  string `json:"resolution"`
	AudioDevice string `json:"audio_device"`
}

type Dashboard struct {
	Title             string `json:"title"`
	Language          string `json:"language"`
	RefreshIntervalMS int    `json:"refresh_interval_ms"`
}

type Automation struct {
	Enabled bool   `json:"enabled"`
	Token   string `json:"token"`
}

type Config struct {
	DataDir     string     `json:"data_dir"`
	MediaRoots  []string   `json:"media_roots"`
	DefaultMode string     `json:"default_mode"`
	AutoStart   bool       `json:"auto_start_display"`
	Volume      int        `json:"volume"`
	Renderer    Renderer   `json:"renderer"`
	Dashboard   Dashboard  `json:"dashboard"`
	Automation  Automation `json:"automation"`
}

type Store struct {
	mu   sync.RWMutex
	path string
	cfg  Config
}

func NewStore(path string) (*Store, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return nil, fmt.Errorf("secure configuration file: %w", err)
	}
	if cfg.Automation.Token == "" {
		cfg.Automation.Token, err = newToken()
		if err != nil {
			return nil, err
		}
		cfg.Automation.Enabled = true
		if err := Save(path, cfg); err != nil {
			return nil, err
		}
	}
	return &Store{path: path, cfg: cfg}, nil
}

func (s *Store) Get() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return clone(s.cfg)
}

func (s *Store) Replace(cfg Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cfg.DataDir != s.cfg.DataDir {
		return errors.New("data_dir cannot be changed at runtime")
	}
	if err := Save(s.path, cfg); err != nil {
		return err
	}
	s.cfg = clone(cfg)
	return nil
}

func Default() Config {
	return Config{
		DataDir:     DefaultDataDir,
		MediaRoots:  []string{"/DATA"},
		DefaultMode: "dashboard",
		AutoStart:   true,
		Volume:      70,
		Renderer: Renderer{
			Backend:     "auto",
			Resolution:  "auto",
			AudioDevice: "auto",
		},
		Dashboard: Dashboard{
			Title:             "Zima Display",
			Language:          "zh-CN",
			RefreshIntervalMS: 1000,
		},
		Automation: Automation{Enabled: true},
	}
}

func Path(dataDir string) string {
	return filepath.Join(dataDir, ConfigName)
}

func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		cfg.DataDir = filepath.Dir(path)
		if err := Save(path, cfg); err != nil {
			return Config{}, err
		}
		return cfg, nil
	}
	if err != nil {
		return Config{}, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return err
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (c *Config) Validate() error {
	if !filepath.IsAbs(c.DataDir) {
		return errors.New("data_dir must be absolute")
	}
	if len(c.MediaRoots) == 0 {
		return errors.New("at least one media root is required")
	}
	for i, root := range c.MediaRoots {
		clean := filepath.Clean(root)
		if !filepath.IsAbs(clean) || !allowedMediaRoot(clean) {
			return fmt.Errorf("media root %q must be below /DATA, /media, or /mnt", root)
		}
		c.MediaRoots[i] = clean
	}
	if !oneOf(c.DefaultMode, "dashboard", "clock", "black", "terminal") {
		return fmt.Errorf("unsupported default mode %q", c.DefaultMode)
	}
	if !oneOf(c.Renderer.Backend, "auto", "wayland", "drm") {
		return fmt.Errorf("unsupported renderer backend %q", c.Renderer.Backend)
	}
	if c.Renderer.Resolution == "" {
		c.Renderer.Resolution = "auto"
	}
	if c.Renderer.AudioDevice == "" {
		c.Renderer.AudioDevice = "auto"
	}
	if c.Volume < 0 || c.Volume > 100 {
		return errors.New("volume must be between 0 and 100")
	}
	if c.Dashboard.RefreshIntervalMS < 500 || c.Dashboard.RefreshIntervalMS > 10000 {
		return errors.New("dashboard refresh interval must be between 500 and 10000 ms")
	}
	c.Dashboard.Title = strings.TrimSpace(c.Dashboard.Title)
	if c.Dashboard.Title == "" {
		c.Dashboard.Title = "Zima Display"
	}
	if c.Dashboard.Language == "" {
		c.Dashboard.Language = "zh-CN"
	}
	if !oneOf(c.Dashboard.Language, "zh-CN", "en-US") {
		return fmt.Errorf("unsupported dashboard language %q", c.Dashboard.Language)
	}
	return nil
}

func allowedMediaRoot(path string) bool {
	for _, base := range []string{"/DATA", "/media", "/mnt"} {
		if path == base || strings.HasPrefix(path, base+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func clone(cfg Config) Config {
	cfg.MediaRoots = append([]string(nil), cfg.MediaRoots...)
	return cfg
}

func newToken() (string, error) {
	data := make([]byte, 24)
	if _, err := rand.Read(data); err != nil {
		return "", fmt.Errorf("generate automation token: %w", err)
	}
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for index := range data {
		data[index] = alphabet[int(data[index])%len(alphabet)]
	}
	return string(data), nil
}
