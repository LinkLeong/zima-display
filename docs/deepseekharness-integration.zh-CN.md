# DeepSeekHarness 投屏集成方案

本文档描述计划中的自动化投屏能力，当前 `v0.2.x` 尚未提供这些接口。

## 目标场景

DeepSeekHarness 完成内容生成后，可以直接执行以下动作：

- 生成 PPTX 后投屏，并控制上一页、下一页和结束播放。
- 生成 Markdown 后按适合电视阅读的页面投屏。
- 投屏 PDF、图片序列和现有媒体文件。
- 查询 HDMI 是否连接、当前内容和播放状态。

## 推荐架构

```text
DeepSeekHarness
    |
    | 调用 zima-displayctl / Harness Skill
    v
控制端转换器
    |-- PPTX -> PDF -> PNG 页面
    |-- Markdown -> HTML -> PNG 页面
    `-- PDF -> PNG 页面
    |
    | HTTPS/HTTP + Bearer Token 上传展示包
    v
Zima Display API
    |-- 保存展示包和 manifest.json
    |-- 激活展示、上一页、下一页、自动翻页
    `-- mpv DRM 输出 PNG 页面序列
```

转换应在运行 DeepSeekHarness 的控制端完成，而不是在 ZimaOS 上完成。当前 ZimaOS 测试机只有 Python 3 和 mpv，没有 LibreOffice、Chromium、Pandoc 或 PDF 转图工具。在 ZimaOS 内加入完整办公和浏览器运行时会明显增加包体、依赖和 DRM 冲突风险。

## Zima Display 调整

### 展示包

增加独立目录：

```text
/DATA/AppData/zima-display/presentations/<presentation-id>/
├── manifest.json
├── 001.png
├── 002.png
└── ...
```

manifest 至少包含标题、页面顺序、当前页、适配方式、自动翻页秒数和循环设置。

### 自动化 API

建议增加：

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `POST` | `/zima-display/api/v1/presentations` | 上传 ZIP 展示包 |
| `POST` | `/zima-display/api/v1/presentations/{id}/activate` | 开始投屏 |
| `DELETE` | `/zima-display/api/v1/presentations/{id}` | 删除展示包 |
| `POST` | `/zima-display/api/v1/presentation/next` | 下一页 |
| `POST` | `/zima-display/api/v1/presentation/previous` | 上一页 |
| `POST` | `/zima-display/api/v1/presentation/stop` | 结束并返回仪表盘 |
| `GET` | `/zima-display/api/v1/presentation/status` | 查询当前展示状态 |

自动化接口需要 Bearer Token、上传大小限制、ZIP 路径穿越检查和幂等请求 ID。现有浏览器控制接口继续通过 CasaOS Gateway 使用。

### 播放能力

- 新增 `presentation` 模式和页面索引状态。
- 使用 mpv 图片播放列表输出静态页面。
- 支持 contain、cover、背景色、手动翻页、定时翻页和循环。
- Web 控制页显示当前文档、页码和自动化调用来源，并允许人工接管。

## DeepSeekHarness 侧调整

提供跨平台 `zima-displayctl`，DeepSeekHarness 不需要理解转换和 API 细节：

```bash
zima-displayctl present report.pptx
zima-displayctl present design.md
zima-displayctl next
zima-displayctl previous
zima-displayctl stop
zima-displayctl status
```

CLI 自动完成格式识别、转换、打包、上传和激活。通过环境变量配置：

```text
ZIMA_DISPLAY_URL=http://zimaos-host/zima-display
ZIMA_DISPLAY_TOKEN=<generated-token>
```

同时提供 DeepSeekHarness Skill/工具定义：

- `present_file`
- `next_slide`
- `previous_slide`
- `stop_presentation`
- `get_display_status`

## 交付物

1. `zima-display.raw`：包含展示包管理、页面播放、认证 API 和 Web 管理界面。
2. `zima-displayctl`：提供 macOS/Linux 的 amd64 与 arm64 二进制。
3. DeepSeekHarness Skill：包含工具定义、调用规则和示例提示词。
4. OpenAPI 文档：描述上传、激活、翻页、停止和状态接口。
5. 示例项目：一份 PPTX、一份 Markdown 以及端到端投屏脚本。
6. 自动化测试：ZIP 安全、Token 鉴权、页面顺序、人工接管和异常恢复。
7. GitHub Release：服务端 `.raw`、控制端 CLI、校验值和升级说明。

## MVP 边界

- PPTX 首版按静态页面投屏，不保留动画、视频、音频和复杂切换效果。
- Markdown 首版分页展示，不提供网页式连续滚动。
- 内容转换依赖控制端可安装 LibreOffice、Chromium/Playwright 和 PDF 转图工具。
- 如果必须完整保留 PPT 动画，需要另行实现浏览器或 Office 渲染后端，不建议放入首版。
