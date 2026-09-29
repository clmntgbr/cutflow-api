package subtitle

import (
	"fmt"
	"strconv"
	"strings"
)

// SRTToASS converts a plain SRT document into a minimal ASS script.
func SRTToASS(srt string) string {
	var b strings.Builder
	b.WriteString("[Script Info]\n")
	b.WriteString("ScriptType: v4.00+\n")
	b.WriteString("PlayResX: 1920\n")
	b.WriteString("PlayResY: 1080\n")
	b.WriteString("WrapStyle: 0\n\n")
	b.WriteString("[V4+ Styles]\n")
	b.WriteString("Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n")
	b.WriteString("Style: Default,Arial,48,&H00FFFFFF,&H000000FF,&H00000000,&H64000000,0,0,0,0,100,100,0,0,1,2,0,2,40,40,40,1\n\n")
	b.WriteString("[Events]\n")
	b.WriteString("Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n")

	blocks := strings.Split(strings.ReplaceAll(strings.TrimSpace(srt), "\r\n", "\n"), "\n\n")
	for _, block := range blocks {
		lines := strings.Split(strings.TrimSpace(block), "\n")
		if len(lines) < 3 {
			continue
		}
		parts := strings.Split(lines[1], "-->")
		if len(parts) != 2 {
			continue
		}
		start := srtTimestampToASS(strings.TrimSpace(parts[0]))
		end := srtTimestampToASS(strings.TrimSpace(parts[1]))
		text := strings.Join(lines[2:], "\\N")
		text = strings.ReplaceAll(text, "\n", "\\N")
		fmt.Fprintf(&b, "Dialogue: 0,%s,%s,Default,,0,0,0,,%s\n", start, end, text)
	}
	return b.String()
}

func srtTimestampToASS(raw string) string {
	raw = strings.ReplaceAll(raw, ",", ".")
	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		return "0:00:00.00"
	}
	secParts := strings.SplitN(parts[2], ".", 2)
	sec := secParts[0]
	cs := "00"
	if len(secParts) == 2 {
		frac := secParts[1]
		if len(frac) >= 2 {
			cs = frac[:2]
		} else {
			cs = frac + strings.Repeat("0", 2-len(frac))
		}
	}
	h, _ := strconv.Atoi(parts[0])
	return fmt.Sprintf("%d:%s:%s.%s", h, parts[1], sec, cs)
}
