package viral

import (
	"fmt"
	"strings"
	"time"

	domaintranscript "go-api/internal/domain/transcript"
)

// FormatTimestamped builds a compact LLM-friendly transcript from words.
func FormatTimestamped(words []domaintranscript.Word, cueMs int64) string {
	if cueMs <= 0 {
		cueMs = 5000
	}
	cues := groupCues(words, cueMs)
	var b strings.Builder
	for _, cue := range cues {
		fmt.Fprintf(&b, "[%s - %s]\n%s\n\n", formatClock(cue.StartMs), formatClock(cue.EndMs), cue.Text)
	}
	return strings.TrimSpace(b.String())
}

type cue struct {
	StartMs int64
	EndMs   int64
	Text    string
}

func groupCues(words []domaintranscript.Word, cueMs int64) []cue {
	if len(words) == 0 {
		return nil
	}
	var out []cue
	start := words[0].SourceStartMs
	end := words[0].SourceEndMs
	var parts []string
	flush := func() {
		if len(parts) == 0 {
			return
		}
		out = append(out, cue{StartMs: start, EndMs: end, Text: strings.Join(parts, " ")})
		parts = nil
	}
	for _, w := range words {
		if len(parts) > 0 && (w.SourceStartMs-start >= cueMs || w.SourceStartMs-end > 1200) {
			flush()
			start = w.SourceStartMs
		}
		if len(parts) == 0 {
			start = w.SourceStartMs
		}
		parts = append(parts, w.Text)
		end = w.SourceEndMs
	}
	flush()
	return out
}

func formatClock(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	d := time.Duration(ms) * time.Millisecond
	m := int(d.Minutes())
	s := d.Seconds() - float64(m*60)
	return fmt.Sprintf("%02d:%06.3f", m, s)
}
