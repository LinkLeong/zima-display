# Zima Display

[简体中文](README.zh-CN.md) | English

Zima Display is a native HDMI output controller for ZimaOS. It adds a control panel to the ZimaOS sidebar and can show a live system dashboard, an ambient clock, local or network media, a black standby screen, or the original system terminal.

It is installed as a `zpkg` / `systemd-sysext` `.raw` package. Docker is not required.

## Features

- Native ZimaOS sidebar UI with Simplified Chinese and English support.
- HDMI dashboard with CPU, memory, storage, network, temperature, GPU, IP, OS and connector status.
- One language selector synchronizes the browser UI and HDMI dashboard.
- Dashboard, clock, black standby, media playback and terminal modes.
- Local video and image browser, playback queue and uploads up to 20 GB.
- HTTP, HTTPS, HLS and RTSP network playback.
- Playback, seek, playlist and HDMI volume controls.
- Persistent configuration under `/DATA/AppData/zima-display`.
- Renderer recovery and a boot watchdog for ZimaOS system extensions.
- One-click DeepSeek Harness integration with automatic Store-container volume discovery.
- Markdown, PPTX, PDF and image-deck presentation with page navigation.

## Requirements

- ZimaOS with an HDMI-connected display.
- Intel or AMD integrated graphics are recommended. NVIDIA hardware is not required.
- `mpv`, `systemd`, `zpkg` and `systemd-sysext`, as provided by ZimaOS.

The tested and recommended renderer on Intel ZimaOS devices is direct DRM. Wayland remains available as an optional backend.

## Install

Download `zima-display.raw` from a GitHub Release, copy it to the ZimaOS host, then run:

```bash
sudo zpkg install --force ./zima-display.raw
sudo systemctl daemon-reload
sudo systemctl enable --now zima-display.service
```

Open **Zima Display** from the ZimaOS sidebar after installation. The page is also available at:

```text
http://<zimaos-host>/modules/zima-display/index.html
```

## Usage

1. Connect the display before switching HDMI modes.
2. Select **System dashboard**, **Ambient clock**, **Black standby**, or **Restore terminal** in the control panel.
3. Open **Display settings** to configure the startup mode, DRM/Wayland backend, display language, media folders and audio device.
4. Use the media panel to browse files, upload media, or play a network URL.

The web UI follows the browser language on first use. A manual language selection synchronizes the web UI and HDMI dashboard, and is stored both in the browser and the device `config.json`.

## DeepSeek Harness

1. Install and start **DeepSeekHarness** from the ZimaOS App Store.
2. Open Zima Display and click **Install Skill** in the DeepSeek Harness panel.
3. Refresh DSH, then ask it to present generated Markdown, PPTX, PDF or images.

The Skill and connection configuration are stored persistently under `/root/.dsh/skills/zima-display`. Zima Display discovers the actual dynamic `/DATA/AppData` bind mount automatically. LibreOffice, Poppler and Noto CJK fonts are installed in the DSH container in the background for PPTX/PDF rendering. Click **Update Skill** after a DSH container upgrade if those converters need repair.

## Development

Go 1.20 or newer is required.

```bash
go test ./...
go run ./cmd/zima-displayd \
  -listen 127.0.0.1:8787 \
  -data-dir /tmp/zima-display-dev \
  -runtime-dir /tmp/zima-display-runtime \
  -web-dir web
```

Open `http://127.0.0.1:8787/`. HDMI operations are unavailable on systems without Linux, systemd and DRM, but the web UI and API can still be developed locally.

## Build and deploy

Build a Linux amd64 package:

```bash
./scripts/build.sh amd64
```

The script always creates `dist/zima-display-stage.tar.gz`. It also creates `dist/zima-display.raw` when `mksquashfs` is installed.

Deploy to a development machine:

```bash
./scripts/deploy.sh user@zimaos-host
```

You can also set `ZIMA_DISPLAY_TARGET`. The deployment script does not store credentials and does not create Docker containers.

## Architecture

```text
ZimaOS sidebar
    |
    | /zima-display via CasaOS Gateway
    v
zima-displayd
    |-- metrics collection -> Go ASS renderer -> persistent mpv overlay
    |-- media browser, uploads and configuration API
    |-- mpv JSON IPC for playback and HDMI mode control
    `-- systemd switching between the renderer and getty@tty1

zima-display-renderer.service
    |-- direct DRM mpv (recommended)
    `-- Weston + Wayland mpv (optional)
```

The HTTP service listens on loopback and registers through CasaOS Gateway, so its control API is not directly exposed to the LAN.

## Automation

DeepSeek Harness and other automation clients can use the packaged `zima-displayctl`. See the [DeepSeekHarness integration guide](docs/deepseekharness-integration.zh-CN.md) and [OpenAPI document](docs/openapi.yaml).

## API

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/zima-display/api/health` | Health and version |
| `GET` | `/zima-display/api/status` | Player, system and display status |
| `GET/PUT` | `/zima-display/api/config` | Read or update settings |
| `POST` | `/zima-display/api/action` | Change modes and control playback |
| `GET` | `/zima-display/api/media` | Browse allowed media folders |
| `POST` | `/zima-display/api/upload` | Upload a video or image |
| `POST` | `/zima-display/api/v1/presentations` | Upload an image presentation archive (Bearer Token) |
| `POST` | `/zima-display/api/v1/presentations/text` | Create a text presentation (Bearer Token) |
| `POST` | `/zima-display/api/v1/presentations/{id}/activate` | Activate a presentation (Bearer Token) |
| `POST` | `/zima-display/api/v1/presentation/next` | Show the next page (Bearer Token) |
| `POST` | `/zima-display/api/v1/presentation/previous` | Show the previous page (Bearer Token) |
| `POST` | `/zima-display/api/v1/presentation/stop` | Stop presenting (Bearer Token) |

Media access is restricted to `/DATA`, `/media`, `/mnt` and the application data directory. The API does not accept arbitrary system commands.

## Recovery

To release HDMI and restore the terminal:

```bash
sudo systemctl stop zima-display-renderer.service
sudo systemctl start getty@tty1.service
```

For diagnostics:

```bash
sudo systemctl status zima-display.service zima-display-renderer.service
sudo journalctl -u zima-display.service -u zima-display-renderer.service -f
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development and pull request guidance. Security issues should follow [SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE)
