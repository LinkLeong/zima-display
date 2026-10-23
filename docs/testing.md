# Testing Standard

Zima Display uses layered release gates. Passing a lower layer never proves that a higher layer works: unit tests do not prove package correctness, and package checks do not prove HDMI output on real ZimaOS hardware.

## Release rules

1. Test the exact commit and package that will be released. Record the commit, version and SHA-256 checksum.
2. A renderer, packaging or display-setting change must pass CI, package verification and real-device HDMI acceptance.
3. A user-visible workflow must be tested from the Web UI, not only by calling its underlying Go function.
4. A bug is closed only by a regression test that fails without the fix and by repeating the original user workflow.
5. Any P0/P1 failure blocks release. A skipped required test is a failure unless the release note names the owner and reason.
6. Credentials, private IP logs, tokens and uploaded media never enter test artifacts or the repository.

## Release gates

| Gate | Environment | Required evidence | Release condition |
| --- | --- | --- | --- |
| G1 Static and unit | Developer machine and GitHub CI | Command log and test report | Every automated check passes |
| G2 Package | Built `.raw` artifact | Version, file inventory and source/package comparison | The tested artifact is internally consistent |
| G3 API and browser | Running development service | API assertions, desktop/mobile screenshots and browser console | Primary workflows work in both locales and supported widths |
| G4 ZimaOS HDMI | Physical machine and display | Device script output, package checksum, environment record and visual confirmation | Renderer, recovery and mode switching pass on hardware |
| G5 Release soak | Physical machine | 10-minute IPC sampling and journal excerpt | No renderer loss, restart loop or user-visible IPC error |

## Standard automated run

```bash
gofmt -w ./cmd ./internal
go test ./...
node --check web/app.js
node scripts/check-i18n.mjs
node scripts/test-web-language.mjs
node scripts/test-home-ui.mjs
node scripts/test-settings-ui.mjs
sh -n scripts/*.sh packaging/raw/usr/libexec/zima-display/start-renderer
./scripts/test-renderer-fallback.sh
./scripts/build.sh amd64
./scripts/test-package.sh
```

GitHub CI runs the same checks except for physical-device tests.

## Test catalog

Automation values: **CI** runs on every push, **Device** runs through `test-device-renderer.sh`, **Manual** requires visual or hardware confirmation, and **Planned** identifies the next automation gap.

### Configuration and API

| ID | Check | How to check | Pass criteria | Automation |
| --- | --- | --- | --- | --- |
| CFG-01 | First-run defaults | Create a config in a temporary data directory | DRM, automatic resolution/audio, dashboard mode and safe media root are created | CI |
| CFG-02 | Existing config migration | Load configs missing language, canvas, media playback or newer renderer fields | Missing values receive defaults without losing existing values | CI |
| CFG-03 | Configuration persistence | PUT changed mode, backend, resolution and audio device; reload from disk | Saved and reloaded values match | CI |
| CFG-04 | Token preservation | PUT the public configuration without an automation token | Existing secret token remains unchanged | CI |
| CFG-05 | Invalid media roots | PUT `/etc`, relative paths and prefix-confusion paths | API returns 400 and stored config is unchanged | CI |
| CFG-06 | Invalid renderer values | Test unknown backend, invalid resolution state and malformed volume | Validation rejects unsupported values | CI/Planned |
| CFG-07 | Dashboard bounds | Test language, layout, font scale and refresh interval boundaries | Valid edges pass; values outside limits fail | CI |
| CFG-08 | Canvas bounds | Test duplicate IDs, unsupported widgets, invalid colors and off-canvas geometry | Invalid canvas is rejected without partial persistence | CI |
| CFG-09 | Remote source validation | Test duplicate/invalid IDs, missing connection fields, port/interval bounds and unknown canvas references | Invalid sources and references are rejected; old configs still load | CI |
| CFG-10 | Media playback defaults | Validate single-video end behavior, collection end behavior and slideshow interval choices | Repeat/stop and 3/5/8/10/15/30 seconds are accepted; unsupported values are rejected | CI |
| API-01 | Health | GET `/health` | HTTP 200 with expected version | CI/Device |
| API-02 | Public status privacy | GET `/status` | Metrics/config/player returned; automation token absent | CI |
| API-03 | Method handling | Send unsupported methods to every endpoint | HTTP 405 with no state change | Planned |
| API-04 | Invalid JSON/body size | Send malformed and oversized requests | HTTP 400/413; service remains responsive | Planned |
| API-05 | Concurrent polling and writes | Poll status while repeatedly saving valid settings | No data race, corruption or intermittent 5xx | Planned |

### Home guidance and content workflows

| ID | Check | How to check | Pass criteria | Automation |
| --- | --- | --- | --- | --- |
| HOME-01 | Unique starting decision | Open the app as a first-time user | The page asks what to show and presents exactly Ready-made, Canvas and Media paths | CI contract + Manual |
| HOME-02 | Mutual exclusion | Select each content card in turn | Only the matching workspace is visible; canvas and media tools never compete on screen | CI contract + Manual |
| HOME-03 | Ready-made path | Choose Ready-made, then Dashboard and Clock | Both choices apply immediately and current-output status updates | CI contract + Device |
| HOME-04 | Canvas templates | Choose Canvas and apply each template | Three template starting points remain available and the editor stays fully editable | CI contract + Manual |
| HOME-05 | Media path | Choose Media, browse/upload/select content and play | Only media controls are shown and playback starts with clear feedback | Manual/Device |
| HOME-06 | On-demand live output | Open and close Current screen in dashboard, canvas and media modes | Dashboard/canvas use the renderer scene; media modes show a frame captured from the active mpv output; no metrics, playback or maintenance controls appear | CI contract + Device |
| HOME-07 | Control placement | Inspect Media and Settings | Playback controls appear only in Media; Black standby and Restore terminal appear only under Settings > Screen maintenance | CI contract + Manual |
| HOME-08 | Product information | Open Settings and inspect About | Installed version and GitHub source/issue link are visible and correct | CI contract + Manual |
| HOME-09 | Dashboard editing placement | Open Ready-made views and Settings | Dashboard title/layout/type controls and preview exist on the main page only; Settings contains only the startup mode choice | CI contract + Manual |

### Settings UI and usability

| ID | Check | How to check | Pass criteria | Automation |
| --- | --- | --- | --- | --- |
| UI-01 | Settings identity | Open Settings without prior product knowledge | Header clearly says Settings; current output and future startup behavior are described as different concepts | CI contract + Manual |
| UI-02 | Recommended defaults | Open a first-run config | DRM is selected and visually marked recommended | CI |
| UI-03 | Device summary | Open Settings with connected/disconnected states | Connector, current mode and renderer readiness match `/status` | Manual/Planned |
| UI-04 | Resolution discovery | Provide repeated DRM modes in status | Select lists `auto` plus unique reported modes and preserves configured custom mode | CI |
| UI-05 | Complete mapping | Change every Settings field and save | Backend, resolution, audio, startup, playback defaults and media roots map to correct JSON without modifying dashboard appearance | CI |
| UI-06 | Progressive disclosure | Open Settings and inspect the initial view, then expand both disclosures | Renderer/audio/media details are hidden initially and remain reachable without losing the draft | CI contract + Manual |
| UI-07 | Non-disruptive save | Select a startup view different from the active view and submit | Configuration is saved without mode actions, renderer restart, black frame or change to current output | CI contract + Device |
| UI-08 | Unsaved state | Modify and reset fields | Unsaved badge and footer state accurately follow the draft | Manual/Planned |
| UI-09 | Clear disclosure names | Inspect all collapsed Settings sections | Every title names the setting domain (for example Display output); no title asks the user a support question | CI contract + Manual |
| UI-10 | Locale completeness | Run i18n checker and switch languages | No missing keys; Web and HDMI language remain synchronized | CI |
| UI-11 | Keyboard access | Tab through controls and toggle disclosure sections | Focus is visible; selects, details, radios, sliders and buttons work without a pointer | Manual/Planned |
| UI-12 | Screen reader semantics | Inspect dialog, status summary, fields, disclosure summaries and live status | Names, roles, expanded states and relationships are announced correctly | Manual/Planned |
| UI-13 | Desktop layout | Capture at 1440x1000 and 1024x768 | No clipping, overlap or hidden primary action | Manual |
| UI-14 | Mobile layout | Capture at 390x844 and 320x568 | Primary choice appears first; disclosures and footer action remain usable | Manual |
| UI-15 | Long content | Test long English labels, 64-character title and long paths | Text wraps or truncates without covering controls | Manual/Planned |
| UI-16 | Error feedback | Force 400, offline and renderer restart failure | Error is visible, specific and does not close the draft unexpectedly | Manual/Planned |
| UI-17 | Remote device management | Add a Linux device with valid/invalid SSH credentials, refresh, then delete | Valid device is pinned and listed with online state; invalid login changes nothing; deletion removes its secret | CI + Manual |
| UI-18 | Media settings hierarchy | Open Settings > Media playback | End behavior and slideshow timing appear first; media folders are clearly labeled as browser access scope | CI contract + Manual |

### Renderer and recovery

| ID | Check | How to check | Pass criteria | Automation |
| --- | --- | --- | --- | --- |
| REN-01 | Real IPC readiness | Accept a Unix connection without replying | Renderer is not reported ready | CI |
| REN-02 | Lost-command recovery | Drop the first real mpv command | Renderer restarts once and the safe operation is retried once | CI |
| REN-03 | Logical mpv error | Return an mpv command error while connection stays valid | No unnecessary service restart | CI |
| REN-04 | GPU start failure | Mock `gpu-next` exit before socket creation | Software DRM starts with `--vo=drm --hwdec=no` | CI |
| REN-05 | Boot-scoped attempt | Restart renderer with GPU-attempt marker present | GPU is not retried in the same boot | CI |
| REN-06 | Socket timeout | Keep mpv alive without creating its socket for 10 seconds | Process is terminated and fallback begins | CI/Planned |
| REN-07 | Display disconnected | Start with no connected DRM connector | Renderer waits without restart storms or false-ready status | Device/Manual |
| REN-08 | Hot-plug | Disconnect and reconnect HDMI | Output returns without reboot and status follows connector state | Manual |
| REN-09 | Kill recovery | Kill only the mpv process and request a mode | API succeeds and IPC returns within 20 seconds | Device |
| REN-10 | Service restart | Restart control and renderer services three times | Each cycle becomes IPC-ready within 20 seconds | Device |
| REN-11 | Reboot behavior | Reboot with auto-start enabled and disabled | Enabled starts configured mode; disabled leaves terminal | Manual |
| REN-12 | No user-visible IPC error | Exercise modes during recovery | UI never leaves `connection refused/reset` as a permanent error | Device + Manual |
| REN-13 | Exclusive tty1 ownership | Check the packaged unit; run `test-device-renderer.sh`, then reboot with ZimaOS Welcome enabled | Both console providers appear in renderer `Conflicts` and `After`; the delayed Welcome-to-dashboard handoff stays ready; Welcome/getty are stopped before renderer startup | CI package + Device + Manual reboot |

### Display modes, dashboard and canvas

| ID | Check | How to check | Pass criteria | Automation |
| --- | --- | --- | --- | --- |
| DSP-01 | Dashboard mode | Select dashboard | Generated video and overlay load; status mode is dashboard | CI/Device |
| DSP-02 | Clock mode | Select clock and observe minute change | Time/date update and remain within bounds | Device/Manual |
| DSP-03 | Black standby | Select black | Screen is black while renderer IPC stays ready | Device/Manual |
| DSP-04 | Terminal restore | Select terminal | Renderer stops and getty/TTY1 becomes available | Device/Manual |
| DSP-05 | Return from terminal | Select dashboard after terminal | Getty stops and renderer starts with current configuration | Device |
| DSP-06 | Compact display | Preview/test 1024x600 and 1280x720 | Automatic layout uses large type without overlap | CI + Manual |
| DSP-07 | Full HD/4K | Preview/test 1920x1080 and 3840x2160 | Standard layout remains aligned and legible | CI + Manual |
| DSP-08 | Font scale bounds | Test 80%, 100%, 150% and 200% | Text changes predictably without neighboring overlap | CI + Manual |
| DSP-09 | Dashboard editor workflow | Change title/layout/type on the main page, then Save and Save-and-show | Preview updates before save; Save persists without mode change; Save-and-show switches to dashboard | CI + Manual |
| CAN-01 | Canvas templates | Apply all built-in templates | Widgets and backgrounds match template definitions | Manual/Planned |
| CAN-02 | Drag and resize | Move/resize at edges and corners | Geometry remains inside 1920x1080 and matches saved values | Manual/Planned |
| CAN-03 | Widget editing | Edit text, colors, font and bold state | Preview and HDMI show the same saved result | CI + Manual |
| CAN-04 | Image background | Upload and apply supported image | Overlay remains transparent and image displays correctly | CI + Manual |
| CAN-05 | Widget data source | Select local and remote sources for every metric widget | Preview and HDMI use the selected device; clock/date/text never expose a source choice | CI + Manual |
| CAN-06 | Remote offline state | Stop SSH or disconnect a configured device | Its widgets show "Device offline" instead of stale values; local widgets continue updating | CI + Device |

### Remote data collection

| ID | Check | How to check | Pass criteria | Automation |
| --- | --- | --- | --- | --- |
| REM-01 | Linux metric parsing | Parse representative `/proc`, `df` and hostname output | CPU, memory, disk, network, temperature, uptime, IP and identity fields map correctly | CI |
| REM-02 | Counter deltas | Feed two CPU/network samples five seconds apart | Percent and bytes/second use counter deltas and never report negative reset values | CI |
| REM-03 | Secret isolation | Add a source, inspect config/status/API/logs and file mode | Password appears only in `source-secrets.json`; file mode is 0600 | CI + Device |
| REM-04 | Host identity | Reconnect with the same key, then simulate a changed SSH host key | Same key succeeds; changed key is rejected until the source is recreated | CI + Manual |
| REM-05 | Polling and recovery | Run up to 10 devices, disconnect/reconnect one | Collection remains responsive; failures back off; recovered data replaces offline state | Device/Planned |
| REM-06 | Privilege requirement | Collect with a normal SSH user without sudo | All supported metrics are readable and no elevation is requested | Manual |

### Media and playback

| ID | Check | How to check | Pass criteria | Automation |
| --- | --- | --- | --- | --- |
| MED-01 | Root browsing | Open each configured root and parent navigation | Only allowed paths are reachable | CI/Manual |
| MED-02 | Path traversal | Try `..`, symlinks and prefix-confusion paths | Access outside configured roots is rejected | CI/Planned |
| MED-03 | Type filtering | Mix video, image, directory and unsupported files | Only supported entries are shown/playable | CI |
| MED-04 | Single-video end policy | Select Repeat, then Stop at end, and play one local video through its end | Repeat restarts the video; Stop leaves it ended without restarting | CI/Device |
| MED-05 | Collection end policy | Select Restart, then Stop after last, and play multiple files or an image directory | Order is preserved; the collection either restarts or stops according to Settings | CI/Device |
| MED-06 | Image duration | Configure a default, start a slideshow without a request value, then override it in Media | Configured default is used unless the Media page supplies a one-time interval | CI/Manual |
| MED-07 | Slideshow ordering | Use mixed-case filenames and non-images | Images sort by filename; other files are ignored | CI |
| MED-08 | Network protocols | Test HTTP, HTTPS, HLS and RTSP fixtures | Supported URLs load or return a specific timeout/error | Manual/Planned |
| MED-09 | Upload | Upload valid, oversized and unsupported files | Valid media appears; limits and failures are explicit | Manual/Planned |
| MED-10 | Playback controls | Pause, resume, seek, next, previous, stop and volume | Player state and HDMI output match each action | Device/Manual |
| MED-11 | Current-screen media frame | Play a video and an image slideshow, then open Current screen | The preview is captured from the active mpv renderer and matches the HDMI frame instead of showing the generic clock placeholder | CI + Device/Manual |

### Presentation and integration

| ID | Check | How to check | Pass criteria | Automation |
| --- | --- | --- | --- | --- |
| PRE-01 | Auth | Call automation endpoints with missing, bad and valid tokens | 401 for invalid credentials; valid token succeeds | CI |
| PRE-02 | Markdown pagination | Present headings and long pages | Page boundaries are stable and readable | CI |
| PRE-03 | Image archive safety | Import ZIP with traversal paths | Archive is rejected and writes nothing outside target | CI |
| PRE-04 | Presentation navigation | Activate, next, previous and stop | Page index and HDMI content stay synchronized | Manual/Planned |
| PRE-05 | PPTX/PDF conversion | Test with and without converters | Full conversion works when installed; documented fallback/error otherwise | Manual |
| DSH-01 | Container discovery | Test running, stopped and missing DSH containers | UI accurately reports availability | Manual/Planned |
| DSH-02 | Skill install/update | Install twice and update after container upgrade | Operation is idempotent and restores missing converters | Manual |
| DSH-03 | Secret handling | Inspect status, logs and UI | Automation token is never exposed | CI + Manual |

### Package, install and upgrade

| ID | Check | How to check | Pass criteria | Automation |
| --- | --- | --- | --- | --- |
| PKG-01 | Required files | Extract `.raw` | Binaries, renderer, units and Web assets exist | CI |
| PKG-02 | Version consistency | Run/inspect both binaries | Both equal `VERSION` | CI |
| PKG-03 | Renderer identity | Compare packaged and source renderer scripts | Byte-for-byte identical and executable | CI |
| PKG-04 | Web asset version | Inspect packaged HTML/JS/CSS | Candidate assets are present and cache version is current | Planned |
| PKG-05 | Clean install | Install on a device without Zima Display | zpkg, sysext, Gateway route and services become ready | Device/Manual |
| PKG-06 | Forced upgrade | Upgrade from previous release with existing config/media | Data persists and new version becomes active | Device/Manual |
| PKG-07 | Reinstall | Install the same candidate twice | Second install succeeds without duplicate units/routes | Device/Manual |
| PKG-08 | Failed install recovery | Interrupt or reject an invalid package | Previous working version and user data remain recoverable | Manual/Planned |

### Security, performance and reliability

| ID | Check | How to check | Pass criteria | Automation |
| --- | --- | --- | --- | --- |
| SEC-01 | Config permissions | Inspect config and token files | Sensitive config is mode 0600 | CI/Device |
| SEC-02 | Upload and archive traversal | Use malicious filenames and ZIP entries | No write escapes the application directories | CI/Planned |
| SEC-03 | Media authorization | Request arbitrary local files | Files outside allowed roots are rejected | CI/Planned |
| SEC-04 | Command injection | Use quotes/metacharacters in audio device and paths | Values are safely quoted; no extra command runs | CI/Planned |
| SEC-05 | HTTP exposure | Inspect listeners and Gateway route | Control service remains loopback-only behind Gateway | Device |
| REL-01 | 10-minute renderer soak | Sample IPC every 5 seconds | Zero failed samples and no restart loop | Device |
| REL-02 | Repeated mode switching | Cycle dashboard/clock/black 30 times | Every request succeeds; memory and process count stay stable | Planned |
| REL-03 | Large media directory | Browse at least 5,000 entries | UI remains responsive and API completes within agreed limit | Planned |
| REL-04 | Large upload | Upload near the 20 GB limit | Progress/failure is controlled; service memory remains bounded | Manual/Planned |
| REL-05 | Browser reconnect | Restart service while UI is open | Status returns online automatically without reload | Manual/Planned |

## Physical-device procedure

Use a designated ZimaOS regression machine and keep its address and credentials outside the repository. Export its SSH destination as `ZIMA_DISPLAY_TEST_TARGET`.

1. Connect and power on the HDMI display.
2. Record ZimaOS version, kernel, GPU, connector and candidate checksum.
3. Install the exact candidate package with a clean install or forced upgrade as required.
4. Set backend to DRM and confirm the installed binary version.
5. Run the automated device acceptance:

```bash
EXPECTED_VERSION=$(tr -d '[:space:]' < VERSION)
ssh "$ZIMA_DISPLAY_TEST_TARGET" "sh -s -- $EXPECTED_VERSION" < scripts/test-device-renderer.sh
```

6. Visually confirm dashboard, clock, black standby, terminal and return-to-dashboard.
7. Check the Web UI during fault injection for permanent `connection refused` or `connection reset by peer` errors.
8. Run the release soak:

```bash
EXPECTED_VERSION=$(tr -d '[:space:]' < VERSION)
ssh "$ZIMA_DISPLAY_TEST_TARGET" "ZIMA_DISPLAY_STABILITY_SECONDS=600 sh -s -- $EXPECTED_VERSION" < scripts/test-device-renderer.sh
```

## Hardware matrix

At least the primary row is required for every renderer release. Run additional rows when hardware is available.

| Priority | GPU/platform | Backend | Resolution | Required scenarios |
| --- | --- | --- | --- | --- |
| Primary | Designated current hardware | DRM with software fallback | 1920x1080 | Install, modes, kill recovery, 10-minute soak |
| Compatibility | Intel iGPU | DRM | 4K and 1080p | Boot, hot-plug, video and audio |
| Compatibility | AMD iGPU | DRM | 1080p | Boot, dashboard, video and audio |
| Optional | Known-good Weston device | Wayland | Native mode | Boot, mode switching and recovery |
| Negative | HDMI disconnected | DRM | N/A | No false-ready state or restart storm |

## Test record

Attach this record to a release or renderer pull request:

```text
Commit:
Version:
Package SHA-256:
Tester/date:
Device/ZimaOS/kernel/GPU:
Display/connector/resolution:
Install path: clean | forced upgrade | reinstall
Backend/fallback observed:
G1 result:
G2 result:
G3 desktop/mobile result:
G4 device result:
G5 soak result:
Failed/skipped test IDs:
Logs/screenshots:
```

## Issue #1 regression rule

Issue #1 is not closed by the controller reconnect unit test alone. The same candidate must pass GPU-start failure simulation, package verification, real-device mpv kill/recovery, software-DRM observation and the 10-minute soak without a permanent UI IPC error.
