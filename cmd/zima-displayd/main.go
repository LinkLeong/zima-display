package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"zima-display/internal/config"
	"zima-display/internal/gateway"
	"zima-display/internal/metrics"
	"zima-display/internal/player"
	"zima-display/internal/server"
)

var version = "dev"

func main() {
	dataDir := flag.String("data-dir", envOr("ZIMA_DISPLAY_DATA_DIR", config.DefaultDataDir), "persistent data directory")
	listenAddress := flag.String("listen", envOr("ZIMA_DISPLAY_LISTEN", "127.0.0.1:0"), "HTTP listen address")
	runtimePath := flag.String("runtime-dir", envOr("ZIMA_DISPLAY_RUNTIME_DIR", "/run/zima-display"), "runtime directory")
	casaRuntimePath := flag.String("casa-runtime-dir", envOr("CASAOS_RUNTIME_PATH", "/var/run/casaos"), "CasaOS runtime directory")
	webDir := flag.String("web-dir", envOr("ZIMA_DISPLAY_WEB_DIR", ""), "optional standalone web directory")
	showVersion := flag.Bool("version", false, "print version")
	flag.Parse()
	if *showVersion {
		fmt.Println(version)
		return
	}

	logger := log.New(os.Stdout, "[zima-display] ", log.LstdFlags|log.Lmsgprefix)
	if err := os.MkdirAll(*dataDir, 0o755); err != nil {
		logger.Fatalf("create data directory: %v", err)
	}
	for _, directory := range []string{filepath.Join(*dataDir, "media"), filepath.Join(*dataDir, "runtime"), filepath.Join(*dataDir, "logs"), *runtimePath} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			logger.Fatalf("create %s: %v", directory, err)
		}
	}

	store, err := config.NewStore(config.Path(*dataDir))
	if err != nil {
		logger.Fatalf("load configuration: %v", err)
	}
	controller := player.New(filepath.Join(*runtimePath, "mpv.sock"), *runtimePath, store.Get)
	api := server.New(logger, store, controller, version)
	collector := metrics.NewCollector(*dataDir)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	go collectMetrics(ctx, logger, store, collector, api, controller)

	mux := http.NewServeMux()
	mux.Handle("/zima-display/api/", api.Handler())
	if *webDir != "" {
		mux.Handle("/", http.FileServer(http.Dir(*webDir)))
	}

	listener, err := net.Listen("tcp", *listenAddress)
	if err != nil {
		logger.Fatalf("listen: %v", err)
	}
	logger.Printf("version %s listening on http://%s", version, listener.Addr())
	gateway.RegisterAsync(ctx, logger, *casaRuntimePath, "/zima-display", "http://"+listener.Addr().String())
	installBootWatchdog(logger)

	if _, err := os.Stat("/run/systemd/system"); err == nil && store.Get().AutoStart {
		go func() {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
			startCtx, startCancel := context.WithTimeout(ctx, 30*time.Second)
			defer startCancel()
			if err := controller.SetMode(startCtx, store.Get().DefaultMode); err != nil {
				logger.Printf("automatic display start failed: %v", err)
			}
		}()
	}

	httpServer := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serveErrors := make(chan error, 1)
	go func() {
		serveErrors <- httpServer.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		logger.Printf("shutdown requested")
	case err := <-serveErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("HTTP server stopped: %v", err)
		}
	}
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Printf("HTTP shutdown: %v", err)
	}
}

func collectMetrics(ctx context.Context, logger *log.Logger, store *config.Store, collector *metrics.Collector, api *server.Server, controller *player.Manager) {
	for {
		snapshot := collector.Sample(ctx)
		api.SetMetrics(snapshot)
		controller.UpdateMetrics(ctx, snapshot)
		path := filepath.Join(store.Get().DataDir, "runtime", "metrics.json")
		if err := metrics.WriteSnapshot(path, snapshot); err != nil {
			logger.Printf("write metrics snapshot: %v", err)
		}
		interval := time.Duration(store.Get().Dashboard.RefreshIntervalMS) * time.Millisecond
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}

func installBootWatchdog(logger *log.Logger) {
	if os.Geteuid() != 0 {
		return
	}
	if _, err := os.Stat("/run/systemd/system"); err != nil {
		return
	}
	units := map[string]string{
		"/etc/systemd/system/zima-display-watchdog.service": `[Unit]
Description=Start Zima Display after system extensions are merged
After=systemd-sysext.service

[Service]
Type=oneshot
ExecStart=/bin/sh -c 'systemctl is-active --quiet zima-display.service || systemctl start zima-display.service'
`,
		"/etc/systemd/system/zima-display-watchdog.timer": `[Unit]
Description=Ensure Zima Display starts after boot

[Timer]
OnBootSec=15

[Install]
WantedBy=timers.target
`,
		"/etc/systemd/system/zima-display-refresh.path": `[Unit]
Description=Watch for Zima Display extension updates

[Path]
PathChanged=/usr/bin/zima-displayd

[Install]
WantedBy=multi-user.target
`,
		"/etc/systemd/system/zima-display-refresh.service": `[Unit]
Description=Restart Zima Display after an extension update

[Service]
Type=oneshot
ExecStart=/bin/sh -c 'sleep 2; systemctl restart zima-display.service'
`,
	}
	changed := false
	for path, content := range units {
		current, err := os.ReadFile(path)
		if err == nil && string(current) == content {
			continue
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			logger.Printf("install watchdog unit %s: %v", path, err)
			return
		}
		changed = true
	}
	if !changed {
		return
	}
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		logger.Printf("reload systemd after watchdog install: %v", err)
		return
	}
	for _, unit := range []string{"zima-display-watchdog.timer", "zima-display-refresh.path"} {
		if err := exec.Command("systemctl", "enable", "--now", unit).Run(); err != nil {
			logger.Printf("enable %s: %v", unit, err)
		}
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
