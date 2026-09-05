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
	if mode == "clock" {
		return renderClock(now, dashboardLocale(cfg))
	}
	return renderDashboard(snapshot, cfg, now)
}

func renderDashboard(snapshot metrics.Snapshot, cfg config.Config, now time.Time) string {
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
	if compactDashboard(snapshot.Display) {
		return renderCompactDashboard(snapshot, compactDashboardContent{
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
		})
	}

	parts := []string{
		assRect(0, 0, 1920, 1080, "151411", "00"),
		assRect(0, 0, 26, 1080, "1866F2", "00"),
		assRect(80, 78, 1760, 2, "45423D", "00"),
		assText(82, 32, 26, "8C8982", title, true),
		assText(1500, 22, 54, "F7F3EA", now.Format("15:04"), true),
		assText(1500, 78, 18, "8C8982", formatDashboardDate(now, locale), false),
		assText(82, 148, 22, "8C8982", labels.systemLoad, true),
		assText(76, 188, 174, "F7F3EA", fmt.Sprintf("%.0f", snapshot.CPUPercent), true),
		assText(378, 294, 36, "1866F2", "%", true),
		assText(84, 390, 22, "8C8982", "CPU", true),
	}
	parts = append(parts, assBar(84, 434, 520, snapshot.CPUPercent, "1866F2")...)
	parts = append(parts,
		assText(84, 500, 19, "8C8982", labels.uptime, true),
		assText(84, 536, 38, "F7F3EA", humanDuration(snapshot.UptimeSeconds, locale), true),
		assText(84, 610, 19, "8C8982", labels.loadAverage, true),
		assText(84, 646, 38, "F7F3EA", fmt.Sprintf("%.2f", snapshot.Load1), true),
		assRect(690, 150, 530, 310, "24231F", "00"),
		assText(730, 184, 20, "8C8982", labels.memory, true),
		assText(730, 236, 74, "F7F3EA", fmt.Sprintf("%.0f%%", snapshot.MemoryPercent), true),
		assText(730, 330, 22, "B9B5AC", humanBytes(snapshot.MemoryUsedBytes)+" / "+humanBytes(snapshot.MemoryTotal), false),
	)
	parts = append(parts, assBar(730, 400, 450, snapshot.MemoryPercent, "63C6A5")...)
	parts = append(parts,
		assRect(1260, 150, 580, 310, "24231F", "00"),
		assText(1300, 184, 20, "8C8982", labels.storage, true),
		assText(1300, 236, 74, "F7F3EA", fmt.Sprintf("%.0f%%", snapshot.DiskPercent), true),
		assText(1300, 330, 22, "B9B5AC", humanBytes(snapshot.DiskUsedBytes)+" / "+humanBytes(snapshot.DiskTotalBytes), false),
	)
	parts = append(parts, assBar(1300, 400, 500, snapshot.DiskPercent, "E6A74C")...)
	parts = append(parts,
		assRect(690, 500, 360, 260, "24231F", "00"),
		assText(730, 534, 20, "8C8982", labels.network, true),
		assText(730, 590, 18, "8C8982", labels.download, true),
		assText(730, 622, 32, "F7F3EA", humanBytesFloat(snapshot.NetworkRXBps)+"/s", true),
		assText(730, 680, 18, "8C8982", labels.upload, true),
		assText(730, 712, 32, "F7F3EA", humanBytesFloat(snapshot.NetworkTXBps)+"/s", true),
		assRect(1090, 500, 350, 260, "24231F", "00"),
		assText(1130, 534, 20, "8C8982", labels.thermal, true),
		assText(1130, 594, 78, "F7F3EA", fmt.Sprintf("%.0f", snapshot.TemperatureC), true),
		assText(1260, 628, 30, "1866F2", "C", true),
		assText(1130, 700, 18, "8C8982", gpuVendor, true),
		assText(1130, 728, 17, "B9B5AC", gpuName, false),
		assRect(1480, 500, 360, 260, "24231F", "00"),
		assText(1520, 534, 20, "8C8982", labels.display, true),
		assText(1520, 594, 42, displayColor, displayStatus, true),
		assText(1520, 668, 18, "8C8982", snapshot.Display.Connector, true),
		assText(1520, 706, 17, "B9B5AC", displayMode, false),
		assRect(80, 850, 1760, 2, "45423D", "00"),
		assText(82, 890, 22, "F7F3EA", hostname, true),
		assText(82, 930, 18, "8C8982", osName, false),
		assText(650, 890, 18, "8C8982", labels.address, true),
		assText(650, 930, 22, "F7F3EA", address, false),
		assText(1370, 890, 18, "8C8982", labels.control, true),
		assText(1370, 930, 22, "1866F2", labels.openControl, true),
	)
	return strings.Join(parts, "\n")
}

type compactDashboardContent struct {
	labels                            dashboardLabels
	title, displayMode, displayStatus string
	displayColor, gpuVendor, gpuName  string
	address, hostname, osName, locale string
	now                               time.Time
}

func renderCompactDashboard(snapshot metrics.Snapshot, content compactDashboardContent) string {
	parts := []string{
		assRect(0, 0, 1920, 1080, "151411", "00"),
		assRect(0, 0, 32, 1080, "1866F2", "00"),
		assRect(72, 128, 1776, 3, "45423D", "00"),
		assText(74, 34, 36, "8C8982", content.title, true),
		assText(1540, 18, 74, "F7F3EA", content.now.Format("15:04"), true),
		assText(1540, 94, 26, "8C8982", formatDashboardDate(content.now, content.locale), false),

		assRect(72, 170, 560, 390, "24231F", "00"),
		assText(112, 206, 32, "8C8982", content.labels.systemLoad, true),
		assText(104, 250, 180, "F7F3EA", fmt.Sprintf("%.0f", snapshot.CPUPercent), true),
		assText(450, 386, 52, "1866F2", "%", true),
		assText(112, 452, 30, "8C8982", "CPU", true),
		assText(112, 512, 28, "B9B5AC", content.labels.uptime+" "+humanDuration(snapshot.UptimeSeconds, content.locale)+"  ·  "+content.labels.loadAverage+" "+fmt.Sprintf("%.2f", snapshot.Load1), false),

		assRect(668, 170, 560, 390, "24231F", "00"),
		assText(708, 206, 32, "8C8982", content.labels.memory, true),
		assText(708, 272, 112, "F7F3EA", fmt.Sprintf("%.0f%%", snapshot.MemoryPercent), true),
		assText(708, 420, 30, "B9B5AC", humanBytes(snapshot.MemoryUsedBytes)+" / "+humanBytes(snapshot.MemoryTotal), false),

		assRect(1264, 170, 584, 390, "24231F", "00"),
		assText(1304, 206, 32, "8C8982", content.labels.storage, true),
		assText(1304, 272, 112, "F7F3EA", fmt.Sprintf("%.0f%%", snapshot.DiskPercent), true),
		assText(1304, 420, 30, "B9B5AC", humanBytes(snapshot.DiskUsedBytes)+" / "+humanBytes(snapshot.DiskTotalBytes), false),
	}
	parts = append(parts, assBar(112, 476, 480, snapshot.CPUPercent, "1866F2")...)
	parts = append(parts, assBar(708, 492, 480, snapshot.MemoryPercent, "63C6A5")...)
	parts = append(parts, assBar(1304, 492, 504, snapshot.DiskPercent, "E6A74C")...)
	parts = append(parts,
		assRect(72, 600, 690, 300, "24231F", "00"),
		assText(112, 636, 30, "8C8982", content.labels.network, true),
		assText(112, 700, 26, "8C8982", content.labels.download, true),
		assText(112, 746, 46, "F7F3EA", humanBytesFloat(snapshot.NetworkRXBps)+"/s", true),
		assText(430, 700, 26, "8C8982", content.labels.upload, true),
		assText(430, 746, 46, "F7F3EA", humanBytesFloat(snapshot.NetworkTXBps)+"/s", true),

		assRect(798, 600, 480, 300, "24231F", "00"),
		assText(838, 636, 30, "8C8982", content.labels.thermal, true),
		assText(838, 690, 115, "F7F3EA", fmt.Sprintf("%.0f", snapshot.TemperatureC), true),
		assText(1035, 770, 42, "1866F2", "C", true),
		assText(838, 838, 25, "B9B5AC", content.gpuVendor+" · "+content.gpuName, false),

		assRect(1314, 600, 534, 300, "24231F", "00"),
		assText(1354, 636, 30, "8C8982", content.labels.display, true),
		assText(1354, 704, 58, content.displayColor, content.displayStatus, true),
		assText(1354, 806, 28, "B9B5AC", content.displayMode+" · "+snapshot.Display.Connector, false),

		assRect(72, 950, 1776, 3, "45423D", "00"),
		assText(74, 976, 30, "F7F3EA", content.hostname, true),
		assText(74, 1020, 22, "8C8982", content.osName, false),
		assText(730, 976, 24, "8C8982", content.labels.address, true),
		assText(730, 1016, 32, "F7F3EA", content.address, true),
	)
	return strings.Join(parts, "\n")
}

func compactDashboard(display metrics.Display) bool {
	if len(display.Modes) == 0 {
		return false
	}
	var width, height int
	if _, err := fmt.Sscanf(display.Modes[0], "%dx%d", &width, &height); err != nil {
		return false
	}
	return width <= 1280 || height <= 720
}

func renderClock(now time.Time, locale string) string {
	return strings.Join([]string{
		assRect(0, 0, 1920, 1080, "151411", "00"),
		assRect(0, 0, 26, 1080, "1866F2", "00"),
		assText(130, 280, 250, "F7F3EA", now.Format("15:04"), true),
		assText(145, 610, 44, "8C8982", formatClockDate(now, locale), false),
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
