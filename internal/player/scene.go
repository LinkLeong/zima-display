package player

import (
	"fmt"
	"math"
	"strings"
)

const (
	sceneWidth  = 1920
	sceneHeight = 1080
)

type Scene struct {
	Width           int            `json:"width"`
	Height          int            `json:"height"`
	Background      string         `json:"background"`
	BackgroundImage string         `json:"background_image,omitempty"`
	Transparent     bool           `json:"transparent,omitempty"`
	Elements        []SceneElement `json:"elements"`
}

type SceneElement struct {
	ID       string  `json:"id"`
	WidgetID string  `json:"widget_id,omitempty"`
	Kind     string  `json:"kind"`
	X        int     `json:"x"`
	Y        int     `json:"y"`
	Width    int     `json:"width,omitempty"`
	Height   int     `json:"height,omitempty"`
	Color    string  `json:"color"`
	Opacity  float64 `json:"opacity,omitempty"`
	Text     string  `json:"text,omitempty"`
	FontSize int     `json:"font_size,omitempty"`
	Bold     bool    `json:"bold,omitempty"`
	Align    string  `json:"align,omitempty"`
}

type sceneBuilder struct {
	scene     Scene
	fontScale float64
}

func newSceneBuilder(background string, fontScale float64) *sceneBuilder {
	return &sceneBuilder{
		scene:     Scene{Width: sceneWidth, Height: sceneHeight, Background: background},
		fontScale: clamp(fontScale, 0.8, 2),
	}
}

func (b *sceneBuilder) rect(id string, x, y, width, height int, color string) {
	b.scene.Elements = append(b.scene.Elements, SceneElement{
		ID: id, Kind: "rect", X: x, Y: y, Width: width, Height: height, Color: color, Opacity: 1,
	})
}

func (b *sceneBuilder) text(id string, x, y, width, height, size int, color, value string, bold bool) {
	b.scene.Elements = append(b.scene.Elements, SceneElement{
		ID: id, Kind: "text", X: x, Y: y, Width: width, Height: height, Color: color, Opacity: 1,
		Text: value, FontSize: scaledDashboardFont(size, b.fontScale), Bold: bold, Align: "left",
	})
}

func (b *sceneBuilder) bar(id string, x, y, width int, percent float64, color string) {
	b.rect(id+"-track", x, y, width, 8, "3A3936")
	b.rect(id+"-value", x, y, int(math.Floor(float64(width)*clamp(percent, 0, 100)/100)), 8, color)
}

func scaledDashboardFont(size int, scale float64) int {
	effective := scale
	if size >= 100 {
		effective = 1 + (scale-1)*0.35
	} else if size >= 60 {
		effective = 1 + (scale-1)*0.6
	}
	scaled := int(math.Round(float64(size) * effective))
	if scaled < 1 {
		return 1
	}
	return scaled
}

func RenderSceneASS(scene Scene) string {
	var parts []string
	if !scene.Transparent {
		parts = append(parts, assRect(0, 0, scene.Width, scene.Height, scene.Background, "00"))
	}
	for _, element := range scene.Elements {
		switch element.Kind {
		case "rect":
			parts = append(parts, assRect(element.X, element.Y, element.Width, element.Height, element.Color, assAlpha(element.Opacity)))
		case "text":
			parts = append(parts, assSceneText(element))
		}
	}
	return strings.Join(parts, "\n")
}

func assSceneText(element SceneElement) string {
	weight := 0
	if element.Bold {
		weight = 1
	}
	clip := ""
	if element.Width > 0 && element.Height > 0 {
		clip = fmt.Sprintf("\\clip(%d,%d,%d,%d)", element.X, element.Y, element.X+element.Width, element.Y+element.Height)
	}
	return fmt.Sprintf("{\\an7\\pos(%d,%d)\\fnNoto Sans CJK SC\\fs%d\\b%d\\bord0\\shad0\\1c&H%s&%s}%s",
		element.X, element.Y, element.FontSize, weight, element.Color, clip, escapeASS(element.Text))
}

func assAlpha(opacity float64) string {
	if opacity <= 0 {
		return "FF"
	}
	if opacity >= 1 {
		return "00"
	}
	return fmt.Sprintf("%02X", int(math.Round((1-opacity)*255)))
}
