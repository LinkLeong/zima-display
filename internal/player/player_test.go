package player

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"zima-display/internal/config"
	"zima-display/internal/metrics"
)

type readyRunner struct{}

func (readyRunner) Run(context.Context, string, ...string) error { return nil }
func (readyRunner) Output(context.Context, string, ...string) ([]byte, error) {
	return nil, nil
}

type restartRunner struct {
	mu        sync.Mutex
	restarts  int
	onRestart func() error
}

func (r *restartRunner) Run(_ context.Context, name string, args ...string) error {
	if name != "systemctl" || len(args) == 0 || args[0] != "restart" {
		return nil
	}
	r.mu.Lock()
	r.restarts++
	r.mu.Unlock()
	return r.onRestart()
}

func (*restartRunner) Output(context.Context, string, ...string) ([]byte, error) {
	return nil, nil
}

func TestShellQuote(t *testing.T) {
	if value := shellQuote("alsa/o'hdmi"); value != "'alsa/o'\\''hdmi'" {
		t.Fatalf("unexpected shell quote %q", value)
	}
}

func TestAsFloat(t *testing.T) {
	if value := asFloat(json.Number("42.5")); value != 42.5 {
		t.Fatalf("asFloat returned %v", value)
	}
}

func TestMPVReadyRequiresIPCResponse(t *testing.T) {
	directory := t.TempDir()
	socketPath := filepath.Join(directory, "mpv.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr == nil {
			_ = connection.Close()
		}
	}()

	manager := New(socketPath, directory, config.Default)
	if manager.mpvReady(context.Background()) {
		t.Fatal("socket without an IPC response must not be considered ready")
	}
}

func TestStatusRequiresReadyMPVIPC(t *testing.T) {
	manager := New(filepath.Join(t.TempDir(), "missing.sock"), t.TempDir(), config.Default)
	manager.runner = readyRunner{}
	status := manager.Status(context.Background())
	if status.RendererReady {
		t.Fatal("active renderer service without mpv IPC must not be reported as ready")
	}
}

func TestDashboardModeRecoversLostMPVConnection(t *testing.T) {
	directory, err := os.MkdirTemp("/tmp", "zd-recover-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(directory) })
	socketPath := filepath.Join(directory, "mpv.sock")
	commands := make(chan []any, 16)

	var listenerMu sync.Mutex
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	serve := func(current net.Listener, dropCommand int) {
		go func() {
			count := 0
			for {
				connection, acceptErr := current.Accept()
				if acceptErr != nil {
					return
				}
				count++
				var request struct {
					Command   []any `json:"command"`
					RequestID int64 `json:"request_id"`
				}
				if json.NewDecoder(connection).Decode(&request) == nil {
					commands <- request.Command
					if count != dropCommand {
						_ = json.NewEncoder(connection).Encode(map[string]any{"error": "success", "request_id": request.RequestID})
					}
				}
				_ = connection.Close()
			}
		}()
	}
	serve(listener, 2) // The first real command loses its connection after the readiness probe.

	runner := &restartRunner{}
	runner.onRestart = func() error {
		listenerMu.Lock()
		defer listenerMu.Unlock()
		_ = listener.Close()
		newListener, listenErr := net.Listen("unix", socketPath)
		if listenErr != nil {
			return listenErr
		}
		listener = newListener
		serve(listener, 0)
		return nil
	}
	t.Cleanup(func() {
		listenerMu.Lock()
		_ = listener.Close()
		listenerMu.Unlock()
	})

	manager := New(socketPath, directory, config.Default)
	manager.runner = runner
	if err := manager.SetMode(context.Background(), "dashboard"); err != nil {
		t.Fatal(err)
	}
	runner.mu.Lock()
	restarts := runner.restarts
	runner.mu.Unlock()
	if restarts != 1 {
		t.Fatalf("renderer restarts = %d, want 1", restarts)
	}
	if manager.mode != "dashboard" || manager.lastError != "" {
		t.Fatalf("manager did not recover cleanly: mode=%q error=%q", manager.mode, manager.lastError)
	}

	loadCommands := 0
	overlayCommands := 0
	for len(commands) > 0 {
		command := <-commands
		if len(command) > 0 && command[0] == "loadfile" {
			loadCommands++
		}
		if len(command) > 0 && command[0] == "osd-overlay" {
			overlayCommands++
		}
	}
	if loadCommands != 2 || overlayCommands != 1 {
		t.Fatalf("recovery commands: loadfile=%d overlay=%d", loadCommands, overlayCommands)
	}
}

func TestMPVLogicalErrorsDoNotTriggerConnectionRecovery(t *testing.T) {
	if isMPVConnectionError(errors.New("mpv command failed: property unavailable")) {
		t.Fatal("logical mpv errors must not be treated as connection failures")
	}
	if !isMPVConnectionError(errMPVConnectionClosed) {
		t.Fatal("closed mpv connection must trigger recovery")
	}
}

func TestDashboardModeLoadsGeneratedVideo(t *testing.T) {
	directory, err := os.MkdirTemp("/tmp", "zd-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(directory) })
	socketPath := filepath.Join(directory, "mpv.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	requests := make(chan []any, 4)
	go func() {
		for {
			connection, acceptErr := listener.Accept()
			if acceptErr != nil {
				return
			}
			var request struct {
				Command   []any `json:"command"`
				RequestID int64 `json:"request_id"`
			}
			if json.NewDecoder(connection).Decode(&request) == nil {
				requests <- request.Command
				_ = json.NewEncoder(connection).Encode(map[string]any{"error": "success", "request_id": request.RequestID})
			}
			_ = connection.Close()
		}
	}()

	manager := New(socketPath, directory, config.Default)
	manager.runner = readyRunner{}
	if err := manager.SetMode(context.Background(), "dashboard"); err != nil {
		t.Fatal(err)
	}
	want := [][]any{
		{"get_property", "idle-active"},
		{"loadfile", dashboardVideoSource, "replace"},
		{"set_property", "pause", false},
	}
	for index := range want {
		select {
		case got := <-requests:
			if !reflect.DeepEqual(got, want[index]) {
				t.Fatalf("request %d: got %#v want %#v", index, got, want[index])
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for request %d", index)
		}
	}
	select {
	case got := <-requests:
		if len(got) != 7 || got[0] != "osd-overlay" || got[2] != "ass-events" {
			t.Fatalf("unexpected overlay request %#v", got)
		}
		data, _ := got[3].(string)
		if !strings.Contains(data, "系统负载") || !strings.Contains(data, "ZIMA DISPLAY") {
			t.Fatalf("overlay is missing dashboard content: %q", data)
		}
		if !strings.Contains(data, "\n") {
			t.Fatal("dashboard elements must be separate ASS events")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for overlay request")
	}
}

func TestDashboardSupportsEnglish(t *testing.T) {
	cfg := config.Default()
	cfg.Dashboard.Language = "en-US"
	data := renderDashboard(metrics.Snapshot{}, cfg, time.Date(2026, time.September, 4, 12, 30, 0, 0, time.UTC))
	for _, expected := range []string{"SYSTEM LOAD", "MEMORY", "Friday", "Waiting for network"} {
		if !strings.Contains(data, expected) {
			t.Fatalf("English dashboard is missing %q", expected)
		}
	}
}

func TestDashboardSupportsSimplifiedChinese(t *testing.T) {
	cfg := config.Default()
	data := renderDashboard(metrics.Snapshot{}, cfg, time.Date(2026, time.September, 4, 12, 30, 0, 0, time.UTC))
	for _, expected := range []string{"系统负载", "内存", "星期五", "等待网络连接"} {
		if !strings.Contains(data, expected) {
			t.Fatalf("Chinese dashboard is missing %q", expected)
		}
	}
}

func TestDashboardUsesLargeTypeOnCompactDisplay(t *testing.T) {
	cfg := config.Default()
	snapshot := metrics.Snapshot{
		Display:     metrics.Display{Connected: true, Connector: "card0-HDMI-A-1", Modes: []string{"1024x600"}},
		IPAddresses: []string{"192.168.1.20"},
	}
	data := renderDashboard(snapshot, cfg, time.Date(2026, time.September, 4, 12, 30, 0, 0, time.UTC))
	for _, expected := range []string{"\\fs112", "\\fs30", "\\fs24", "192.168.1.20"} {
		if !strings.Contains(data, expected) {
			t.Fatalf("compact dashboard is missing %q", expected)
		}
	}
	if strings.Contains(data, "打开 ZimaOS") {
		t.Fatal("compact dashboard should reserve space for core metrics instead of the control hint")
	}
}

func TestLargeDashboardUsesAlignedCardGrid(t *testing.T) {
	cfg := config.Default()
	cfg.Dashboard.Layout = "large"
	scene := DashboardScene(metrics.Snapshot{}, cfg, time.Now())

	elements := make(map[string]SceneElement, len(scene.Elements))
	for _, element := range scene.Elements {
		elements[element.ID] = element
	}
	for index, id := range []string{"cpu-card", "memory-card", "storage-card"} {
		element := elements[id]
		if element.X != 72+index*600 || element.Y != 170 || element.Width != 576 || element.Height != 380 {
			t.Fatalf("%s bounds = %#v", id, element)
		}
	}
	for index, id := range []string{"network-card", "thermal-card", "display-card"} {
		element := elements[id]
		if element.X != 72+index*600 || element.Y != 584 || element.Width != 576 || element.Height != 320 {
			t.Fatalf("%s bounds = %#v", id, element)
		}
	}
	cpuDetail := elements["cpu-uptime-value"]
	cpuBar := elements["cpu-bar-track"]
	if cpuDetail.Y+cpuDetail.Height > cpuBar.Y {
		t.Fatalf("CPU detail overlaps usage bar: detail=%#v bar=%#v", cpuDetail, cpuBar)
	}
}

func TestDashboardKeepsStandardLayoutAt1080p(t *testing.T) {
	cfg := config.Default()
	snapshot := metrics.Snapshot{Display: metrics.Display{Modes: []string{"1920x1080"}}}
	data := renderDashboard(snapshot, cfg, time.Now())
	if strings.Contains(data, "\\fs180") {
		t.Fatal("1080p dashboard unexpectedly used the compact layout")
	}
	if !strings.Contains(data, "打开 ZimaOS") {
		t.Fatal("1080p dashboard is missing the standard control hint")
	}
}

func TestDashboardFontScaleChangesSceneText(t *testing.T) {
	cfg := config.Default()
	cfg.Dashboard.Layout = "large"
	cfg.Dashboard.FontScale = 2
	scene := DashboardScene(metrics.Snapshot{}, cfg, time.Now())
	for _, element := range scene.Elements {
		if element.ID == "address" {
			if element.FontSize != 64 {
				t.Fatalf("address font size = %d, want 64", element.FontSize)
			}
			return
		}
	}
	t.Fatal("address element was not rendered")
}

func TestCanvasSceneResolvesLiveWidgets(t *testing.T) {
	cfg := config.Default()
	snapshot := metrics.Snapshot{CPUPercent: 42, MemoryPercent: 63, TemperatureC: 55, Hostname: "zima", IPAddresses: []string{"192.168.1.8"}}
	scene := CanvasScene(snapshot, cfg, time.Date(2026, time.September, 5, 9, 7, 0, 0, time.UTC))
	data := RenderSceneASS(scene)
	for _, expected := range []string{"09:07", "CPU 42%", "内存 63%", "温度 55°C", "192.168.1.8", "zima"} {
		if !strings.Contains(data, expected) {
			t.Fatalf("canvas output is missing %q", expected)
		}
	}
	if len(scene.Elements) <= len(cfg.Canvas.Widgets) {
		t.Fatal("gradient background elements were not generated")
	}
}

func TestCanvasImageBackgroundLeavesOverlayTransparent(t *testing.T) {
	cfg := config.Default()
	cfg.Canvas.BackgroundType = "image"
	cfg.Canvas.BackgroundImage = "/DATA/background.png"
	scene := CanvasScene(metrics.Snapshot{}, cfg, time.Now())
	if !scene.Transparent || scene.BackgroundImage != cfg.Canvas.BackgroundImage {
		t.Fatalf("unexpected image background scene: %#v", scene)
	}
}
