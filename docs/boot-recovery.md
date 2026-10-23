# Boot recovery

## Failure addressed

On ZimaOS, `/var` may be a temporary filesystem. A locally installed package can
remain in `/DATA/.extensions/zima-display.raw` after reboot while
`/var/lib/extensions` is not bound to the persistent directory. The daemon,
renderer script, Web assets and main service unit are then absent from `/usr`.
A watchdog that only runs `systemctl start zima-display.service` cannot recover
this state: the unit it wants to start is inside the unmerged image.

The observed host also reported a boot ordering cycle involving a data-disk swap
file and the data mount. This is a host configuration issue; this application
does not edit `fstab`, swap units, or storage mounts to work around it.

## Recovery shipped with the application

The first daemon start installs an embedded shell helper at
`/etc/zima-display/recover-extension.sh`, outside the extension's `/usr` tree.
It also installs and enables `zima-display-watchdog.timer` and the existing
extension-refresh watcher. Files are replaced atomically. Every daemon start
reloads and reconciles the watchers, even if their files already match, so an
interrupted installation does not leave recovery permanently disabled.

At approximately 45 seconds after boot, the watchdog:

1. Waits until `/DATA` is mounted. A failure is retried by systemd after 15 seconds.
2. If the daemon is missing, locates the existing local `zima-display.raw` in
   `/DATA/.extensions` or `/var/lib/casaos_data/.extensions`.
3. Registers that image through a symlink in `/run/extensions`, a standard
   `systemd-sysext` discovery directory. Conflicting files are not overwritten.
4. Refreshes the extensions, reloads systemd and starts the official main service.
5. Leaves an already running service alone. The daemon still owns the configured
   startup mode and the user's terminal/display preference.

The helper does not download packages or reinstall media/configuration. If the
persistent package has been removed, recovery reports a failure instead of
silently installing a different version. Successful boot recovery is oneshot;
it does not repeatedly force a stopped renderer or selected terminal mode back on.

The packaged renderer also sets `TTYVHangup=no`. Its supervisor and mpv own tty1;
hanging up that terminal can terminate them while systemd treats the exit as
successful. The source unit carries this setting without a device-specific
drop-in or a replacement renderer script.

`TTYVHangup=no` alone is insufficient when another service owns tty1. ZimaOS
also starts `zimaos-welcome@tty1.service`, a `Type=idle` console provider with
`TTYVHangup=yes`. It can execute late in boot, after the renderer is ready, and
hang up the renderer's terminal. This was reproduced on the physical host by
restarting Welcome while the renderer was active: the renderer exited cleanly
and its mpv socket refused connections. Disabling hangup only in the renderer
does not change Welcome's behavior.

The renderer now declares both `Conflicts=` and `After=` for
`getty@tty1.service` and `zimaos-welcome@tty1.service`. Systemd stops either console
provider, including its pending idle startup, before starting the renderer. The
ordering is important: a conflict alone could allow the old console's shutdown
and terminal reset to overlap renderer startup. These console services are not
masked or permanently disabled. Explicit terminal mode can still start getty.

## Verification

Automated regressions in `internal/boot` reproduce a missing unit with a persistent
image and show that the former start-only watchdog fails. They check recovery,
storage delay, failed merges, conflicting images, non-disruption of a running
service and interrupted activation. Package verification checks the actual
shipped renderer unit, its terminal setting, and both console conflicts and
their stop-before-start ordering.

These tests do not replace physical reboot acceptance. On a designated ZimaOS
test device, install the exact built package, confirm the persisted helper and
enabled timer, reboot, then verify after the recovery delay:

```bash
systemctl status zima-display-watchdog.service zima-display.service zima-display-renderer.service
systemctl is-enabled zima-display-watchdog.timer
systemd-sysext status
curl -fsS http://127.0.0.1/zima-display/api/status
journalctl -b -u zima-display-watchdog.service -u zima-display.service -u zima-display-renderer.service --no-pager
```

Require `renderer_ready: true` when automatic non-terminal output is configured
and HDMI is connected. Repeat with delayed HDMI connection and with automatic
display disabled, and complete the release soak in [testing.md](testing.md).
Use a dedicated machine for missing-registration fault injection; do not
unmerge all system extensions on an active NAS merely to simulate this failure.
