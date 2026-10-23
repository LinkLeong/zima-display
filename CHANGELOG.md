# Changelog

All notable changes to this project are documented in this file. The project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Fixed

- Recovered the persistent local package after reboot when ZimaOS does not restore its volatile system-extension registration. A standalone helper in `/etc` waits for storage, registers only Zima Display through `/run/extensions`, merges it and starts the control service.
- Retried boot recovery when storage or extension merging is not ready, and reconciled watchdog activation even after an interrupted install left unchanged unit files.
- Disabled terminal hangup in the HDMI renderer service to prevent the renderer/supervisor from exiting during terminal handover.
- Made the HDMI renderer conflict with and start after both tty1 console providers, including ZimaOS Welcome, so its delayed boot-time startup cannot hang up an already running renderer.

## [0.4.3] - 2026-10-21

### Added

- Added SSH-based CPU, memory, storage, network, temperature and identity metrics from remote Linux devices for free-canvas widgets.
- Added persistent media playback defaults for single-video completion, playlist/directory completion and slideshow timing.
- Added an actual mpv frame capture for the on-demand current-screen preview during video, slideshow and presentation playback.
- Added layered release gates, package verification, renderer fallback tests and physical-device acceptance procedures.

### Fixed

- Fell back to software DRM rendering for the rest of the boot when the GPU DRM renderer exits or cannot create its mpv IPC socket.
- Replaced the generic clock placeholder shown during media playback with the frame currently rendered on HDMI.

### Changed

- Reorganized the home page into guided, mutually exclusive Ready-made, Free canvas and Media workflows.
- Moved dashboard appearance editing beside the Ready-made workspace and reduced Settings to persistent device behavior.
- Reorganized Settings around startup behavior, display output, media playback, remote data sources and screen maintenance.
- Made direct DRM the default renderer for new installations while retaining Auto and Wayland compatibility options.

## [0.4.2] - 2026-09-21

### Added

- Added an exhibition slideshow action that loops images from the current media folder with a configurable interval.

### Changed

- Made single videos and media playlists loop continuously until the user stops or switches display mode.

### Fixed

- Generated installable `.raw` packages on macOS when `mksquashfs` is available instead of leaving a stale package in `dist`.

## [0.4.1] - 2026-09-07

### Fixed

- Verified mpv readiness with a real IPC response instead of only checking whether the Unix socket accepts connections.
- Restarted the HDMI renderer and retried safe display operations once after refused, reset, broken or vanished mpv socket connections.
- Reported the renderer as ready only after mpv answers IPC, avoiding false-ready status while no display is connected or mpv is restarting.

## [0.4.0] - 2026-09-05

### Added

- Added a live dashboard preview using the same scene data and coordinates as the HDMI renderer.
- Added adjustable dashboard font scaling from 80% to 200% and automatic, standard and large-type layouts.
- Added a free-canvas HDMI mode with draggable and resizable system, clock and text widgets.
- Added system overview, giant clock and minimal status canvas templates.
- Added solid, gradient and uploaded-image canvas backgrounds.
- Added `canvas` mode support to `zima-displayctl` and the DeepSeek Harness Skill.

### Changed

- Rebuilt dashboard rendering around a shared scene model so previews and HDMI output remain aligned.
- Added text clipping and explicit component bounds to prevent dashboard content overlap.

### Fixed

- Aligned both large-dashboard rows to the same three-column grid and separated CPU details from its usage bar.

## [0.3.2] - 2026-09-05

### Added

- Added an automatic large-type dashboard layout for 1024x600 and other displays at or below 1280x720.

### Fixed

- Limited displayed IP addresses to active physical Ethernet, Thunderbolt networking and Wi-Fi interfaces.
- Prioritized IPv4 and excluded Docker, bridge, VPN and other virtual network interfaces from the dashboard address.

## [0.3.1] - 2026-09-05

### Fixed

- Accepted the DSH integration endpoint with or without a trailing slash.
- Retried DSH detection automatically after a transient API or Gateway error.
- Updated upgrade instructions to restart an already-running Zima Display service explicitly.

## [0.3.0] - 2026-09-05

### Added

- Added Bearer-authenticated presentation and automation APIs.
- Added native HDMI document pages for Markdown and text-based presentation fallback.
- Added image presentation archives with safe ZIP extraction and page navigation.
- Added the cross-platform `zima-displayctl` automation client.
- Added automatic discovery and one-click persistent Skill installation for the ZimaOS Store DeepSeek Harness container.
- Added background installation of LibreOffice, Poppler and Noto CJK fonts inside the DSH container for full-fidelity PPTX/PDF conversion.

### Changed

- Extended player status and the Web UI with presentation title and page information.
- Packaged `zima-displayctl` alongside the native Zima Display service.

## [0.2.1] - 2026-09-05

### Fixed

- Synchronized manual language selection across the browser UI and HDMI dashboard.
- Updated settings and documentation to make the unified language behavior explicit.

## [0.2.0] - 2026-09-04

### Added

- Simplified Chinese and English browser UI with browser detection and a saved manual preference.
- Independent Simplified Chinese and English language setting for the HDMI dashboard and clock.
- GitHub-ready bilingual documentation, CI, contribution and security guides.

### Changed

- Documented direct DRM as the recommended Intel ZimaOS renderer.
- Removed the unused legacy mpv Lua renderer from the package.
- Removed the private development host from the deployment script defaults.

## [0.1.8] - 2026-09-04

### Fixed

- Replaced the unavailable mpv Lua integration with a Go-generated ASS overlay over JSON IPC.
- Kept generated HDMI modes active with a low-frame-rate black video source.
- Fixed renderer shutdown, stale socket detection and multi-event ASS rendering.
