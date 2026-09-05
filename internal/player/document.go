package player

import (
	"fmt"
	"strings"
)

func renderDocument(title, content string, page, pageCount int) string {
	lines := documentLines(content, 18)
	parts := []string{
		assRect(0, 0, 1920, 1080, "151411", "00"),
		assRect(0, 0, 26, 1080, "1866F2", "00"),
		assText(92, 55, 24, "8C8982", strings.ToUpper(title), true),
		assText(1690, 55, 22, "8C8982", fmt.Sprintf("%02d / %02d", page+1, pageCount), true),
		assRect(92, 108, 1736, 2, "45423D", "00"),
	}
	y := 160
	for index, line := range lines {
		size := 34
		color := "E6E1D7"
		bold := false
		if index == 0 && len(lines) > 1 {
			size = 54
			color = "F7F3EA"
			bold = true
		}
		parts = append(parts, assText(105, y, size, color, line, bold))
		y += size + 17
	}
	parts = append(parts,
		assRect(92, 1015, 1736, 2, "45423D", "00"),
		assText(92, 1030, 17, "1866F2", "ZIMA DISPLAY / DEEPSEEK HARNESS", true),
	)
	return strings.Join(parts, "\n")
}

func documentLines(content string, limit int) []string {
	var result []string
	for _, raw := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		line = strings.TrimSpace(strings.TrimLeft(line, "#>*-0123456789. "))
		if line == "" {
			continue
		}
		for len([]rune(line)) > 54 {
			runes := []rune(line)
			result = append(result, string(runes[:54]))
			line = string(runes[54:])
			if len(result) >= limit {
				return result
			}
		}
		result = append(result, line)
		if len(result) >= limit {
			return result
		}
	}
	if len(result) == 0 {
		return []string{"Empty page"}
	}
	return result
}
