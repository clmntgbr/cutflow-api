package fixture

import (
	"fmt"
	"strconv"
	"strings"

	"go-api/internal/domain/port"
)

type cue struct {
	StartMs int64
	EndMs   int64
	Text    string
}

func wordsFromSRT(srt string) ([]port.TranscriptWord, string) {
	cues := parseSRTCues(srt)
	words := make([]port.TranscriptWord, 0)
	textParts := make([]string, 0, len(cues))

	for _, c := range cues {
		textParts = append(textParts, c.Text)
		tokens := strings.Fields(c.Text)
		if len(tokens) == 0 {
			continue
		}
		duration := c.EndMs - c.StartMs
		if duration <= 0 {
			duration = int64(len(tokens)) * 100
		}
		step := duration / int64(len(tokens))
		if step <= 0 {
			step = 1
		}
		for i, token := range tokens {
			start := c.StartMs + int64(i)*step
			end := start + step
			if i == len(tokens)-1 {
				end = c.EndMs
			}
			if end <= start {
				end = start + 1
			}
			words = append(words, port.TranscriptWord{
				Text:    token,
				StartMs: start,
				EndMs:   end,
			})
		}
	}
	return words, strings.Join(textParts, " ")
}

func parseSRTCues(srt string) []cue {
	normalized := strings.ReplaceAll(srt, "\r\n", "\n")
	blocks := strings.Split(strings.TrimSpace(normalized), "\n\n")
	out := make([]cue, 0, len(blocks))

	for _, block := range blocks {
		lines := strings.Split(strings.TrimSpace(block), "\n")
		if len(lines) < 3 {
			continue
		}
		parts := strings.Split(lines[1], "-->")
		if len(parts) != 2 {
			continue
		}
		startMs, err1 := parseSRTTimestamp(strings.TrimSpace(parts[0]))
		endMs, err2 := parseSRTTimestamp(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil || endMs <= startMs {
			continue
		}
		text := strings.Join(lines[2:], " ")
		text = strings.Join(strings.Fields(text), " ")
		if text == "" {
			continue
		}
		out = append(out, cue{StartMs: startMs, EndMs: endMs, Text: text})
	}
	return out
}

func parseSRTTimestamp(raw string) (int64, error) {
	raw = strings.ReplaceAll(raw, ",", ".")
	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid srt timestamp %q", raw)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	secParts := strings.SplitN(parts[2], ".", 2)
	s, err := strconv.Atoi(secParts[0])
	if err != nil {
		return 0, err
	}
	ms := 0
	if len(secParts) == 2 {
		frac := secParts[1] + strings.Repeat("0", 3)
		ms, _ = strconv.Atoi(frac[:3])
	}
	return int64(h)*3_600_000 + int64(m)*60_000 + int64(s)*1000 + int64(ms), nil
}
