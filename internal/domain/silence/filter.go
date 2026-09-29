package silence

// AnalysisMinSilenceMs is the ffmpeg silencedetect floor when capturing raw
// DetectedSilence rows. Config SilenceMinDurationMs / paddings are applied later
// via ApplyEditFilters so they can change without re-running analysis.
const AnalysisMinSilenceMs int64 = 50

// ApplyEditFilters turns raw DetectedSilence intervals into cuttable windows:
//  1. keep paddingAfterMs after the preceding speech (shrink start)
//  2. keep paddingBeforeMs before the following speech (shrink end)
//  3. drop intervals whose cuttable duration is shorter than minDurationMs
//
// Naming matches media_configuration: padding before/after speech edges
// (architecture pre_roll / post_roll). speech_min_duration_ms is applied
// separately when building speech keep-segments, not here.
func ApplyEditFilters(raw []Interval, minDurationMs, paddingBeforeMs, paddingAfterMs int) []Interval {
	if minDurationMs < 0 {
		minDurationMs = 0
	}
	if paddingBeforeMs < 0 {
		paddingBeforeMs = 0
	}
	if paddingAfterMs < 0 {
		paddingAfterMs = 0
	}

	out := make([]Interval, 0, len(raw))
	for _, in := range raw {
		start := in.StartMs + int64(paddingAfterMs)
		end := in.EndMs - int64(paddingBeforeMs)
		if end <= start {
			continue
		}
		cut := Interval{StartMs: start, EndMs: end}
		if cut.DurationMs() < int64(minDurationMs) {
			continue
		}
		out = append(out, cut)
	}
	return out
}

// IntervalsFromDetected maps persisted rows to SourceTime intervals.
func IntervalsFromDetected(rows []*DetectedSilence) []Interval {
	out := make([]Interval, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, row.Interval())
	}
	return out
}
