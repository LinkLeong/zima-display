package player

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"zima-display/internal/config"
	"zima-display/internal/metrics"
)

func CanvasScene(snapshot metrics.Snapshot, cfg config.Config, now time.Time) Scene {
	canvas := cfg.Canvas
	background := rgbToASS(canvas.BackgroundColor)
	b := newSceneBuilder(background, 1)
	if canvas.BackgroundType == "image" && canvas.BackgroundImage != "" {
		b.scene.Transparent = true
		b.scene.BackgroundImage = canvas.BackgroundImage
	}
	if canvas.BackgroundType == "gradient" {
		addGradient(b, canvas.BackgroundColor, canvas.SecondaryColor)
	}
	for _, widget := range canvas.Widgets {
		b.scene.Elements = append(b.scene.Elements, SceneElement{
			ID: widget.ID, WidgetID: widget.ID, Kind: "text",
			X: widget.X, Y: widget.Y, Width: widget.Width, Height: widget.Height,
			Color: rgbToASS(widget.Color), Opacity: 1, Text: canvasWidgetValue(widget, snapshot, cfg, now),
			FontSize: widget.FontSize, Bold: widget.Bold, Align: "left",
		})
	}
	return b.scene
}

func canvasWidgetValue(widget config.CanvasWidget, snapshot metrics.Snapshot, cfg config.Config, now time.Time) string {
	locale := dashboardLocale(cfg)
	labels := dashboardLabelsFor(locale)
	switch widget.Type {
	case "clock":
		return now.Format("15:04")
	case "date":
		return formatDashboardDate(now, locale)
	case "cpu":
		return fmt.Sprintf("CPU %.0f%%", snapshot.CPUPercent)
	case "memory":
		return fmt.Sprintf("%s %.0f%%", labels.memory, snapshot.MemoryPercent)
	case "storage":
		return fmt.Sprintf("%s %.0f%%", labels.storage, snapshot.DiskPercent)
	case "temperature":
		return fmt.Sprintf("%s %.0f°C", labels.thermal, snapshot.TemperatureC)
	case "ip":
		if len(snapshot.IPAddresses) > 0 {
			return snapshot.IPAddresses[0]
		}
		return labels.waitingNetwork
	case "hostname":
		if snapshot.Hostname != "" {
			return snapshot.Hostname
		}
		return "ZimaOS"
	case "network":
		return fmt.Sprintf("%s %s/s   %s %s/s", labels.download, humanBytesFloat(snapshot.NetworkRXBps), labels.upload, humanBytesFloat(snapshot.NetworkTXBps))
	case "text":
		return widget.Text
	default:
		return widget.Type
	}
}

func addGradient(b *sceneBuilder, first, second string) {
	start, okStart := parseRGB(first)
	end, okEnd := parseRGB(second)
	if !okStart || !okEnd {
		return
	}
	const bands = 32
	for index := 0; index < bands; index++ {
		ratio := float64(index) / float64(bands-1)
		red := int(float64(start[0]) + float64(end[0]-start[0])*ratio)
		green := int(float64(start[1]) + float64(end[1]-start[1])*ratio)
		blue := int(float64(start[2]) + float64(end[2]-start[2])*ratio)
		x := index * sceneWidth / bands
		next := (index + 1) * sceneWidth / bands
		b.rect(fmt.Sprintf("gradient-%02d", index), x, 0, next-x, sceneHeight, fmt.Sprintf("%02X%02X%02X", blue, green, red))
	}
}

func rgbToASS(value string) string {
	rgb, ok := parseRGB(value)
	if !ok {
		return "000000"
	}
	return fmt.Sprintf("%02X%02X%02X", rgb[2], rgb[1], rgb[0])
}

func parseRGB(value string) ([3]int, bool) {
	var result [3]int
	value = strings.TrimPrefix(value, "#")
	if len(value) != 6 {
		return result, false
	}
	for index := range result {
		parsed, err := strconv.ParseUint(value[index*2:index*2+2], 16, 8)
		if err != nil {
			return result, false
		}
		result[index] = int(parsed)
	}
	return result, true
}
