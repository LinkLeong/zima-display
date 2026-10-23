// Package boot installs a recovery path that survives an unmerged /usr.
package boot

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed recover-extension.sh
var recoveryScript []byte

type installedFile struct {
	path    string
	content []byte
	mode    os.FileMode
}

func watchdogFiles() []installedFile {
	return []installedFile{
		{"/etc/zima-display/recover-extension.sh", recoveryScript, 0o755},
		{"/etc/systemd/system/zima-display-watchdog.service", []byte(`[Unit]
Description=Recover Zima Display after persistent storage and system extensions are ready
After=local-fs.target casaos-bind.target systemd-sysext.service
StartLimitIntervalSec=0

[Service]
Type=oneshot
ExecStart=/etc/zima-display/recover-extension.sh
RemainAfterExit=yes
Restart=on-failure
RestartSec=15
TimeoutStartSec=60
`), 0o644},
		{"/etc/systemd/system/zima-display-watchdog.timer", []byte(`[Unit]
Description=Ensure Zima Display starts after boot

[Timer]
OnBootSec=45
AccuracySec=1

[Install]
WantedBy=timers.target
`), 0o644},
		{"/etc/systemd/system/zima-display-refresh.path", []byte(`[Unit]
Description=Watch for Zima Display extension updates

[Path]
PathChanged=/usr/bin/zima-displayd

[Install]
WantedBy=multi-user.target
`), 0o644},
		{"/etc/systemd/system/zima-display-refresh.service", []byte(`[Unit]
Description=Restart Zima Display after an extension update

[Service]
Type=oneshot
ExecStart=/bin/sh -c 'sleep 2; systemctl restart zima-display.service'
`), 0o644},
	}
}

// Install provisions recovery only when running as root under systemd.
func Install(logger *log.Logger) {
	if os.Geteuid() != 0 {
		return
	}
	if _, err := os.Stat("/run/systemd/system"); err != nil {
		return
	}
	run := func(args ...string) error { return exec.Command("systemctl", args...).Run() }
	if err := install("/", run); err != nil {
		logger.Printf("install boot recovery: %v", err)
	}
}

func install(root string, run func(...string) error) error {
	for _, file := range watchdogFiles() {
		path := filepath.Join(root, file.path)
		current, err := os.ReadFile(path)
		info, statErr := os.Stat(path)
		if err == nil && statErr == nil && string(current) == string(file.content) && info.Mode().Perm() == file.mode {
			continue
		}
		if err := replaceFile(path, file.content, file.mode); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	// Reconcile the manager even when files match: the previous process might
	// have exited after writing the files but before reloading or enabling them.
	if err := run("daemon-reload"); err != nil {
		return fmt.Errorf("reload systemd: %w", err)
	}
	for _, unit := range []string{"zima-display-watchdog.timer", "zima-display-refresh.path"} {
		if err := run("enable", "--now", unit); err != nil {
			return fmt.Errorf("enable %s: %w", unit, err)
		}
	}
	return nil
}

func replaceFile(path string, content []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".zima-display-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(content); err != nil {
		file.Close()
		return err
	}
	if err := file.Chmod(mode); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
