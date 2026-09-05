package dsh

import (
	"crypto/subtle"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const skillName = "zima-display"

//go:embed assets/zima-display/SKILL.md
var assets embed.FS

type Installer struct {
	binaryPath string
}

type Status struct {
	Available      bool   `json:"available"`
	Installed      bool   `json:"installed"`
	Version        string `json:"version,omitempty"`
	Path           string `json:"path,omitempty"`
	ConverterState string `json:"converter_state,omitempty"`
	Error          string `json:"error,omitempty"`
}

type Config struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
	Version string `json:"version"`
}

func New(binaryPath string) *Installer {
	return &Installer{binaryPath: binaryPath}
}

func (i *Installer) Status() Status {
	home, err := dshHome()
	if err != nil {
		return Status{Error: err.Error()}
	}
	directory := filepath.Join(home, "skills", skillName)
	status := Status{Available: true, Path: directory, ConverterState: converterState(directory)}
	data, err := os.ReadFile(filepath.Join(directory, "config.json"))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			status.Error = err.Error()
		}
		return status
	}
	var cfg Config
	if json.Unmarshal(data, &cfg) == nil {
		status.Version = cfg.Version
	}
	status.Installed = true
	return status
}

func (i *Installer) Install(cfg Config) (Status, error) {
	if cfg.BaseURL == "" || cfg.Token == "" || cfg.Version == "" {
		return Status{}, errors.New("base_url, token and version are required")
	}
	home, err := dshHome()
	if err != nil {
		return Status{}, err
	}
	directory := filepath.Join(home, "skills", skillName)
	if err := os.MkdirAll(filepath.Join(directory, "scripts"), 0o755); err != nil {
		return Status{}, err
	}
	skill, err := assets.ReadFile("assets/zima-display/SKILL.md")
	if err != nil {
		return Status{}, err
	}
	if err := writeAtomic(filepath.Join(directory, "SKILL.md"), skill, 0o644); err != nil {
		return Status{}, err
	}
	configData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return Status{}, err
	}
	if err := writeAtomic(filepath.Join(directory, "config.json"), append(configData, '\n'), 0o600); err != nil {
		return Status{}, err
	}
	binary, err := os.ReadFile(i.binaryPath)
	if err != nil {
		return Status{}, fmt.Errorf("read zima-displayctl: %w", err)
	}
	if err := writeAtomic(filepath.Join(directory, "scripts", "zima-displayctl"), binary, 0o755); err != nil {
		return Status{}, err
	}
	if converterState(directory) == "missing" {
		if err := startConverterInstall(directory); err != nil {
			return Status{}, fmt.Errorf("install skill succeeded, but converter setup failed to start: %w", err)
		}
	}
	return i.Status(), nil
}

func converterState(directory string) string {
	if err := exec.Command("docker", "exec", "dsh-harness", "sh", "-lc", "command -v libreoffice >/dev/null && command -v pdftoppm >/dev/null").Run(); err == nil {
		return "ready"
	}
	if info, err := os.Stat(filepath.Join(directory, "converter-installing")); err == nil {
		if time.Since(info.ModTime()) < 30*time.Minute {
			return "installing"
		}
		_ = os.Remove(filepath.Join(directory, "converter-installing"))
	}
	return "missing"
}

func startConverterInstall(directory string) error {
	marker := filepath.Join(directory, "converter-installing")
	if err := os.WriteFile(marker, []byte("Installing LibreOffice and PDF conversion tools.\n"), 0o644); err != nil {
		return err
	}
	command := "{ trap 'rm -f /root/.dsh/skills/zima-display/converter-installing' EXIT; " +
		"export DEBIAN_FRONTEND=noninteractive; " +
		"apt-get update && apt-get install -y --no-install-recommends libreoffice-impress poppler-utils fonts-noto-cjk && " +
		"rm -rf /var/lib/apt/lists/*; } > /root/.dsh/skills/zima-display/converter-install.log 2>&1"
	if err := exec.Command("docker", "exec", "-d", "dsh-harness", "sh", "-lc", command).Run(); err != nil {
		_ = os.Remove(marker)
		return err
	}
	return nil
}

func (i *Installer) Uninstall() error {
	home, err := dshHome()
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(home, "skills", skillName))
}

func dshHome() (string, error) {
	output, err := exec.Command("docker", "inspect", "dsh-harness", "--format", "{{json .Mounts}}").Output()
	if err != nil {
		return "", errors.New("DeepSeek Harness container dsh-harness is not running or is not installed")
	}
	var mounts []struct {
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
		RW          bool   `json:"RW"`
	}
	if err := json.Unmarshal(output, &mounts); err != nil {
		return "", fmt.Errorf("decode DSH mount information: %w", err)
	}
	for _, mount := range mounts {
		if mount.Destination != "/root/.dsh" || !mount.RW {
			continue
		}
		clean := filepath.Clean(mount.Source)
		if clean != "/DATA/AppData" && !strings.HasPrefix(clean, "/DATA/AppData/") {
			return "", errors.New("DSH persistent directory is outside /DATA/AppData")
		}
		return clean, nil
	}
	return "", errors.New("DSH does not expose a writable /root/.dsh persistent volume")
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".zima-display-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func TokenMatches(expected, actual string) bool {
	if expected == "" || len(expected) != len(actual) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}
