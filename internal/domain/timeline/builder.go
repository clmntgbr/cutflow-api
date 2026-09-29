package timeline

import (
	"fmt"

	"github.com/google/uuid"
)

// BuildTimeline builds KEEP segments + OUTPUT TIME mapping from REMOVE decisions.
func BuildTimeline(
	projectID, mediaFileID uuid.UUID,
	mediaDurationMs int64,
	decisions []Decision,
	speechMinDurationMs int,
) (*Timeline, error) {
	if mediaDurationMs <= 0 {
		return nil, fmt.Errorf("media duration must be > 0")
	}

	removeRanges := make([]Range, 0, len(decisions))
	for _, d := range decisions {
		if d.Action != ActionRemove {
			continue
		}
		if d.SourceEndMs <= d.SourceStartMs {
			continue
		}
		start := d.SourceStartMs
		end := d.SourceEndMs
		if start < 0 {
			start = 0
		}
		if end > mediaDurationMs {
			end = mediaDurationMs
		}
		if end <= start {
			continue
		}
		removeRanges = append(removeRanges, Range{
			StartMs: start,
			EndMs:   end,
			Reasons: append([]string(nil), d.Reasons...),
		})
	}
	removeRanges = NormalizeRanges(removeRanges)

	keeps := complementRanges(0, mediaDurationMs, removeRanges)
	keeps = filterShortSpeech(keeps, speechMinDurationMs, removeRanges)

	segments := make([]Segment, 0, len(keeps))
	var outputCursor int64
	for i, k := range keeps {
		dur := k.DurationMs()
		if dur <= 0 {
			continue
		}
		segments = append(segments, Segment{
			Index:         i,
			MediaFileID:   mediaFileID,
			SourceStartMs: k.StartMs,
			SourceEndMs:   k.EndMs,
			OutputStartMs: outputCursor,
			OutputEndMs:   outputCursor + dur,
		})
		outputCursor += dur
	}

	normalizedDecisions := make([]Decision, 0, len(removeRanges))
	for _, r := range removeRanges {
		typ := DecisionManual
		if len(r.Reasons) > 0 {
			typ = r.Reasons[0]
		}
		normalizedDecisions = append(normalizedDecisions, Decision{
			ID:            uuid.New(),
			MediaFileID:   mediaFileID,
			Type:          typ,
			SourceStartMs: r.StartMs,
			SourceEndMs:   r.EndMs,
			Action:        ActionRemove,
			Source:        SourceAutomatic,
			Reasons:       r.Reasons,
		})
	}

	tl := &Timeline{
		ID:            uuid.New(),
		ProjectID:     projectID,
		MediaFileID:   mediaFileID,
		DurationMs:    outputCursor,
		EngineVersion: EngineVersion,
		IsActive:      true,
		Segments:      segments,
		Decisions:     normalizedDecisions,
	}
	if err := ValidateTimeline(tl, mediaDurationMs); err != nil {
		return nil, err
	}
	return tl, nil
}

func complementRanges(start, end int64, removes []Range) []Range {
	keeps := make([]Range, 0, len(removes)+1)
	cursor := start
	for _, r := range removes {
		if r.StartMs > cursor {
			keeps = append(keeps, Range{StartMs: cursor, EndMs: r.StartMs})
		}
		if r.EndMs > cursor {
			cursor = r.EndMs
		}
	}
	if cursor < end {
		keeps = append(keeps, Range{StartMs: cursor, EndMs: end})
	}
	return keeps
}

// filterShortSpeech drops isolated KEEP fragments shorter than speechMin when
// surrounded by removes (or media edges treated as cut). Short keep at start/end
// of media is kept if it's the only remaining content.
func filterShortSpeech(keeps []Range, speechMin int, removes []Range) []Range {
	if speechMin <= 0 || len(keeps) == 0 {
		return keeps
	}
	_ = removes
	out := make([]Range, 0, len(keeps))
	for _, k := range keeps {
		if k.DurationMs() < int64(speechMin) && len(keeps) > 1 {
			continue
		}
		out = append(out, k)
	}
	if len(out) == 0 {
		return keeps // never delete everything via speech_min alone
	}
	return out
}
