package timeline

import "fmt"

// ValidateTimeline checks segment invariants before persistence.
func ValidateTimeline(tl *Timeline, mediaDurationMs int64) error {
	if tl == nil {
		return fmt.Errorf("timeline is nil")
	}
	if len(tl.Segments) == 0 {
		return fmt.Errorf("timeline has no keep segments")
	}
	var prevOutputEnd int64 = -1
	for i, s := range tl.Segments {
		if s.SourceStartMs < 0 {
			return fmt.Errorf("segment %d source_start < 0", i)
		}
		if s.SourceEndMs <= s.SourceStartMs {
			return fmt.Errorf("segment %d invalid source range", i)
		}
		if mediaDurationMs > 0 && s.SourceEndMs > mediaDurationMs {
			return fmt.Errorf("segment %d source_end beyond media", i)
		}
		if s.OutputStartMs < 0 || s.OutputEndMs <= s.OutputStartMs {
			return fmt.Errorf("segment %d invalid output range", i)
		}
		if prevOutputEnd >= 0 && s.OutputStartMs != prevOutputEnd {
			return fmt.Errorf("segment %d output gap/overlap", i)
		}
		prevOutputEnd = s.OutputEndMs
		if s.Index != i {
			return fmt.Errorf("segment %d index mismatch", i)
		}
	}
	if tl.DurationMs != prevOutputEnd {
		return fmt.Errorf("duration_ms mismatch")
	}
	return nil
}
