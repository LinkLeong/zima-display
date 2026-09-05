package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"zima-display/internal/config"
	"zima-display/internal/player"
)

func TestWithin(t *testing.T) {
	if !within("/DATA", "/DATA/Media/movie.mp4") {
		t.Fatal("expected child path to be within root")
	}
	if within("/DATA", "/DATA-other/movie.mp4") {
		t.Fatal("prefix-only match escaped root")
	}
}

func TestResolveTargetAllowsSupportedNetworkMedia(t *testing.T) {
	cfg := config.Default()
	for _, target := range []string{"https://example.com/live.m3u8", "rtsp://camera.local/stream"} {
		resolved, err := resolveTarget(cfg, target)
		if err != nil || resolved != target {
			t.Fatalf("resolveTarget(%q) = %q, %v", target, resolved, err)
		}
	}
	if _, err := resolveTarget(cfg, "file:///etc/passwd"); err == nil {
		t.Fatal("expected file URL to be rejected")
	}
}

func TestMediaKind(t *testing.T) {
	if mediaKind("movie.MKV", false) != "video" || mediaKind("photo.webp", false) != "image" || mediaKind("notes.txt", false) != "" {
		t.Fatal("media type detection returned an unexpected result")
	}
}

func TestHealthEndpoint(t *testing.T) {
	path := filepath.Join(t.TempDir(), config.ConfigName)
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	controller := player.New(filepath.Join(t.TempDir(), "mpv.sock"), t.TempDir(), store.Get)
	api := New(log.New(io.Discard, "", 0), store, controller, "test-version")
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, apiPrefix+"/health", nil)
	api.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("health returned %d", recorder.Code)
	}
}

func TestMediaRootUsesLanguageNeutralUploadLabel(t *testing.T) {
	path := filepath.Join(t.TempDir(), config.ConfigName)
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	controller := player.New(filepath.Join(t.TempDir(), "mpv.sock"), t.TempDir(), store.Get)
	api := New(log.New(io.Discard, "", 0), store, controller, "test-version")
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, apiPrefix+"/media", nil)
	api.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("media returned %d", recorder.Code)
	}
	var response mediaResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Entries) == 0 || response.Entries[0].Label != "uploads" {
		t.Fatalf("upload root label = %#v, want uploads", response.Entries)
	}
}

func TestAutomationEndpointsRequireBearerToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), config.ConfigName)
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	controller := player.New(filepath.Join(t.TempDir(), "mpv.sock"), t.TempDir(), store.Get)
	api := New(log.New(io.Discard, "", 0), store, controller, "test-version")
	body := []byte(`{"title":"Demo","pages":["One"]}`)

	unauthorized := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, apiPrefix+"/v1/presentations/text", bytes.NewReader(body))
	api.Handler().ServeHTTP(unauthorized, request)
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized request returned %d", unauthorized.Code)
	}

	authorized := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, apiPrefix+"/v1/presentations/text", bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+store.Get().Automation.Token)
	api.Handler().ServeHTTP(authorized, request)
	if authorized.Code != http.StatusCreated {
		t.Fatalf("authorized request returned %d: %s", authorized.Code, authorized.Body.String())
	}
}

func TestStatusDoesNotExposeAutomationToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), config.ConfigName)
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	controller := player.New(filepath.Join(t.TempDir(), "mpv.sock"), t.TempDir(), store.Get)
	api := New(log.New(io.Discard, "", 0), store, controller, "test-version")
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, apiPrefix+"/status", nil)
	api.Handler().ServeHTTP(recorder, request)
	if bytes.Contains(recorder.Body.Bytes(), []byte(store.Get().Automation.Token)) {
		t.Fatal("status response exposed the automation token")
	}
}

func TestDashboardPreviewUsesDraftAppearance(t *testing.T) {
	path := filepath.Join(t.TempDir(), config.ConfigName)
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	controller := player.New(filepath.Join(t.TempDir(), "mpv.sock"), t.TempDir(), store.Get)
	api := New(log.New(io.Discard, "", 0), store, controller, "test-version")
	body := []byte(`{"dashboard":{"title":"Preview","language":"en-US","layout":"large","font_scale":1.5,"refresh_interval_ms":1000},"resolution":"1024x600"}`)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, apiPrefix+"/preview/dashboard", bytes.NewReader(body))
	api.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("preview returned %d: %s", recorder.Code, recorder.Body.String())
	}
	var scene player.Scene
	if err := json.Unmarshal(recorder.Body.Bytes(), &scene); err != nil {
		t.Fatal(err)
	}
	if scene.Width != 1920 || scene.Height != 1080 || len(scene.Elements) == 0 {
		t.Fatalf("unexpected preview scene: %#v", scene)
	}
	foundTitle := false
	for _, element := range scene.Elements {
		if element.ID == "title" && element.Text == "PREVIEW" {
			foundTitle = true
		}
	}
	if !foundTitle {
		t.Fatal("preview scene did not use draft dashboard settings")
	}
}

func TestCanvasPreviewUsesDraftLayout(t *testing.T) {
	path := filepath.Join(t.TempDir(), config.ConfigName)
	store, err := config.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	controller := player.New(filepath.Join(t.TempDir(), "mpv.sock"), t.TempDir(), store.Get)
	api := New(log.New(io.Discard, "", 0), store, controller, "test-version")
	canvas := store.Get().Canvas
	canvas.Widgets = []config.CanvasWidget{{ID: "message", Type: "text", Text: "Hello Canvas", X: 100, Y: 120, Width: 800, Height: 100, FontSize: 64, Color: "#ffffff", Bold: true}}
	payload, err := json.Marshal(canvasPreviewRequest{Canvas: canvas})
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, apiPrefix+"/preview/canvas", bytes.NewReader(payload))
	api.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("canvas preview returned %d: %s", recorder.Code, recorder.Body.String())
	}
	var scene player.Scene
	if err := json.Unmarshal(recorder.Body.Bytes(), &scene); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, element := range scene.Elements {
		if element.WidgetID == "message" && element.Text == "Hello Canvas" {
			found = true
		}
	}
	if !found {
		t.Fatal("canvas preview did not render the draft widget")
	}
}
