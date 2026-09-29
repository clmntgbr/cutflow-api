package subtitle

import (
	"fmt"
	"strings"
)

// GenerateASS builds a word-highlight ASS script from timed words and a style.
// Timestamps must already be in the desired timeline (source or remapped output).
func GenerateASS(words []Word, style Style) string {
	style = style.normalized()
	groups := groupWords(words, style)

	var b strings.Builder
	b.WriteString("[Script Info]\n")
	b.WriteString("ScriptType: v4.00+\n")
	b.WriteString(fmt.Sprintf("PlayResX: %d\n", style.PlayResX))
	b.WriteString(fmt.Sprintf("PlayResY: %d\n", style.PlayResY))
	b.WriteString("WrapStyle: 0\n")
	b.WriteString("ScaledBorderAndShadow: yes\n\n")

	b.WriteString("[V4+ Styles]\n")
	b.WriteString("Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n")
	bold := 0
	if style.Bold {
		bold = -1
	}
	fmt.Fprintf(
		&b,
		"Style: Default,%s,%d,%s,%s,%s,&H00000000,%d,0,0,0,100,100,0,0,1,%.1f,%.1f,%d,%d,%d,%d,1\n\n",
		style.FontFamily,
		style.FontSize,
		cssColorToASS(style.TextColor),
		cssColorToASS(style.ActiveColor),
		cssColorToASS(style.OutlineColor),
		bold,
		style.OutlineSize,
		style.Shadow,
		style.Alignment,
		style.MarginL,
		style.MarginR,
		style.MarginV,
	)

	b.WriteString("[Events]\n")
	b.WriteString("Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n")

	activeColor := cssColorToASS(style.ActiveColor)
	textColor := cssColorToASS(style.TextColor)
	scalePct := int(style.ActiveScale * 100)
	if scalePct <= 0 {
		scalePct = 100
	}

	for _, g := range groups {
		for i := range g.Words {
			start := g.Words[i].StartMs
			end := g.Words[i].EndMs
			if i+1 < len(g.Words) && g.Words[i+1].StartMs > start {
				end = g.Words[i+1].StartMs
			}
			if end <= start {
				continue
			}
			text := renderGroupLine(g.Words, i, activeColor, textColor, scalePct)
			fmt.Fprintf(
				&b,
				"Dialogue: 0,%s,%s,Default,,0,0,0,,%s\n",
				msToASS(start),
				msToASS(end),
				text,
			)
		}
	}
	return b.String()
}

func renderGroupLine(words []Word, activeIdx int, activeColor, textColor string, scalePct int) string {
	parts := make([]string, 0, len(words))
	for i, w := range words {
		escaped := escapeASS(w.Text)
		if i == activeIdx {
			if scalePct != 100 {
				parts = append(parts, fmt.Sprintf(
					`{\c%s&\fscx%d\fscy%d}%s{\c%s&\fscx100\fscy100}`,
					activeColor, scalePct, scalePct, escaped, textColor,
				))
			} else {
				parts = append(parts, fmt.Sprintf(
					`{\c%s&}%s{\c%s&}`,
					activeColor, escaped, textColor,
				))
			}
			continue
		}
		parts = append(parts, escaped)
	}
	return strings.Join(parts, " ")
}

func escapeASS(text string) string {
	text = strings.ReplaceAll(text, `\`, `\\`)
	text = strings.ReplaceAll(text, `{`, `\{`)
	text = strings.ReplaceAll(text, `}`, `\}`)
	return text
}

// cssColorToASS converts #RRGGBB into ASS &H00BBGGRR.
func cssColorToASS(css string) string {
	css = strings.TrimPrefix(strings.TrimSpace(css), "#")
	if len(css) != 6 {
		return "&H00FFFFFF"
	}
	r := css[0:2]
	g := css[2:4]
	b := css[4:6]
	return "&H00" + strings.ToUpper(b+g+r)
}

func msToASS(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	centis := ms / 10
	h := centis / 360000
	centis %= 360000
	m := centis / 6000
	centis %= 6000
	s := centis / 100
	cs := centis % 100
	return fmt.Sprintf("%d:%02d:%02d.%02d", h, m, s, cs)
}
