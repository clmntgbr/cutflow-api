package viral

import domaintranscript "go-api/internal/domain/transcript"

// Chunk is a SourceTime window of words for long-video LLM analysis.
type Chunk struct {
	StartMs int64
	EndMs   int64
	Words   []domaintranscript.Word
}

// ChunkWords splits words into overlapping windows when the media is long.
// durationMs/overlapMs are in milliseconds; if total span <= durationMs, one chunk is returned.
func ChunkWords(words []domaintranscript.Word, durationMs, overlapMs int64) []Chunk {
	if len(words) == 0 {
		return nil
	}
	if durationMs <= 0 {
		durationMs = 10 * 60 * 1000
	}
	if overlapMs < 0 {
		overlapMs = 0
	}
	if overlapMs >= durationMs {
		overlapMs = durationMs / 10
	}

	first := words[0].SourceStartMs
	last := words[len(words)-1].SourceEndMs
	if last-first <= durationMs {
		return []Chunk{{StartMs: first, EndMs: last, Words: words}}
	}

	var out []Chunk
	windowStart := first
	for windowStart < last {
		windowEnd := windowStart + durationMs
		chunkWords := wordsInRange(words, windowStart, windowEnd)
		if len(chunkWords) > 0 {
			out = append(out, Chunk{
				StartMs: chunkWords[0].SourceStartMs,
				EndMs:   chunkWords[len(chunkWords)-1].SourceEndMs,
				Words:   chunkWords,
			})
		}
		next := windowStart + durationMs - overlapMs
		if next <= windowStart {
			next = windowStart + durationMs
		}
		windowStart = next
		if windowStart >= last {
			break
		}
	}
	return out
}

func wordsInRange(words []domaintranscript.Word, startMs, endMs int64) []domaintranscript.Word {
	out := make([]domaintranscript.Word, 0)
	for _, w := range words {
		if w.SourceEndMs < startMs {
			continue
		}
		if w.SourceStartMs > endMs {
			break
		}
		out = append(out, w)
	}
	return out
}
