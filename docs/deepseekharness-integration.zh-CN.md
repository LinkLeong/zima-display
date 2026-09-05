# DeepSeek Harness 投屏集成

Zima Display `v0.3.0` 开始支持 ZimaOS 应用商店中的 DeepSeek Harness Docker 版。Zima Display 本身仍使用 `.raw` 原生包安装，不使用 Docker。

## 最简单的安装方式

1. 在 ZimaOS 应用商店安装并启动 DeepSeekHarness。
2. 打开 Zima Display 控制页。
3. 在 DeepSeek Harness 区域点击“一键安装 Skill”。
4. 等待状态变为“Skill 已安装”；PPT/PDF 转换组件会继续在后台安装。
5. 刷新 DSH 页面。

商店容器把 `/root/.dsh` 挂载到 `/DATA/AppData/$AppID`，但 `$AppID` 在实际设备上可能是动态名称。Zima Display 使用 `docker inspect` 查找目标为 `/root/.dsh` 的可写挂载，因此不会硬编码目录。

安装内容：

```text
/root/.dsh/skills/zima-display/
├── SKILL.md
├── config.json
├── converter-install.log
└── scripts/
    └── zima-displayctl
```

`config.json` 权限为 `0600`，包含自动生成的 API Token。Skill 明确禁止在回复中展示 Token。

## 使用示例

```text
把 report.md 投到 HDMI 上
生成一份产品介绍 PPTX 并投屏
展示这个 PDF
下一页
上一页
停止投屏
显示系统仪表盘
```

也可以直接调用：

```bash
zima-displayctl --config config.json present report.md
zima-displayctl --config config.json present slides.pptx
zima-displayctl --config config.json present ./rendered-pages
zima-displayctl --config config.json play video.mp4
zima-displayctl --config config.json next
zima-displayctl --config config.json previous
zima-displayctl --config config.json stop
zima-displayctl --config config.json status
```

## 格式支持

| 格式 | 处理方式 | 外部依赖 |
| --- | --- | --- |
| Markdown / TXT | CLI 分页，Zima Display 使用原生 ASS 文档页输出 | 无 |
| PNG / JPG / WebP / BMP | 打包上传并按图片播放列表输出 | 无 |
| 图片目录 | 按文件名排序后打包展示 | 无 |
| PPTX | LibreOffice 转 PDF，再由 Poppler 转 PNG | LibreOffice、`pdftoppm` |
| PPTX 回退 | 转换器不可用时提取每张幻灯片文字并投屏 | 无 |
| PDF | Poppler 转 PNG | `pdftoppm` |
| 视频 | 上传到 Zima Display 后由 mpv 播放 | 无 |
| HTTP/HTTPS/RTSP | 直接交给 mpv 播放 | 无 |

首次点击安装按钮时，Zima Display 会在 DSH Debian 容器后台安装 `libreoffice-impress`、`poppler-utils` 和 `fonts-noto-cjk`。这些系统包属于容器层，DSH 商店升级后可能需要重新安装。Skill、CLI 和 Token 配置位于持久卷，不受容器升级影响；回到 Zima Display 点击“更新 Skill”即可自动检测和修复。

## 安全模型

- `/zima-display/api/v1/*` 必须携带 `Authorization: Bearer <token>`。
- Token 在首次启动时随机生成并持久保存。
- 展示包拒绝绝对路径、子目录和 `../` 路径穿越。
- 单页最大 100 MB，单个展示最多 500 页，整个上传最大 2 GB。
- DSH 持久目录必须位于 `/DATA/AppData`，并映射到容器的 `/root/.dsh`。

接口定义见 [OpenAPI 文档](openapi.yaml)。
