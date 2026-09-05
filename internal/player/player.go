package player

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"zima-display/internal/config"
	"zima-display/internal/metrics"
)

const (
	rendererService      = "zima-display-renderer.service"
	dashboardVideoSource = "av://lavfi:color=c=black:s=1920x1080:r=1"
)

type Status struct {
	Mode          string              `json:"mode"`
	RendererReady bool                `json:"renderer_ready"`
	Paused        bool                `json:"paused"`
	Position      float64             `json:"position_seconds"`
	Duration      float64             `json:"duration_seconds"`
	Volume        float64             `json:"volume"`
	MediaTitle    string              `json:"media_title,omitempty"`
	Path          string              `json:"path,omitempty"`
	PlaylistIndex int                 `json:"playlist_index"`
	PlaylistCount int                 `json:"playlist_count"`
	LastError     string              `json:"last_error,omitempty"`
	Presentation  *PresentationStatus `json:"presentation,omitempty"`
}

type PresentationStatus struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Kind      string `json:"kind"`
	Page      int    `json:"page"`
	PageCount int    `json:"page_count"`
}

type commandRunner interface {
	Run(ctx context.Context, name string, args ...string) error
	Output(ctx context.Context, name string, args ...string) ([]byte, error)
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args ...string) error {
	return exec.CommandContext(ctx, name, args...).Run()
}

func (execRunner) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}

type Manager struct {
	mu                sync.Mutex
	socketPath        string
	runtimePath       string
	config            func() config.Config
	runner            commandRunner
	mode              string
	lastError         string
	snapshot          metrics.Snapshot
	overlayConn       net.Conn
	overlayRead       *bufio.Reader
	requestID         atomic.Int64
	presentationID    string
	presentationTitle string
	presentationKind  string
	presentationPages []string
	presentationIndex int
}

func New(socketPath, runtimePath string, provider func() config.Config) *Manager {
	return &Manager{
		socketPath:  socketPath,
		runtimePath: runtimePath,
		config:      provider,
		runner:      execRunner{},
		mode:        "terminal",
	}
}

func (m *Manager) SetMode(ctx context.Context, mode string) error {
	if mode != "dashboard" && mode != "clock" && mode != "black" && mode != "terminal" {
		return fmt.Errorf("unsupported display mode %q", mode)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if mode == "terminal" {
		stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		m.closeOverlayLocked()
		_ = m.runner.Run(stopCtx, "systemctl", "stop", rendererService)
		if err := m.runner.Run(stopCtx, "systemctl", "start", "getty@tty1.service"); err != nil {
			return m.fail(err)
		}
		m.clearPresentationLocked()
		m.mode = mode
		m.lastError = ""
		return nil
	}

	if err := m.ensureRendererLocked(ctx); err != nil {
		return m.fail(err)
	}
	if mode == "dashboard" || mode == "clock" || mode == "black" {
		if err := m.showGeneratedModeLocked(ctx, mode); err != nil {
			return m.fail(err)
		}
	}
	m.clearPresentationLocked()
	m.mode = mode
	m.lastError = ""
	return nil
}

func (m *Manager) Play(ctx context.Context, targets []string) error {
	if len(targets) == 0 {
		return errors.New("no media target supplied")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.ensureRendererLocked(ctx); err != nil {
		return m.fail(err)
	}
	m.closeOverlayLocked()
	for index, target := range targets {
		mode := "append-play"
		if index == 0 {
			mode = "replace"
		}
		if err := m.send(ctx, []any{"loadfile", target, mode}, nil); err != nil {
			return m.fail(err)
		}
	}
	m.clearPresentationLocked()
	m.mode = "video"
	m.lastError = ""
	return nil
}

func (m *Manager) Present(ctx context.Context, id, title, kind string, pages []string) error {
	if len(pages) == 0 {
		return errors.New("presentation contains no pages")
	}
	if kind != "images" && kind != "text" {
		return fmt.Errorf("unsupported presentation kind %q", kind)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.ensureRendererLocked(ctx); err != nil {
		return m.fail(err)
	}
	if kind == "text" {
		if err := m.send(ctx, []any{"loadfile", dashboardVideoSource, "replace"}, nil); err != nil {
			return m.fail(err)
		}
		if err := m.send(ctx, []any{"set_property", "pause", false}, nil); err != nil {
			return m.fail(err)
		}
		if err := m.setOverlayLocked(ctx, renderDocument(title, pages[0], 0, len(pages))); err != nil {
			return m.fail(err)
		}
	} else {
		m.closeOverlayLocked()
		for index, page := range pages {
			mode := "append-play"
			if index == 0 {
				mode = "replace"
			}
			if err := m.send(ctx, []any{"loadfile", page, mode}, nil); err != nil {
				return m.fail(err)
			}
		}
	}
	m.presentationID = id
	m.presentationTitle = title
	m.presentationKind = kind
	m.presentationPages = append([]string(nil), pages...)
	m.presentationIndex = 0
	m.mode = "presentation"
	m.lastError = ""
	return nil
}

func (m *Manager) Pause(ctx context.Context, paused bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.send(ctx, []any{"set_property", "pause", paused}, nil); err != nil {
		return m.fail(err)
	}
	return nil
}

func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.showGeneratedModeLocked(ctx, "dashboard"); err != nil {
		return m.fail(err)
	}
	m.clearPresentationLocked()
	m.mode = "dashboard"
	m.lastError = ""
	return nil
}

func (m *Manager) showGeneratedModeLocked(ctx context.Context, mode string) error {
	if err := m.send(ctx, []any{"loadfile", dashboardVideoSource, "replace"}, nil); err != nil {
		return err
	}
	if err := m.send(ctx, []any{"set_property", "pause", false}, nil); err != nil {
		return err
	}
	if mode == "black" {
		m.closeOverlayLocked()
		return nil
	}
	return m.setOverlayLocked(ctx, renderMode(mode, m.snapshot, m.config(), time.Now()))
}

func (m *Manager) UpdateMetrics(ctx context.Context, snapshot metrics.Snapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshot = snapshot
	if m.mode != "dashboard" && m.mode != "clock" {
		return
	}
	if err := m.setOverlayLocked(ctx, renderMode(m.mode, snapshot, m.config(), time.Now())); err != nil {
		m.lastError = err.Error()
		return
	}
	m.lastError = ""
}

func (m *Manager) setOverlayLocked(ctx context.Context, data string) error {
	if m.overlayConn == nil {
		dialer := net.Dialer{Timeout: 2 * time.Second}
		connection, err := dialer.DialContext(ctx, "unix", m.socketPath)
		if err != nil {
			return fmt.Errorf("connect to mpv overlay: %w", err)
		}
		m.overlayConn = connection
		m.overlayRead = bufio.NewReader(connection)
	}

	requestID := m.requestID.Add(1)
	request := struct {
		Command   []any `json:"command"`
		RequestID int64 `json:"request_id"`
	}{
		Command:   []any{"osd-overlay", 1, "ass-events", data, 1920, 1080, 10},
		RequestID: requestID,
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}
	_ = m.overlayConn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := m.overlayConn.Write(append(payload, '\n')); err != nil {
		m.closeOverlayLocked()
		return err
	}
	for {
		line, err := m.overlayRead.ReadBytes('\n')
		if err != nil {
			m.closeOverlayLocked()
			return err
		}
		var response struct {
			Error     string `json:"error"`
			RequestID int64  `json:"request_id"`
		}
		if json.Unmarshal(line, &response) != nil || response.RequestID != requestID {
			continue
		}
		if response.Error != "success" {
			return fmt.Errorf("mpv overlay failed: %s", response.Error)
		}
		return nil
	}
}

func (m *Manager) closeOverlayLocked() {
	if m.overlayConn != nil {
		_ = m.overlayConn.Close()
	}
	m.overlayConn = nil
	m.overlayRead = nil
}

func (m *Manager) Seek(ctx context.Context, seconds float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.send(ctx, []any{"seek", seconds, "relative+exact"}, nil); err != nil {
		return m.fail(err)
	}
	return nil
}

func (m *Manager) SetVolume(ctx context.Context, volume float64) error {
	if volume < 0 || volume > 100 {
		return errors.New("volume must be between 0 and 100")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.send(ctx, []any{"set_property", "volume", volume}, nil); err != nil {
		return m.fail(err)
	}
	return nil
}

func (m *Manager) Next(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mode == "presentation" && m.presentationKind == "text" {
		if m.presentationIndex < len(m.presentationPages)-1 {
			m.presentationIndex++
		}
		return m.setOverlayLocked(ctx, renderDocument(m.presentationTitle, m.presentationPages[m.presentationIndex], m.presentationIndex, len(m.presentationPages)))
	}
	if err := m.send(ctx, []any{"playlist-next", "force"}, nil); err != nil {
		return m.fail(err)
	}
	return nil
}

func (m *Manager) Previous(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mode == "presentation" && m.presentationKind == "text" {
		if m.presentationIndex > 0 {
			m.presentationIndex--
		}
		return m.setOverlayLocked(ctx, renderDocument(m.presentationTitle, m.presentationPages[m.presentationIndex], m.presentationIndex, len(m.presentationPages)))
	}
	if err := m.send(ctx, []any{"playlist-prev", "force"}, nil); err != nil {
		return m.fail(err)
	}
	return nil
}

func (m *Manager) Status(ctx context.Context) Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	status := Status{Mode: m.mode, LastError: m.lastError, PlaylistIndex: -1}
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := m.runner.Run(checkCtx, "systemctl", "is-active", "--quiet", rendererService); err != nil {
		return status
	}
	status.RendererReady = true
	properties := map[string]any{}
	for _, property := range []string{"pause", "time-pos", "duration", "volume", "media-title", "path", "playlist-pos", "playlist-count"} {
		var value any
		if err := m.send(ctx, []any{"get_property", property}, &value); err == nil {
			properties[property] = value
		}
	}
	status.Paused, _ = properties["pause"].(bool)
	status.Position = asFloat(properties["time-pos"])
	status.Duration = asFloat(properties["duration"])
	status.Volume = asFloat(properties["volume"])
	status.MediaTitle, _ = properties["media-title"].(string)
	status.Path, _ = properties["path"].(string)
	status.PlaylistIndex = int(asFloat(properties["playlist-pos"]))
	status.PlaylistCount = int(asFloat(properties["playlist-count"]))
	if m.mode == "presentation" {
		page := m.presentationIndex
		if m.presentationKind == "images" && status.PlaylistIndex >= 0 {
			page = status.PlaylistIndex
			m.presentationIndex = page
		}
		status.Presentation = &PresentationStatus{
			ID: m.presentationID, Title: m.presentationTitle, Kind: m.presentationKind,
			Page: page + 1, PageCount: len(m.presentationPages),
		}
		status.MediaTitle = m.presentationTitle
		status.Position = 0
		status.Duration = 0
	} else if m.mode != "video" {
		status.Position = 0
		status.Duration = 0
		status.MediaTitle = ""
		status.Path = ""
		status.PlaylistIndex = -1
		status.PlaylistCount = 0
	}
	return status
}

func (m *Manager) clearPresentationLocked() {
	m.presentationID = ""
	m.presentationTitle = ""
	m.presentationKind = ""
	m.presentationPages = nil
	m.presentationIndex = 0
}

func (m *Manager) ensureRendererLocked(ctx context.Context) error {
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	err := m.runner.Run(checkCtx, "systemctl", "is-active", "--quiet", rendererService)
	if err == nil && m.socketReady(checkCtx) {
		cancel()
		return nil
	}
	cancel()
	if err := m.prepareRuntimeLocked(); err != nil {
		return err
	}
	commandCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	m.closeOverlayLocked()
	_ = m.runner.Run(commandCtx, "systemctl", "stop", "getty@tty1.service")
	_ = os.Remove(m.socketPath)
	if err := m.runner.Run(commandCtx, "systemctl", "restart", rendererService); err != nil {
		return fmt.Errorf("start renderer: %w", err)
	}
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if m.socketReady(ctx) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
	return errors.New("renderer started but mpv control socket did not appear")
}

func (m *Manager) socketReady(ctx context.Context) bool {
	dialer := net.Dialer{Timeout: 250 * time.Millisecond}
	connection, err := dialer.DialContext(ctx, "unix", m.socketPath)
	if err != nil {
		return false
	}
	_ = connection.Close()
	return true
}

func (m *Manager) prepareRuntimeLocked() error {
	if err := os.MkdirAll(m.runtimePath, 0o755); err != nil {
		return err
	}
	cfg := m.config()
	values := map[string]string{
		"ZIMA_DISPLAY_BACKEND":      cfg.Renderer.Backend,
		"ZIMA_DISPLAY_AUDIO_DEVICE": cfg.Renderer.AudioDevice,
		"ZIMA_DISPLAY_RESOLUTION":   cfg.Renderer.Resolution,
		"ZIMA_DISPLAY_VOLUME":       strconv.Itoa(cfg.Volume),
	}
	var lines []string
	for _, key := range []string{"ZIMA_DISPLAY_BACKEND", "ZIMA_DISPLAY_AUDIO_DEVICE", "ZIMA_DISPLAY_RESOLUTION", "ZIMA_DISPLAY_VOLUME"} {
		lines = append(lines, key+"="+shellQuote(values[key]))
	}
	return os.WriteFile(filepath.Join(m.runtimePath, "renderer.env"), []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

func (m *Manager) send(ctx context.Context, command []any, destination *any) error {
	requestID := m.requestID.Add(1)
	request := struct {
		Command   []any `json:"command"`
		RequestID int64 `json:"request_id"`
	}{Command: command, RequestID: requestID}
	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}
	dialer := net.Dialer{Timeout: 2 * time.Second}
	connection, err := dialer.DialContext(ctx, "unix", m.socketPath)
	if err != nil {
		return fmt.Errorf("connect to mpv: %w", err)
	}
	defer connection.Close()
	_ = connection.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := connection.Write(append(payload, '\n')); err != nil {
		return err
	}
	scanner := bufio.NewScanner(connection)
	for scanner.Scan() {
		var response struct {
			Error     string          `json:"error"`
			Data      json.RawMessage `json:"data"`
			RequestID int64           `json:"request_id"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &response); err != nil || response.RequestID != requestID {
			continue
		}
		if response.Error != "success" {
			return fmt.Errorf("mpv command failed: %s", response.Error)
		}
		if destination != nil && len(response.Data) > 0 {
			if err := json.Unmarshal(response.Data, destination); err != nil {
				return err
			}
		}
		return nil
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return errors.New("mpv closed the control connection without a response")
}

func (m *Manager) fail(err error) error {
	m.lastError = err.Error()
	return err
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func asFloat(value any) float64 {
	switch number := value.(type) {
	case float64:
		return number
	case int:
		return float64(number)
	case json.Number:
		parsed, _ := number.Float64()
		return parsed
	default:
		return 0
	}
}
