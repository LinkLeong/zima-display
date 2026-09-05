package player

import (
	"fmt"
	"math"
	"strings"
	"time"

	"zima-display/internal/config"
	"zima-display/internal/metrics"
)

func renderMode(mode string, snapshot metrics.Snapshot, cfg config.Config, now time.Time) string {
	if mode == "canvas" {
		return RenderSceneASS(CanvasScene(snapshot, cfg, now))
	}
	if mode == "clock" {
		return renderClock(now, dashboardLocale(cfg), cfg.Dashboard.FontScale)
	}
	return renderDashboard(snapshot, cfg, now)
}

func renderDashboard(snapshot metrics.Snapshot, cfg config.Config, now time.Time) string {
	return RenderSceneASS(DashboardScene(snapshot, cfg, now))
}

func DashboardScene(snapshot metrics.Snapshot, cfg config.Config, now time.Time) Scene {
	locale := dashboardLocale(cfg)
	labels := dashboardLabelsFor(locale)
	title := strings.ToUpper(cfg.Dashboard.Title)
	if title == "" {
		title = "ZIMA DISPLAY"
	}
	displayMode := labels.preferredMode
	if len(snapshot.Display.Modes) > 0 {
		displayMode = snapshot.Display.Modes[0]
	}
	displayStatus := labels.noSignal
	displayColor := "5C87E8"
	if snapshot.Display.Connected {
		displayStatus = labels.online
		displayColor = "63C6A5"
	}
	gpuVendor := snapshot.GPU.Vendor
	if gpuVendor == "" {
		gpuVendor = "GRAPHICS"
	}
	gpuName := snapshot.GPU.Name
	if gpuName == "" {
		gpuName = labels.automatic
	}
	address := labels.waitingNetwork
	if len(snapshot.IPAddresses) > 0 {
		address = snapshot.IPAddresses[0]
	}
	hostname := snapshot.Hostname
	if hostname == "" {
		hostname = "ZimaOS"
	}
	osName := snapshot.OS
	if osName == "" {
		osName = "ZimaOS"
	}
	content := compactDashboardContent{
		labels:        labels,
		title:         title,
		displayMode:   displayMode,
		displayStatus: displayStatus,
		displayColor:  displayColor,
		gpuVendor:     gpuVendor,
		gpuName:       gpuName,
		address:       address,
		hostname:      hostname,
		osName:        osName,
		locale:        locale,
		now:           now,
	}
	if largeDashboard(cfg, snapshot.Display) {
		return renderLargeDashboardScene(snapshot, content, cfg.Dashboard.FontScale)
	}
	return renderStandardDashboardScene(snapshot, content, cfg.Dashboard.FontScale)
}

func renderStandardDashboardScene(snapshot metrics.Snapshot, content compactDashboardContent, fontScale float64) Scene {
	b := newSceneBuilder("151411", fontScale)
	b.rect("accent", 0, 0, 26, 1080, "1866F2")
	b.rect("header-line", 80, 78, 1760, 2, "45423D")
	b.text("title", 82, 32, 1120, 50, 26, "8C8982", content.title, true)
	b.text("clock", 1500, 22, 340, 70, 54, "F7F3EA", content.now.Format("15:04"), true)
	b.text("date", 1500, 78, 340, 36, 18, "8C8982", formatDashboardDate(content.now, content.locale), false)
	b.text("cpu-label", 82, 148, 520, 42, 22, "8C8982", content.labels.systemLoad, true)
	b.text("cpu-value", 76, 188, 380, 205, 174, "F7F3EA", fmt.Sprintf("%.0f", snapshot.CPUPercent), true)
	b.text("cpu-unit", 378, 294, 120, 60, 36, "1866F2", "%", true)
	b.text("cpu-name", 84, 390, 520, 40, 22, "8C8982", "CPU", true)
	b.bar("cpu-bar", 84, 434, 520, snapshot.CPUPercent, "1866F2")
	b.text("uptime-label", 84, 500, 520, 34, 19, "8C8982", content.labels.uptime, true)
	b.text("uptime-value", 84, 536, 520, 58, 38, "F7F3EA", humanDuration(snapshot.UptimeSeconds, content.locale), true)
	b.text("load-label", 84, 610, 520, 34, 19, "8C8982", content.labels.loadAverage, true)
	b.text("load-value", 84, 646, 520, 58, 38, "F7F3EA", fmt.Sprintf("%.2f", snapshot.Load1), true)

	b.rect("memory-card", 690, 150, 530, 310, "24231F")
	b.text("memory-label", 730, 184, 450, 42, 20, "8C8982", content.labels.memory, true)
	b.text("memory-value", 730, 236, 450, 88, 74, "F7F3EA", fmt.Sprintf("%.0f%%", snapshot.MemoryPercent), true)
	b.text("memory-detail", 730, 330, 450, 44, 22, "B9B5AC", humanBytes(snapshot.MemoryUsedBytes)+" / "+humanBytes(snapshot.MemoryTotal), false)
	b.bar("memory-bar", 730, 400, 450, snapshot.MemoryPercent, "63C6A5")

	b.rect("storage-card", 1260, 150, 580, 310, "24231F")
	b.text("storage-label", 1300, 184, 500, 42, 20, "8C8982", content.labels.storage, true)
	b.text("storage-value", 1300, 236, 500, 88, 74, "F7F3EA", fmt.Sprintf("%.0f%%", snapshot.DiskPercent), true)
	b.text("storage-detail", 1300, 330, 500, 44, 22, "B9B5AC", humanBytes(snapshot.DiskUsedBytes)+" / "+humanBytes(snapshot.DiskTotalBytes), false)
	b.bar("storage-bar", 1300, 400, 500, snapshot.DiskPercent, "E6A74C")

	b.rect("network-card", 690, 500, 360, 260, "24231F")
	b.text("network-label", 730, 534, 280, 38, 20, "8C8982", content.labels.network, true)
	b.text("download-label", 730, 590, 280, 30, 18, "8C8982", content.labels.download, true)
	b.text("download-value", 730, 622, 280, 46, 32, "F7F3EA", humanBytesFloat(snapshot.NetworkRXBps)+"/s", true)
	b.text("upload-label", 730, 680, 280, 30, 18, "8C8982", content.labels.upload, true)
	b.text("upload-value", 730, 712, 280, 46, 32, "F7F3EA", humanBytesFloat(snapshot.NetworkTXBps)+"/s", true)

	b.rect("thermal-card", 1090, 500, 350, 260, "24231F")
	b.text("thermal-label", 1130, 534, 270, 38, 20, "8C8982", content.labels.thermal, true)
	b.text("thermal-value", 1130, 594, 150, 94, 78, "F7F3EA", fmt.Sprintf("%.0f", snapshot.TemperatureC), true)
	b.text("thermal-unit", 1260, 628, 90, 54, 30, "1866F2", "C", true)
	b.text("gpu-vendor", 1130, 700, 270, 28, 18, "8C8982", content.gpuVendor, true)
	b.text("gpu-name", 1130, 728, 270, 28, 17, "B9B5AC", content.gpuName, false)

	b.rect("display-card", 1480, 500, 360, 260, "24231F")
	b.text("display-label", 1520, 534, 280, 38, 20, "8C8982", content.labels.display, true)
	b.text("display-status", 1520, 594, 280, 58, 42, content.displayColor, content.displayStatus, true)
	b.text("display-connector", 1520, 668, 280, 30, 18, "8C8982", snapshot.Display.Connector, true)
	b.text("display-mode", 1520, 706, 280, 30, 17, "B9B5AC", content.displayMode, false)

	b.rect("footer-line", 80, 850, 1760, 2, "45423D")
	b.text("hostname", 82, 890, 500, 38, 22, "F7F3EA", content.hostname, true)
	b.text("os", 82, 930, 500, 34, 18, "8C8982", content.osName, false)
	b.text("address-label", 650, 890, 650, 34, 18, "8C8982", content.labels.address, true)
	b.text("address", 650, 930, 650, 40, 22, "F7F3EA", content.address, false)
	b.text("control-label", 1370, 890, 470, 34, 18, "8C8982", content.labels.control, true)
	b.text("control", 1370, 930, 470, 40, 22, "1866F2", content.labels.openControl, true)
	return b.scene
}

type compactDashboardContent struct {
	labels                            dashboardLabels
	title, displayMode, displayStatus string
	displayColor, gpuVendor, gpuName  string
	address, hostname, osName, locale string
	now                               time.Time
}

func renderLargeDashboardScene(snapshot metrics.Snapshot, content compactDashboardContent, fontScale float64) Scene {
	b := newSceneBuilder("151411", fontScale)
	b.rect("accent", 0, 0, 32, 1080, "1866F2")
	b.rect("header-line", 72, 128, 1776, 3, "45423D")
	b.text("title", 74, 34, 1180, 62, 36, "8C8982", content.title, true)
	b.text("clock", 1540, 18, 308, 82, 74, "F7F3EA", content.now.Format("15:04"), true)
	b.text("date", 1540, 94, 308, 38, 26, "8C8982", formatDashboardDate(content.now, content.locale), false)

	// Both rows share the same three-column grid so the compact layout reads as one composition.
	const cardWidth = 576
	cardX := [...]int{72, 672, 1272}
	contentX := [...]int{112, 712, 1312}

	b.rect("cpu-card", cardX[0], 170, cardWidth, 380, "24231F")
	b.text("cpu-label", contentX[0], 206, 496, 46, 30, "8C8982", content.labels.systemLoad, true)
	b.text("cpu-value", contentX[0], 270, 496, 132, 112, "F7F3EA", fmt.Sprintf("%.0f%%", snapshot.CPUPercent), true)
	b.text("cpu-uptime-label", contentX[0], 414, 220, 26, 18, "8C8982", content.labels.uptime, true)
	b.text("cpu-uptime-value", contentX[0], 446, 220, 36, 24, "B9B5AC", humanDuration(snapshot.UptimeSeconds, content.locale), true)
	b.text("cpu-load-label", contentX[0]+258, 414, 220, 26, 18, "8C8982", content.labels.loadAverage, true)
	b.text("cpu-load-value", contentX[0]+258, 446, 220, 36, 24, "B9B5AC", fmt.Sprintf("%.2f", snapshot.Load1), true)
	b.bar("cpu-bar", contentX[0], 510, 496, snapshot.CPUPercent, "1866F2")

	b.rect("memory-card", cardX[1], 170, cardWidth, 380, "24231F")
	b.text("memory-label", contentX[1], 206, 496, 46, 30, "8C8982", content.labels.memory, true)
	b.text("memory-value", contentX[1], 270, 496, 132, 112, "F7F3EA", fmt.Sprintf("%.0f%%", snapshot.MemoryPercent), true)
	b.text("memory-detail", contentX[1], 420, 496, 40, 24, "B9B5AC", humanBytes(snapshot.MemoryUsedBytes)+" / "+humanBytes(snapshot.MemoryTotal), false)
	b.bar("memory-bar", contentX[1], 500, 496, snapshot.MemoryPercent, "63C6A5")

	b.rect("storage-card", cardX[2], 170, cardWidth, 380, "24231F")
	b.text("storage-label", contentX[2], 206, 496, 46, 30, "8C8982", content.labels.storage, true)
	b.text("storage-value", contentX[2], 270, 496, 132, 112, "F7F3EA", fmt.Sprintf("%.0f%%", snapshot.DiskPercent), true)
	b.text("storage-detail", contentX[2], 420, 496, 40, 24, "B9B5AC", humanBytes(snapshot.DiskUsedBytes)+" / "+humanBytes(snapshot.DiskTotalBytes), false)
	b.bar("storage-bar", contentX[2], 500, 496, snapshot.DiskPercent, "E6A74C")

	b.rect("network-card", cardX[0], 584, cardWidth, 320, "24231F")
	b.text("network-label", contentX[0], 620, 496, 44, 30, "8C8982", content.labels.network, true)
	b.text("download-label", contentX[0], 690, 220, 38, 24, "8C8982", content.labels.download, true)
	b.text("download-value", contentX[0], 736, 220, 62, 44, "F7F3EA", humanBytesFloat(snapshot.NetworkRXBps)+"/s", true)
	b.text("upload-label", contentX[0]+258, 690, 220, 38, 24, "8C8982", content.labels.upload, true)
	b.text("upload-value", contentX[0]+258, 736, 220, 62, 44, "F7F3EA", humanBytesFloat(snapshot.NetworkTXBps)+"/s", true)

	b.rect("thermal-card", cardX[1], 584, cardWidth, 320, "24231F")
	b.text("thermal-label", contentX[1], 620, 496, 44, 30, "8C8982", content.labels.thermal, true)
	b.text("thermal-value", contentX[1], 684, 200, 138, 112, "F7F3EA", fmt.Sprintf("%.0f", snapshot.TemperatureC), true)
	b.text("thermal-unit", contentX[1]+220, 760, 90, 56, 40, "1866F2", "C", true)
	b.text("gpu", contentX[1], 838, 496, 38, 24, "B9B5AC", content.gpuVendor+" · "+content.gpuName, false)

	b.rect("display-card", cardX[2], 584, cardWidth, 320, "24231F")
	b.text("display-label", contentX[2], 620, 496, 44, 30, "8C8982", content.labels.display, true)
	b.text("display-status", contentX[2], 692, 496, 76, 58, content.displayColor, content.displayStatus, true)
	b.text("display-connector", contentX[2], 792, 496, 38, 26, "B9B5AC", snapshot.Display.Connector, true)
	b.text("display-mode", contentX[2], 838, 496, 38, 24, "B9B5AC", content.displayMode, false)

	b.rect("footer-line", 72, 950, 1776, 3, "45423D")
	b.text("hostname", 74, 976, 570, 42, 30, "F7F3EA", content.hostname, true)
	b.text("os", 74, 1020, 570, 34, 22, "8C8982", content.osName, false)
	b.text("address-label", 730, 976, 1118, 36, 24, "8C8982", content.labels.address, true)
	b.text("address", 730, 1016, 1118, 46, 32, "F7F3EA", content.address, true)
	return b.scene
}

func largeDashboard(cfg config.Config, display metrics.Display) bool {
	switch cfg.Dashboard.Layout {
	case "large":
		return true
	case "standard":
		return false
	}
	if cfg.Dashboard.FontScale >= 1.25 {
		return true
	}
	if len(display.Modes) == 0 {
		return false
	}
	var width, height int
	if _, err := fmt.Sscanf(display.Modes[0], "%dx%d", &width, &height); err != nil {
		return false
	}
	return width <= 1280 || height <= 720
}

func renderClock(now time.Time, locale string, fontScale float64) string {
	return strings.Join([]string{
		assRect(0, 0, 1920, 1080, "151411", "00"),
		assRect(0, 0, 26, 1080, "1866F2", "00"),
		assText(130, 280, scaledDashboardFont(250, fontScale), "F7F3EA", now.Format("15:04"), true),
		assText(145, 610, scaledDashboardFont(44, fontScale), "8C8982", formatClockDate(now, locale), false),
		assRect(145, 704, 780, 10, "1866F2", "00"),
	}, "\n")
}

func assText(x, y, size int, color, value string, bold bool) string {
	weight := 0
	if bold {
		weight = 1
	}
	return fmt.Sprintf("{\\an7\\pos(%d,%d)\\fnNoto Sans CJK SC\\fs%d\\b%d\\bord0\\shad0\\1c&H%s&}%s", x, y, size, weight, color, escapeASS(value))
}

func assRect(x, y, width, height int, color, alpha string) string {
	return fmt.Sprintf("{\\an7\\pos(0,0)\\bord0\\shad0\\1c&H%s&\\1a&H%s&\\p1}m %d %d l %d %d l %d %d l %d %d{\\p0}", color, alpha, x, y, x+width, y, x+width, y+height, x, y+height)
}

func assBar(x, y, width int, percent float64, color string) []string {
	filled := int(math.Floor(float64(width) * clamp(percent, 0, 100) / 100))
	return []string{
		assRect(x, y, width, 8, "3A3936", "00"),
		assRect(x, y, filled, 8, color, "00"),
	}
}

func escapeASS(value string) string {
	return strings.NewReplacer("\\", "\\\\", "{", "\\{", "}", "\\}", "\n", " ").Replace(value)
}

func clamp(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func humanBytes(value uint64) string {
	return humanBytesFloat(float64(value))
}

func humanBytesFloat(value float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	index := 0
	for value >= 1024 && index < len(units)-1 {
		value /= 1024
		index++
	}
	if index <= 1 {
		return fmt.Sprintf("%.0f %s", value, units[index])
	}
	return fmt.Sprintf("%.1f %s", value, units[index])
}

func humanDuration(seconds float64, locale string) string {
	total := int(math.Floor(seconds))
	days := total / 86400
	hours := total % 86400 / 3600
	minutes := total % 3600 / 60
	if days > 0 {
		if locale == "zh-CN" {
			return fmt.Sprintf("%d天 %02d时", days, hours)
		}
		return fmt.Sprintf("%dd %02dh", days, hours)
	}
	if locale == "zh-CN" {
		return fmt.Sprintf("%02d时 %02d分", hours, minutes)
	}
	return fmt.Sprintf("%02dh %02dm", hours, minutes)
}

type dashboardLabels struct {
	systemLoad, uptime, loadAverage, memory, storage string
	network, download, upload, thermal, display      string
	online, noSignal, address, control               string
	openControl, waitingNetwork, preferredMode       string
	automatic                                        string
}

func dashboardLocale(cfg config.Config) string {
	if cfg.Dashboard.Language == "en-US" {
		return "en-US"
	}
	return "zh-CN"
}

func dashboardLabelsFor(locale string) dashboardLabels {
	if locale == "en-US" {
		return dashboardLabels{
			systemLoad: "SYSTEM LOAD", uptime: "UPTIME", loadAverage: "LOAD AVERAGE",
			memory: "MEMORY", storage: "STORAGE", network: "NETWORK", download: "DOWN",
			upload: "UP", thermal: "THERMAL", display: "DISPLAY", online: "ONLINE",
			noSignal: "NO SIGNAL", address: "ADDRESS", control: "CONTROL",
			openControl: "Open ZimaOS > Zima Display", waitingNetwork: "Waiting for network",
			preferredMode: "Preferred mode", automatic: "Automatic",
		}
	}
	return dashboardLabels{
		systemLoad: "系统负载", uptime: "运行时间", loadAverage: "平均负载",
		memory: "内存", storage: "存储", network: "网络", download: "下行",
		upload: "上行", thermal: "温度", display: "显示器", online: "已连接",
		noSignal: "无信号", address: "网络地址", control: "控制入口",
		openControl: "打开 ZimaOS > HDMI 显示", waitingNetwork: "等待网络连接",
		preferredMode: "首选分辨率", automatic: "自动检测",
	}
}

func formatDashboardDate(now time.Time, locale string) string {
	if locale == "zh-CN" {
		return now.Format("2006.01.02") + "  " + chineseWeekday(now.Weekday())
	}
	return now.Format("2006.01.02  Monday")
}

func formatClockDate(now time.Time, locale string) string {
	if locale == "zh-CN" {
		return chineseWeekday(now.Weekday()) + "  /  " + now.Format("2006.01.02")
	}
	return now.Format("Monday  /  2006.01.02")
}

func chineseWeekday(day time.Weekday) string {
	return [...]string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}[day]
}
