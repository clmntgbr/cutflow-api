package transcriptissue

import (
	"strings"
	"unicode"

	domaintranscript "go-api/internal/domain/transcript"
)

const maxPhraseLen = 6

// DetectOptions controls which detectors run (from media_configuration).
type DetectOptions struct {
	DetectFillers      bool
	DetectRepetitions  bool
	DetectFalseStarts  bool
	Language           string
}

// Detect finds fillers, consecutive repetitions, and false starts on SourceTime words.
// Proposed removals always target the first occurrence (keep the last / fluent one).
func Detect(words []domaintranscript.Word, opts DetectOptions) []*Issue {
	if len(words) == 0 {
		return nil
	}

	tokens := make([]string, len(words))
	for i, w := range words {
		tokens[i] = normalizeToken(w.Text)
	}

	marked := make([]string, len(words)) // issue type per word index, empty = unmarked

	if opts.DetectFalseStarts {
		markFalseStarts(tokens, marked)
	}
	if opts.DetectRepetitions {
		markRepetitions(tokens, marked)
	}
	if opts.DetectFillers {
		markFillers(tokens, marked, opts.Language)
	}

	return issuesFromMarks(words, marked)
}

func markFalseStarts(tokens []string, marked []string) {
	n := len(tokens)
	for size := maxPhraseLen; size >= 2; size-- {
		i := 0
		for i+2*size <= n {
			if !rangeFree(marked, i, i+2*size-1) {
				i++
				continue
			}
			if !equalRange(tokens, i, i+size, i+size, i+2*size) {
				i++
				continue
			}
			if anyEmpty(tokens, i, i+size-1) {
				i++
				continue
			}
			for k := i; k < i+size; k++ {
				marked[k] = TypeFalseStart
			}
			i += size // keep the second occurrence
		}
	}
}

func markRepetitions(tokens []string, marked []string) {
	n := len(tokens)
	i := 0
	for i < n-1 {
		if tokens[i] == "" || tokens[i] != tokens[i+1] {
			i++
			continue
		}
		j := i + 1
		for j < n && tokens[j] == tokens[i] {
			j++
		}
		// Keep last occurrence (j-1); mark earlier ones as repetition if free.
		for k := i; k < j-1; k++ {
			if marked[k] == "" {
				marked[k] = TypeRepetition
			}
		}
		i = j
	}
}

func markFillers(tokens []string, marked []string, language string) {
	single, phrases := fillerDict(language)
	n := len(tokens)

	for size := 3; size >= 2; size-- {
		for i := 0; i+size <= n; i++ {
			if !rangeFree(marked, i, i+size-1) {
				continue
			}
			key := strings.Join(tokens[i:i+size], " ")
			if !phrases[key] {
				continue
			}
			for k := i; k < i+size; k++ {
				marked[k] = TypeFiller
			}
		}
	}

	for i := 0; i < n; i++ {
		if marked[i] != "" || tokens[i] == "" {
			continue
		}
		if single[tokens[i]] {
			marked[i] = TypeFiller
		}
	}
}

func issuesFromMarks(words []domaintranscript.Word, marked []string) []*Issue {
	var out []*Issue
	i := 0
	for i < len(words) {
		if marked[i] == "" {
			i++
			continue
		}
		typ := marked[i]
		j := i + 1
		for j < len(words) && marked[j] == typ {
			j++
		}
		parts := make([]string, 0, j-i)
		var confSum float64
		var confCount int
		for k := i; k < j; k++ {
			parts = append(parts, words[k].Text)
			if words[k].Confidence != nil {
				confSum += *words[k].Confidence
				confCount++
			}
		}
		var confidence *float64
		if confCount > 0 {
			avg := confSum / float64(confCount)
			confidence = &avg
		}
		out = append(out, &Issue{
			Type:           typ,
			Text:           strings.Join(parts, " "),
			SourceStartMs:  words[i].SourceStartMs,
			SourceEndMs:    words[j-1].SourceEndMs,
			Confidence:     confidence,
			WordStartIndex: words[i].WordIndex,
			WordEndIndex:   words[j-1].WordIndex,
		})
		i = j
	}
	return out
}

// ApplyKinds sets Word.Kind from detected issues (non-speech kinds win).
func ApplyKinds(words []domaintranscript.Word, issues []*Issue) []domaintranscript.Word {
	out := make([]domaintranscript.Word, len(words))
	copy(out, words)
	for i := range out {
		if out[i].Kind == "" {
			out[i].Kind = domaintranscript.KindSpeech
		}
	}
	indexByWord := make(map[int]int, len(out))
	for i, w := range out {
		indexByWord[w.WordIndex] = i
	}
	for _, issue := range issues {
		kind := domaintranscript.KindSpeech
		switch issue.Type {
		case TypeFiller:
			kind = domaintranscript.KindFiller
		case TypeRepetition:
			kind = domaintranscript.KindRepetition
		case TypeFalseStart:
			kind = domaintranscript.KindFalseStart
		}
		for wi := issue.WordStartIndex; wi <= issue.WordEndIndex; wi++ {
			if idx, ok := indexByWord[wi]; ok {
				out[idx].Kind = kind
			}
		}
	}
	return out
}

func normalizeToken(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range strings.ToLower(strings.TrimSpace(text)) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '\'' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func rangeFree(marked []string, start, endInclusive int) bool {
	for i := start; i <= endInclusive && i < len(marked); i++ {
		if marked[i] != "" {
			return false
		}
	}
	return true
}

func equalRange(tokens []string, a0, a1, b0, b1 int) bool {
	if a1-a0 != b1-b0 {
		return false
	}
	for i := 0; i < a1-a0; i++ {
		if tokens[a0+i] != tokens[b0+i] {
			return false
		}
	}
	return true
}

func anyEmpty(tokens []string, start, endInclusive int) bool {
	for i := start; i <= endInclusive; i++ {
		if tokens[i] == "" {
			return true
		}
	}
	return false
}
