package subtitle

import (
	"strings"
)

// Word is a timed caption token in output (or source) time.
type Word struct {
	Text    string
	StartMs int64
	EndMs   int64
}

type group struct {
	StartMs int64
	EndMs   int64
	Words   []Word
}

func groupWords(words []Word, style Style) []group {
	cleaned := make([]Word, 0, len(words))
	for _, w := range words {
		text := strings.TrimSpace(w.Text)
		if text == "" || w.EndMs <= w.StartMs {
			continue
		}
		cleaned = append(cleaned, Word{Text: text, StartMs: w.StartMs, EndMs: w.EndMs})
	}
	if len(cleaned) == 0 {
		return nil
	}

	var groups []group
	current := group{
		StartMs: cleaned[0].StartMs,
		EndMs:   cleaned[0].EndMs,
		Words:   []Word{cleaned[0]},
	}
	chars := len([]rune(cleaned[0].Text))

	for i := 1; i < len(cleaned); i++ {
		w := cleaned[i]
		gap := w.StartMs - current.EndMs
		nextChars := chars + 1 + len([]rune(w.Text))
		shouldSplit := len(current.Words) >= style.MaxWords ||
			gap > style.MaxGapMs ||
			endsSentence(current.Words[len(current.Words)-1].Text) ||
			(style.MaxChars > 0 && nextChars > style.MaxChars)

		if shouldSplit {
			groups = append(groups, current)
			current = group{StartMs: w.StartMs, EndMs: w.EndMs, Words: []Word{w}}
			chars = len([]rune(w.Text))
			continue
		}
		current.Words = append(current.Words, w)
		current.EndMs = w.EndMs
		chars = nextChars
	}
	groups = append(groups, current)
	return groups
}

func endsSentence(text string) bool {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) == 0 {
		return false
	}
	switch runes[len(runes)-1] {
	case '.', '!', '?', '…':
		return true
	default:
		return false
	}
}
