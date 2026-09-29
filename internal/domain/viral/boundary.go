package viral

import domaintranscript "go-api/internal/domain/transcript"

// ResolveBoundaries snaps LLM timestamps onto TranscriptWord edges (SOURCE TIME).
func ResolveBoundaries(words []domaintranscript.Word, startMs, endMs int64) (int64, int64) {
	if len(words) == 0 {
		return startMs, endMs
	}
	if endMs <= startMs {
		return startMs, endMs
	}

	resolvedStart := words[0].SourceStartMs
	for _, w := range words {
		if w.SourceEndMs < startMs {
			continue
		}
		// Prefer word that contains or starts at/after the LLM start.
		resolvedStart = w.SourceStartMs
		break
	}

	resolvedEnd := words[len(words)-1].SourceEndMs
	for i := len(words) - 1; i >= 0; i-- {
		w := words[i]
		if w.SourceStartMs > endMs {
			continue
		}
		resolvedEnd = w.SourceEndMs
		break
	}

	if resolvedEnd <= resolvedStart {
		// Fallback: nearest word around start.
		for _, w := range words {
			if w.SourceStartMs >= startMs {
				return w.SourceStartMs, w.SourceEndMs
			}
		}
		last := words[len(words)-1]
		return last.SourceStartMs, last.SourceEndMs
	}
	return resolvedStart, resolvedEnd
}
