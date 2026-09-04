package server

import (
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
