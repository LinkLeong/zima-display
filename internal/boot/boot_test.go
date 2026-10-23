package boot

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestInstallReconcilesInterruptedActivation(t *testing.T) {
	for _, failAt := range []string{"daemon-reload", "enable --now zima-display-watchdog.timer", "enable --now zima-display-refresh.path"} {
		t.Run(failAt, func(t *testing.T) {
			root := t.TempDir()
			if err := install(root, func(args ...string) error {
				if strings.Join(args, " ") == failAt {
					return errors.New("simulated interruption")
				}
				return nil
			}); err == nil {
				t.Fatal("interrupted install must report the failed activation")
			}

			var calls []string
			if err := install(root, func(args ...string) error {
				calls = append(calls, strings.Join(args, " "))
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			want := []string{"daemon-reload", "enable --now zima-display-watchdog.timer", "enable --now zima-display-refresh.path"}
			if !reflect.DeepEqual(calls, want) {
				t.Fatalf("unchanged files did not repair activation: got %v, want %v", calls, want)
			}
			for _, file := range watchdogFiles() {
				path := filepath.Join(root, file.path)
				info, err := os.Stat(path)
				if err != nil {
					t.Fatal(err)
				}
				if info.Mode().Perm() != file.mode {
					t.Fatalf("%s mode = %o, want %o", path, info.Mode().Perm(), file.mode)
				}
			}
		})
	}
}

type bootFixture struct {
	t      *testing.T
	root   string
	bin    string
	log    string
	helper string
	ready  bool
	active bool
	merged bool
}

func newBootFixture(t *testing.T) *bootFixture {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("system extension boot recovery requires Linux")
	}
	root := t.TempDir()
	f := &bootFixture{t: t, root: root, bin: filepath.Join(root, "bin"), log: filepath.Join(root, "calls"), helper: filepath.Join(root, "recover-extension.sh"), ready: true, merged: true}
	f.write(f.helper, string(recoveryScript), 0o755)
	f.write(filepath.Join(f.bin, "mountpoint"), "#!/bin/sh\n[ \"$MOCK_DATA_READY\" = yes ]\n", 0o755)
	f.write(filepath.Join(f.bin, "systemd-sysext"), `#!/bin/sh
printf 'sysext %s\n' "$*" >> "$MOCK_LOG"
[ "$MOCK_MERGE_SUCCESS" = yes ] || exit 1
mkdir -p "$ZIMA_DISPLAY_BOOT_ROOT/usr/bin"
touch "$ZIMA_DISPLAY_BOOT_ROOT/usr/bin/zima-displayd"
chmod 0755 "$ZIMA_DISPLAY_BOOT_ROOT/usr/bin/zima-displayd"
`, 0o755)
	f.write(filepath.Join(f.bin, "systemctl"), `#!/bin/sh
printf 'systemctl %s\n' "$*" >> "$MOCK_LOG"
if [ "$1" = is-active ]; then
    [ "$MOCK_SERVICE_ACTIVE" = yes ]
elif [ "$1" = start ]; then
    [ -x "$ZIMA_DISPLAY_BOOT_ROOT/usr/bin/zima-displayd" ] || exit 5
fi
`, 0o755)
	return f
}

func (f *bootFixture) write(path, data string, mode os.FileMode) {
	f.t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), mode); err != nil {
		f.t.Fatal(err)
	}
}

func (f *bootFixture) run() (string, error) {
	f.t.Helper()
	return f.runCommand(exec.Command("/bin/sh", f.helper))
}

func (f *bootFixture) runCommand(command *exec.Cmd) (string, error) {
	f.t.Helper()
	yesNo := func(value bool) string {
		if value {
			return "yes"
		}
		return "no"
	}
	command.Env = append(os.Environ(),
		"PATH="+f.bin+":"+os.Getenv("PATH"),
		"ZIMA_DISPLAY_BOOT_ROOT="+f.root,
		"MOCK_LOG="+f.log,
		"MOCK_DATA_READY="+yesNo(f.ready),
		"MOCK_SERVICE_ACTIVE="+yesNo(f.active),
		"MOCK_MERGE_SUCCESS="+yesNo(f.merged),
	)
	output, err := command.CombinedOutput()
	return string(output), err
}

func (f *bootFixture) calls() string {
	f.t.Helper()
	data, err := os.ReadFile(f.log)
	if err != nil && !os.IsNotExist(err) {
		f.t.Fatal(err)
	}
	return string(data)
}

func TestRecoverMissingExtensionAfterBoot(t *testing.T) {
	for _, location := range []string{"DATA/.extensions", "var/lib/casaos_data/.extensions"} {
		t.Run(location, func(t *testing.T) {
			f := newBootFixture(t)
			image := filepath.Join(f.root, location, "zima-display.raw")
			f.write(image, "official image", 0o644)
			// Reproduce the old watchdog: systemctl cannot start a unit that is
			// still inside the unmerged package, even though the image persists.
			legacy := exec.Command("/bin/sh", "-c", "systemctl is-active --quiet zima-display.service || systemctl start zima-display.service")
			if _, err := f.runCommand(legacy); err == nil {
				t.Fatal("fixture did not reproduce the old missing-unit boot failure")
			}
			f.write(f.log, "", 0o644)
			if output, err := f.run(); err != nil {
				t.Fatalf("boot recovery failed: %v\n%s", err, output)
			}
			link, err := os.Readlink(filepath.Join(f.root, "run/extensions/zima-display.raw"))
			if err != nil || link != image {
				t.Fatalf("persistent image was not registered: link=%q err=%v", link, err)
			}
			want := "sysext refresh\nsystemctl is-active --quiet zima-display.service\nsystemctl daemon-reload\nsystemctl start zima-display.service\n"
			if got := f.calls(); got != want {
				t.Fatalf("boot sequence = %q, want %q", got, want)
			}
			// A second invocation must not remerge /usr or restart a running app.
			f.active = true
			f.write(f.log, "", 0o644)
			if output, err := f.run(); err != nil {
				t.Fatalf("repeat boot check failed: %v\n%s", err, output)
			}
			if got := f.calls(); got != "systemctl is-active --quiet zima-display.service\n" {
				t.Fatalf("running app was disrupted: %q", got)
			}
		})
	}
}

func TestRecoveryWaitsForStorageAndPackage(t *testing.T) {
	for _, missing := range []string{"storage", "package"} {
		t.Run(missing, func(t *testing.T) {
			f := newBootFixture(t)
			if missing == "storage" {
				f.ready = false
				f.write(filepath.Join(f.root, "DATA/.extensions/zima-display.raw"), "image", 0o644)
			}
			if _, err := f.run(); err == nil {
				t.Fatal("unavailable storage/package must fail so systemd can retry")
			}
			if got := f.calls(); got != "" {
				t.Fatalf("started services before storage/package was ready: %q", got)
			}
		})
	}
}

func TestRecoveryReloadsUnitsWhenAnotherProcessMergedTheImage(t *testing.T) {
	f := newBootFixture(t)
	f.write(filepath.Join(f.root, "usr/bin/zima-displayd"), "daemon", 0o755)
	if output, err := f.run(); err != nil {
		t.Fatalf("recovery failed after external merge: %v\n%s", err, output)
	}
	want := "systemctl is-active --quiet zima-display.service\nsystemctl daemon-reload\nsystemctl start zima-display.service\n"
	if got := f.calls(); got != want {
		t.Fatalf("cached unit lookup was not refreshed: got %q, want %q", got, want)
	}
}

func TestRecoveryDoesNotStartServiceAfterFailedMerge(t *testing.T) {
	f := newBootFixture(t)
	f.merged = false
	f.write(filepath.Join(f.root, "DATA/.extensions/zima-display.raw"), "image", 0o644)
	if _, err := f.run(); err == nil {
		t.Fatal("merge failure must trigger a retry")
	}
	if got := f.calls(); got != "sysext refresh\n" {
		t.Fatalf("service was started after a failed merge: %q", got)
	}
}

func TestRecoveryRejectsConflictingRuntimeImage(t *testing.T) {
	f := newBootFixture(t)
	f.write(filepath.Join(f.root, "DATA/.extensions/zima-display.raw"), "image", 0o644)
	link := filepath.Join(f.root, "run/extensions/zima-display.raw")
	f.write(link, "different image", 0o644)
	if _, err := f.run(); err == nil {
		t.Fatal("conflicting image must not be overwritten")
	}
	data, err := os.ReadFile(link)
	if err != nil || string(data) != "different image" {
		t.Fatalf("conflicting image was modified: %q, %v", data, err)
	}
	if got := f.calls(); got != "" {
		t.Fatalf("unexpected merge/start after image conflict: %q", got)
	}
}
