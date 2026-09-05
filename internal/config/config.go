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
	Title             string  `json:"title"`
	Language          string  `json:"language"`
	Layout            string  `json:"layout"`
	FontScale         float64 `json:"font_scale"`
	RefreshIntervalMS int     `json:"refresh_interval_ms"`
}

type Automation struct {
	Enabled bool   `json:"enabled"`
	Token   string `json:"token"`
}

type CanvasWidget struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FontSize int    `json:"font_size"`
	Color    string `json:"color"`
	Bold     bool   `json:"bold"`
}

type Canvas struct {
	BackgroundType  string         `json:"background_type"`
	BackgroundColor string         `json:"background_color"`
	SecondaryColor  string         `json:"secondary_color"`
	BackgroundImage string         `json:"background_image,omitempty"`
	Widgets         []CanvasWidget `json:"widgets"`
}

type Config struct {
	DataDir     string     `json:"data_dir"`
	MediaRoots  []string   `json:"media_roots"`
	DefaultMode string     `json:"default_mode"`
	AutoStart   bool       `json:"auto_start_display"`
	Volume      int        `json:"volume"`
	Renderer    Renderer   `json:"renderer"`
	Dashboard   Dashboard  `json:"dashboard"`
	Canvas      Canvas     `json:"canvas"`
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
			Layout:            "auto",
			FontScale:         1,
			RefreshIntervalMS: 1000,
		},
		Canvas:     defaultCanvas(),
		Automation: Automation{Enabled: true},
	}
}

func defaultCanvas() Canvas {
	return Canvas{
		BackgroundType: "gradient", BackgroundColor: "#151411", SecondaryColor: "#30251D",
		Widgets: []CanvasWidget{
			{ID: "clock", Type: "clock", X: 90, Y: 70, Width: 760, Height: 210, FontSize: 180, Color: "#F7F3EA", Bold: true},
			{ID: "date", Type: "date", X: 100, Y: 290, Width: 720, Height: 70, FontSize: 38, Color: "#B9B5AC"},
			{ID: "cpu", Type: "cpu", X: 1040, Y: 100, Width: 360, Height: 100, FontSize: 64, Color: "#F26618", Bold: true},
			{ID: "memory", Type: "memory", X: 1040, Y: 230, Width: 420, Height: 100, FontSize: 58, Color: "#F7F3EA", Bold: true},
			{ID: "temperature", Type: "temperature", X: 1040, Y: 360, Width: 420, Height: 100, FontSize: 58, Color: "#F7F3EA", Bold: true},
			{ID: "ip", Type: "ip", X: 100, Y: 900, Width: 900, Height: 70, FontSize: 36, Color: "#F7F3EA", Bold: true},
			{ID: "hostname", Type: "hostname", X: 1040, Y: 900, Width: 700, Height: 70, FontSize: 36, Color: "#B9B5AC"},
		},
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
	if !oneOf(c.DefaultMode, "dashboard", "canvas", "clock", "black", "terminal") {
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
	if c.Dashboard.Layout == "" {
		c.Dashboard.Layout = "auto"
	}
	if !oneOf(c.Dashboard.Layout, "auto", "standard", "large") {
		return fmt.Errorf("unsupported dashboard layout %q", c.Dashboard.Layout)
	}
	if c.Dashboard.FontScale == 0 {
		c.Dashboard.FontScale = 1
	}
	if c.Dashboard.FontScale < 0.8 || c.Dashboard.FontScale > 2 {
		return errors.New("dashboard font scale must be between 0.8 and 2.0")
	}
	if c.Canvas.BackgroundType == "" && len(c.Canvas.Widgets) == 0 {
		c.Canvas = defaultCanvas()
	}
	if !oneOf(c.Canvas.BackgroundType, "solid", "gradient", "image") {
		return fmt.Errorf("unsupported canvas background type %q", c.Canvas.BackgroundType)
	}
	if !validColor(c.Canvas.BackgroundColor) || !validColor(c.Canvas.SecondaryColor) {
		return errors.New("canvas colors must use #RRGGBB format")
	}
	if c.Canvas.BackgroundImage != "" {
		clean := filepath.Clean(c.Canvas.BackgroundImage)
		if !filepath.IsAbs(clean) || !allowedMediaRoot(clean) {
			return errors.New("canvas background image must be below /DATA, /media, or /mnt")
		}
		c.Canvas.BackgroundImage = clean
	}
	seenWidgets := make(map[string]struct{}, len(c.Canvas.Widgets))
	for i := range c.Canvas.Widgets {
		widget := &c.Canvas.Widgets[i]
		widget.ID = strings.TrimSpace(widget.ID)
		if widget.ID == "" {
			return errors.New("canvas widget id cannot be empty")
		}
		if _, exists := seenWidgets[widget.ID]; exists {
			return fmt.Errorf("duplicate canvas widget id %q", widget.ID)
		}
		seenWidgets[widget.ID] = struct{}{}
		if !oneOf(widget.Type, "clock", "date", "cpu", "memory", "storage", "temperature", "ip", "hostname", "network", "text") {
			return fmt.Errorf("unsupported canvas widget type %q", widget.Type)
		}
		if widget.X < 0 || widget.Y < 0 || widget.Width < 80 || widget.Height < 30 || widget.X+widget.Width > 1920 || widget.Y+widget.Height > 1080 {
			return fmt.Errorf("canvas widget %q is outside the 1920x1080 canvas", widget.ID)
		}
		if widget.FontSize < 12 || widget.FontSize > 300 {
			return fmt.Errorf("canvas widget %q font size must be between 12 and 300", widget.ID)
		}
		if !validColor(widget.Color) {
			return fmt.Errorf("canvas widget %q color must use #RRGGBB format", widget.ID)
		}
		if widget.Type == "text" {
			widget.Text = strings.TrimSpace(widget.Text)
			if widget.Text == "" {
				widget.Text = "Text"
			}
		}
	}
	return nil
}

func validColor(value string) bool {
	if len(value) != 7 || value[0] != '#' {
		return false
	}
	for _, char := range value[1:] {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
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
	cfg.Canvas.Widgets = append([]CanvasWidget(nil), cfg.Canvas.Widgets...)
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
