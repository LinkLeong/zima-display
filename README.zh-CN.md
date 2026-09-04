# Zima Display

简体中文 | [English](README.md)

Zima Display 是一个面向 ZimaOS 的原生 HDMI 输出控制应用。它通过 ZimaOS 侧边栏提供控制界面，可以显示系统性能仪表盘、环境时钟、本地或网络媒体、黑屏待机，也可以随时释放 HDMI 并恢复系统终端。

应用使用 `zpkg` / `systemd-sysext` 的 `.raw` 包安装，不依赖 Docker。

## 功能

- ZimaOS 侧边栏原生控制界面，支持简体中文和英文。
- HDMI 系统仪表盘：CPU、内存、存储、网络、温度、GPU、IP、系统版本和显示器状态。
- Web 控制页和 HDMI 仪表盘可以分别设置语言。
- 系统仪表盘、环境时钟、黑屏待机、媒体播放和恢复终端模式。
- 本地视频和图片浏览、播放队列以及最大 20 GB 的媒体上传。
- HTTP、HTTPS、HLS 和 RTSP 网络媒体播放。
- 播放/暂停、进度、播放列表和 HDMI 音量控制。
- 配置持久化到 `/DATA/AppData/zima-display`。
- HDMI 渲染进程异常恢复和 ZimaOS 扩展启动 watchdog。

## 环境要求

- ZimaOS 和通过 HDMI 连接的显示器。
- 推荐使用 Intel 或 AMD 核显，不需要 NVIDIA 显卡。
- ZimaOS 自带的 `mpv`、`systemd`、`zpkg` 和 `systemd-sysext`。

Intel ZimaOS 设备上已验证并推荐直接使用 DRM 渲染；Wayland 保留为可选后端。

## 安装

从 GitHub Release 下载 `zima-display.raw`，复制到 ZimaOS 后执行：

```bash
sudo zpkg install --force ./zima-display.raw
sudo systemctl daemon-reload
sudo systemctl enable --now zima-display.service
```

安装完成后从 ZimaOS 侧边栏打开 **HDMI 显示**。也可以直接访问：

```text
http://<zimaos-host>/modules/zima-display/index.html
```

## 使用

1. 切换 HDMI 模式前先连接显示器。
2. 在控制页选择“系统仪表盘”“环境时钟”“黑屏待机”或“恢复终端”。
3. 打开“显示设置”，配置开机模式、DRM/Wayland 后端、HDMI 仪表盘语言、媒体目录和音频设备。
4. 在媒体面板浏览或上传文件，也可以直接输入网络媒体地址播放。

Web 控制页首次打开时跟随浏览器语言，手动选择后保存在当前浏览器。HDMI 仪表盘语言属于设备配置，保存在 `config.json` 中。

## 本地开发

需要 Go 1.20 或更新版本。

```bash
go test ./...
go run ./cmd/zima-displayd \
  -listen 127.0.0.1:8787 \
  -data-dir /tmp/zima-display-dev \
  -runtime-dir /tmp/zima-display-runtime \
  -web-dir web
```

浏览器打开 `http://127.0.0.1:8787/`。非 Linux/systemd/DRM 环境仍可开发 Web 界面和 API，但 HDMI 操作不可用。

## 构建与部署

构建 Linux amd64 安装包：

```bash
./scripts/build.sh amd64
```

构建脚本始终生成 `dist/zima-display-stage.tar.gz`。如果已经安装 `mksquashfs`，还会生成 `dist/zima-display.raw`。

部署到开发机：

```bash
./scripts/deploy.sh user@zimaos-host
```

也可以通过 `ZIMA_DISPLAY_TARGET` 指定目标。部署脚本不保存密码，也不会创建 Docker 容器。

## 架构

```text
ZimaOS 侧边栏
    |
    | CasaOS Gateway /zima-display
    v
zima-displayd
    |-- 指标采集 -> Go ASS 渲染 -> mpv 持久 overlay
    |-- 媒体浏览、上传与配置 API
    |-- mpv JSON IPC 播放和 HDMI 模式控制
    `-- systemd 在渲染器与 getty@tty1 之间切换

zima-display-renderer.service
    |-- mpv 直接 DRM（推荐）
    `-- Weston + mpv Wayland（可选）
```

HTTP 服务只监听本机回环地址，并通过 CasaOS Gateway 注册，因此控制 API 不会直接暴露到局域网。

## API

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/zima-display/api/health` | 健康状态和版本 |
| `GET` | `/zima-display/api/status` | 播放器、系统和显示状态 |
| `GET/PUT` | `/zima-display/api/config` | 获取或保存设置 |
| `POST` | `/zima-display/api/action` | 切换模式和控制播放 |
| `GET` | `/zima-display/api/media` | 浏览允许的媒体目录 |
| `POST` | `/zima-display/api/upload` | 上传视频或图片 |

媒体访问被限制在 `/DATA`、`/media`、`/mnt` 和应用数据目录内，API 不接受任意系统命令。

## 恢复终端

紧急释放 HDMI 并恢复终端：

```bash
sudo systemctl stop zima-display-renderer.service
sudo systemctl start getty@tty1.service
```

查看运行状态和日志：

```bash
sudo systemctl status zima-display.service zima-display-renderer.service
sudo journalctl -u zima-display.service -u zima-display-renderer.service -f
```

## 参与开发

开发与 Pull Request 说明见 [CONTRIBUTING.md](CONTRIBUTING.md)，安全问题请参考 [SECURITY.md](SECURITY.md)。

## 许可证

[MIT](LICENSE)
