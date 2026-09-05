---
name: zima-display
description: Control Zima Display HDMI output, present Markdown/PPTX/images, play media, change slides, and query display status.
---

# Zima Display

Use this skill when the user asks to show, present, project, play, or control content on the ZimaOS HDMI display.

The bundled CLI is at `scripts/zima-displayctl` and its generated connection configuration is at `config.json`. Resolve both to absolute paths from the directory containing this `SKILL.md` (provided by the skill resource metadata) before running a command. Do not print or reveal the token from `config.json`.

Commands:

```sh
scripts/zima-displayctl --config config.json status
scripts/zima-displayctl --config config.json present <file-or-image-directory>
scripts/zima-displayctl --config config.json play <video-or-network-url>
scripts/zima-displayctl --config config.json next
scripts/zima-displayctl --config config.json previous
scripts/zima-displayctl --config config.json stop
scripts/zima-displayctl --config config.json mode dashboard
scripts/zima-displayctl --config config.json mode clock
scripts/zima-displayctl --config config.json mode black
scripts/zima-displayctl --config config.json mode terminal
```

For a generated presentation, finish writing the requested file before invoking `present`. Markdown is rendered as readable native HDMI pages without external tools. Image files and image directories preserve their visuals. PPTX uses LibreOffice plus `pdftoppm` when available; otherwise it falls back to extracting readable slide text. PDF requires `pdftoppm`.

After a successful command, briefly tell the user what is on screen and include the current page when relevant. If conversion fails, report the missing dependency or unsupported format; do not claim the HDMI output changed.
