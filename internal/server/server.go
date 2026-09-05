package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"zima-display/internal/config"
	"zima-display/internal/dsh"
	"zima-display/internal/metrics"
	"zima-display/internal/player"
	"zima-display/internal/presentation"
)

const (
	apiPrefix      = "/zima-display/api"
	maxJSONBody    = 1 << 20
	maxUploadBytes = int64(20 << 30)
)

type Server struct {
	logger        *log.Logger
	config        *config.Store
	player        *player.Manager
	version       string
	started       time.Time
	mu            sync.RWMutex
	metrics       metrics.Snapshot
	presentations *presentation.Store
	dshInstaller  *dsh.Installer
}

type statusResponse struct {
	Version       string           `json:"version"`
	ServiceUptime float64          `json:"service_uptime_seconds"`
	Metrics       metrics.Snapshot `json:"metrics"`
	Player        player.Status    `json:"player"`
	Config        config.Config    `json:"config"`
}

type actionRequest struct {
	Action string   `json:"action"`
	Mode   string   `json:"mode,omitempty"`
	Path   string   `json:"path,omitempty"`
	Paths  []string `json:"paths,omitempty"`
	Value  float64  `json:"value,omitempty"`
	Paused bool     `json:"paused,omitempty"`
}

type mediaEntry struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Kind     string    `json:"kind"`
	Label    string    `json:"label,omitempty"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

type mediaResponse struct {
	Path    string       `json:"path"`
	Parent  string       `json:"parent,omitempty"`
	Entries []mediaEntry `json:"entries"`
}

type textPresentationRequest struct {
	Title  string   `json:"title"`
	Source string   `json:"source,omitempty"`
	Pages  []string `json:"pages"`
}

type dshInstallRequest struct {
	BaseURL string `json:"base_url"`
}

func New(logger *log.Logger, store *config.Store, controller *player.Manager, version string) *Server {
	return &Server{
		logger:        logger,
		config:        store,
		player:        controller,
		version:       version,
		started:       time.Now(),
		presentations: presentation.New(filepath.Join(store.Get().DataDir, "presentations")),
		dshInstaller:  dsh.New("/usr/bin/zima-displayctl"),
	}
}

func (s *Server) SetMetrics(snapshot metrics.Snapshot) {
	s.mu.Lock()
	s.metrics = snapshot
	s.mu.Unlock()
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(apiPrefix+"/health", s.health)
	mux.HandleFunc(apiPrefix+"/status", s.status)
	mux.HandleFunc(apiPrefix+"/config", s.configuration)
	mux.HandleFunc(apiPrefix+"/action", s.action)
	mux.HandleFunc(apiPrefix+"/media", s.media)
	mux.HandleFunc(apiPrefix+"/upload", s.upload)
	mux.HandleFunc(apiPrefix+"/integration/dsh", s.dshIntegration)
	mux.Handle(apiPrefix+"/v1/", s.automationAuth(http.HandlerFunc(s.automation)))
	return s.middleware(mux)
}

func (s *Server) automation(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, apiPrefix+"/v1")
	switch {
	case path == "/presentations" && r.Method == http.MethodPost:
		s.createImagePresentation(w, r)
	case path == "/presentations/text" && r.Method == http.MethodPost:
		s.createTextPresentation(w, r)
	case strings.HasPrefix(path, "/presentations/"):
		s.presentationResource(w, r, strings.TrimPrefix(path, "/presentations/"))
	case path == "/presentation/status" && r.Method == http.MethodGet:
		s.status(w, r)
	case path == "/presentation/next" && r.Method == http.MethodPost:
		s.presentationAction(w, r, "next")
	case path == "/presentation/previous" && r.Method == http.MethodPost:
		s.presentationAction(w, r, "previous")
	case path == "/presentation/stop" && r.Method == http.MethodPost:
		s.presentationAction(w, r, "stop")
	case path == "/action" && r.Method == http.MethodPost:
		s.action(w, r)
	case path == "/upload" && r.Method == http.MethodPost:
		s.upload(w, r)
	default:
		writeError(w, http.StatusNotFound, errors.New("automation endpoint not found"))
	}
}

func (s *Server) automationAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := s.config.Get()
		if !cfg.Automation.Enabled {
			writeError(w, http.StatusForbidden, errors.New("automation API is disabled"))
			return
		}
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !dsh.TokenMatches(cfg.Automation.Token, token) {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, errors.New("invalid automation token"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) createTextPresentation(w http.ResponseWriter, r *http.Request) {
	var request textPresentationRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	manifest, err := s.presentations.CreateText(request.Title, request.Source, request.Pages)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, manifest)
}

func (s *Server) createImagePresentation(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 2<<30)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("parse presentation upload: %w", err))
		return
	}
	file, header, err := r.FormFile("archive")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer file.Close()
	tmp, err := os.CreateTemp(s.config.Get().DataDir, ".presentation-*.zip")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := tmp.Close(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	title := r.FormValue("title")
	if title == "" {
		title = strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	}
	manifest, err := s.presentations.ImportZip(tmpPath, title, r.FormValue("source"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, manifest)
}

func (s *Server) presentationResource(w http.ResponseWriter, r *http.Request, resource string) {
	parts := strings.Split(strings.Trim(resource, "/"), "/")
	if len(parts) == 2 && parts[1] == "activate" && r.Method == http.MethodPost {
		manifest, err := s.presentations.Get(parts[0])
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		var pages []string
		if manifest.Kind == "text" {
			pages, err = s.presentations.TextPages(manifest)
		} else {
			pages, err = s.presentations.PagePaths(manifest)
		}
		if err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
			defer cancel()
			err = s.player.Present(ctx, manifest.ID, manifest.Title, manifest.Kind, pages)
		}
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "presentation": manifest})
		return
	}
	if len(parts) == 1 && r.Method == http.MethodDelete {
		if err := s.presentations.Delete(parts[0]); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	writeError(w, http.StatusNotFound, errors.New("presentation endpoint not found"))
}

func (s *Server) presentationAction(w http.ResponseWriter, r *http.Request, name string) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	var err error
	switch name {
	case "next":
		err = s.player.Next(ctx)
	case "previous":
		err = s.player.Previous(ctx)
	case "stop":
		err = s.player.Stop(ctx)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) dshIntegration(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.dshInstaller.Status())
	case http.MethodPost:
		var request dshInstallRequest
		if err := decodeJSON(w, r, &request); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		parsed, err := url.Parse(strings.TrimRight(request.BaseURL, "/"))
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			writeError(w, http.StatusBadRequest, errors.New("base_url must be an HTTP or HTTPS origin with the /zima-display path"))
			return
		}
		status, err := s.dshInstaller.Install(dsh.Config{BaseURL: parsed.String(), Token: s.config.Get().Automation.Token, Version: s.version})
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, status)
	case http.MethodDelete:
		if err := s.dshInstaller.Uninstall(); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, s.dshInstaller.Status())
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPost, http.MethodDelete)
	}
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "version": s.version})
}

func (s *Server) status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	s.mu.RLock()
	snapshot := s.metrics
	s.mu.RUnlock()
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	writeJSON(w, http.StatusOK, statusResponse{
		Version:       s.version,
		ServiceUptime: time.Since(s.started).Seconds(),
		Metrics:       snapshot,
		Player:        s.player.Status(ctx),
		Config:        publicConfig(s.config.Get()),
	})
}

func (s *Server) configuration(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, publicConfig(s.config.Get()))
	case http.MethodPut:
		var cfg config.Config
		if err := decodeJSON(w, r, &cfg); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		current := s.config.Get()
		if cfg.Automation.Token != "" && cfg.Automation.Token != current.Automation.Token {
			writeError(w, http.StatusBadRequest, errors.New("automation token cannot be changed through the configuration API"))
			return
		}
		cfg.Automation.Token = current.Automation.Token
		if err := s.config.Replace(cfg); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, publicConfig(s.config.Get()))
	default:
		methodNotAllowed(w, http.MethodGet, http.MethodPut)
	}
}

func publicConfig(cfg config.Config) config.Config {
	cfg.Automation.Token = ""
	return cfg
}

func (s *Server) action(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var request actionRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	var err error
	switch request.Action {
	case "mode":
		err = s.player.SetMode(ctx, request.Mode)
	case "play":
		targets := request.Paths
		if request.Path != "" {
			targets = append([]string{request.Path}, targets...)
		}
		resolved := make([]string, 0, len(targets))
		for _, target := range targets {
			value, resolveErr := resolveTarget(s.config.Get(), target)
			if resolveErr != nil {
				err = resolveErr
				break
			}
			resolved = append(resolved, value)
		}
		if err == nil {
			err = s.player.Play(ctx, resolved)
		}
	case "pause":
		err = s.player.Pause(ctx, true)
	case "resume":
		err = s.player.Pause(ctx, false)
	case "stop":
		err = s.player.Stop(ctx)
	case "seek":
		err = s.player.Seek(ctx, request.Value)
	case "volume":
		err = s.player.SetVolume(ctx, request.Value)
		if err == nil {
			cfg := s.config.Get()
			cfg.Volume = int(request.Value)
			err = s.config.Replace(cfg)
		}
	case "next":
		err = s.player.Next(ctx)
	case "previous":
		err = s.player.Previous(ctx)
	default:
		err = fmt.Errorf("unsupported action %q", request.Action)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) media(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	cfg := s.config.Get()
	requested := strings.TrimSpace(r.URL.Query().Get("path"))
	if requested == "" {
		entries := make([]mediaEntry, 0, len(cfg.MediaRoots)+1)
		uploadRoot := filepath.Join(cfg.DataDir, "media")
		entries = append(entries, mediaEntry{Name: "Uploads", Path: uploadRoot, Kind: "root", Label: "uploads"})
		for _, root := range cfg.MediaRoots {
			if filepath.Clean(root) == filepath.Clean(uploadRoot) {
				continue
			}
			entries = append(entries, mediaEntry{Name: filepath.Base(root), Path: root, Kind: "root"})
		}
		writeJSON(w, http.StatusOK, mediaResponse{Entries: entries})
		return
	}
	path, err := resolveLocalPath(cfg, requested, true)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result := make([]mediaEntry, 0, len(entries))
	for _, entry := range entries {
		info, infoErr := entry.Info()
		if infoErr != nil || entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		kind := mediaKind(entry.Name(), entry.IsDir())
		if kind == "" {
			continue
		}
		result = append(result, mediaEntry{
			Name:     entry.Name(),
			Path:     filepath.Join(path, entry.Name()),
			Kind:     kind,
			Size:     info.Size(),
			Modified: info.ModTime(),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		leftDir := result[i].Kind == "directory"
		rightDir := result[j].Kind == "directory"
		if leftDir != rightDir {
			return leftDir
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})
	parent := filepath.Dir(path)
	if _, err := resolveLocalPath(cfg, parent, true); err != nil || parent == path {
		parent = ""
	}
	writeJSON(w, http.StatusOK, mediaResponse{Path: path, Parent: parent, Entries: result})
}

func (s *Server) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("parse upload: %w", err))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer file.Close()
	if mediaKind(header.Filename, false) == "" {
		writeError(w, http.StatusBadRequest, errors.New("unsupported media file type"))
		return
	}
	cfg := s.config.Get()
	directory := filepath.Join(cfg.DataDir, "media")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	name := safeFilename(header)
	destination := uniqueDestination(directory, name)
	tmp, err := os.CreateTemp(directory, ".upload-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := tmp.Close(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := os.Rename(tmpPath, destination); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, mediaEntry{Name: filepath.Base(destination), Path: destination, Kind: mediaKind(destination, false)})
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		defer func() {
			if recovered := recover(); recovered != nil {
				s.logger.Printf("panic serving %s: %v", r.URL.Path, recovered)
				writeError(w, http.StatusInternalServerError, errors.New("internal server error"))
			}
			s.logger.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(started).Round(time.Millisecond))
		}()
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			if !strings.HasSuffix(r.URL.Path, "/upload") && !strings.HasSuffix(r.URL.Path, "/presentations") {
				r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func resolveTarget(cfg config.Config, target string) (string, error) {
	target = strings.TrimSpace(target)
	parsed, err := url.Parse(target)
	if err == nil && parsed.Scheme != "" {
		if parsed.Scheme != "http" && parsed.Scheme != "https" && parsed.Scheme != "rtsp" {
			return "", fmt.Errorf("unsupported URL scheme %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("media URL has no host")
		}
		return target, nil
	}
	path, err := resolveLocalPath(cfg, target, false)
	if err != nil {
		return "", err
	}
	if mediaKind(path, false) == "" {
		return "", errors.New("unsupported media file type")
	}
	return path, nil
}

func resolveLocalPath(cfg config.Config, requested string, requireDirectory bool) (string, error) {
	clean := filepath.Clean(requested)
	if !filepath.IsAbs(clean) {
		return "", errors.New("media path must be absolute")
	}
	real, err := filepath.EvalSymlinks(clean)
	if err != nil {
		return "", err
	}
	allowed := false
	for _, root := range cfg.MediaRoots {
		realRoot, rootErr := filepath.EvalSymlinks(root)
		if rootErr != nil {
			continue
		}
		if within(realRoot, real) {
			allowed = true
			break
		}
	}
	if !allowed && within(cfg.DataDir, real) {
		allowed = true
	}
	if !allowed {
		return "", errors.New("media path is outside configured roots")
	}
	info, err := os.Stat(real)
	if err != nil {
		return "", err
	}
	if requireDirectory && !info.IsDir() {
		return "", errors.New("media path is not a directory")
	}
	if !requireDirectory && !info.Mode().IsRegular() {
		return "", errors.New("media path is not a regular file")
	}
	return real, nil
}

func within(root, path string) bool {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))
}

func mediaKind(name string, directory bool) string {
	if directory {
		return "directory"
	}
	extension := strings.ToLower(filepath.Ext(name))
	for _, value := range []string{".mp4", ".mkv", ".mov", ".avi", ".webm", ".m4v", ".ts", ".m2ts", ".flv", ".mpg", ".mpeg"} {
		if extension == value {
			return "video"
		}
	}
	for _, value := range []string{".jpg", ".jpeg", ".png", ".webp", ".bmp", ".gif"} {
		if extension == value {
			return "image"
		}
	}
	return ""
}

func safeFilename(header *multipart.FileHeader) string {
	name := strings.TrimSpace(filepath.Base(header.Filename))
	name = strings.ReplaceAll(name, "\x00", "")
	if name == "" || name == "." {
		return "upload.bin"
	}
	return name
}

func uniqueDestination(directory, name string) string {
	destination := filepath.Join(directory, name)
	if _, err := os.Stat(destination); errors.Is(err, os.ErrNotExist) {
		return destination
	}
	extension := filepath.Ext(name)
	base := strings.TrimSuffix(name, extension)
	for index := 1; ; index++ {
		candidate := filepath.Join(directory, fmt.Sprintf("%s-%d%s", base, index, extension))
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate
		}
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func methodNotAllowed(w http.ResponseWriter, methods ...string) {
	w.Header().Set("Allow", strings.Join(methods, ", "))
	writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
}
