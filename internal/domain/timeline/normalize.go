package timeline

import "sort"

// Range is a SOURCE TIME interval.
type Range struct {
	StartMs int64
	EndMs   int64
	Reasons []string
}

func (r Range) DurationMs() int64 {
	if r.EndMs <= r.StartMs {
		return 0
	}
	return r.EndMs - r.StartMs
}

// NormalizeRanges merges overlapping/adjacent REMOVE ranges and unions reasons.
func NormalizeRanges(ranges []Range) []Range {
	if len(ranges) == 0 {
		return nil
	}
	sorted := append([]Range(nil), ranges...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].StartMs == sorted[j].StartMs {
			return sorted[i].EndMs < sorted[j].EndMs
		}
		return sorted[i].StartMs < sorted[j].StartMs
	})

	out := []Range{sorted[0]}
	for i := 1; i < len(sorted); i++ {
		cur := sorted[i]
		last := &out[len(out)-1]
		if cur.StartMs <= last.EndMs {
			if cur.EndMs > last.EndMs {
				last.EndMs = cur.EndMs
			}
			last.Reasons = mergeReasons(last.Reasons, cur.Reasons)
			continue
		}
		out = append(out, cur)
	}
	return out
}

func mergeReasons(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, r := range append(a, b...) {
		if r == "" {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	return out
}
