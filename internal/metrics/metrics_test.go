package metrics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkipInterface(t *testing.T) {
	for _, name := range []string{"lo", "docker0", "veth12", "br-home", "br0", "virbr0", "tailscale0", "ztabc", "tun0", "wg0", "bond0"} {
		if !skipInterface(name) {
			t.Fatalf("expected %q to be skipped", name)
		}
	}
	if skipInterface("eth0") {
		t.Fatal("physical interface was skipped")
	}
}

func TestDisplayInterfacePriority(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{
		filepath.Join(root, "enp1s0", "device"),
		filepath.Join(root, "wlan0", "wireless"),
		filepath.Join(root, "docker0", "device"),
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	for _, test := range []struct {
		name     string
		priority int
		allowed  bool
	}{
		{name: "enp1s0", priority: 0, allowed: true},
		{name: "thunderbolt0", priority: 1, allowed: true},
		{name: "wlan0", priority: 2, allowed: true},
		{name: "docker0", allowed: false},
		{name: "br0", allowed: false},
		{name: "tun0", allowed: false},
		{name: "eth0", allowed: false},
	} {
		priority, allowed := displayInterfacePriority(test.name, root)
		if priority != test.priority || allowed != test.allowed {
			t.Fatalf("displayInterfacePriority(%q) = (%d, %v), want (%d, %v)", test.name, priority, allowed, test.priority, test.allowed)
		}
	}
}

func TestDisplayInterfaceFallbackOutsideLinux(t *testing.T) {
	missingRoot := filepath.Join(t.TempDir(), "missing")
	if priority, allowed := displayInterfacePriority("en0", missingRoot); !allowed || priority != 0 {
		t.Fatalf("en0 fallback = (%d, %v), want (0, true)", priority, allowed)
	}
	if _, allowed := displayInterfacePriority("utun0", missingRoot); allowed {
		t.Fatal("virtual utun interface was allowed")
	}
}

func TestRound(t *testing.T) {
	if value := round(12.345, 1); value != 12.3 {
		t.Fatalf("round returned %v", value)
	}
}
