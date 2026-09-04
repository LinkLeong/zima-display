package metrics

import "testing"

func TestSkipInterface(t *testing.T) {
	for _, name := range []string{"lo", "docker0", "veth12", "br-home", "virbr0", "tailscale0", "ztabc"} {
		if !skipInterface(name) {
			t.Fatalf("expected %q to be skipped", name)
		}
	}
	if skipInterface("eth0") {
		t.Fatal("physical interface was skipped")
	}
}

func TestRound(t *testing.T) {
	if value := round(12.345, 1); value != 12.3 {
		t.Fatalf("round returned %v", value)
	}
}
