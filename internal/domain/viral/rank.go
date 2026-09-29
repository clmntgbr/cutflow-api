package viral

import "sort"

// Deduplicate merges overlapping candidates (IoU >= threshold), keeping the higher score.
func Deduplicate(candidates []*Candidate, iouThreshold float64) []*Candidate {
	if iouThreshold <= 0 {
		iouThreshold = 0.5
	}
	if len(candidates) == 0 {
		return nil
	}
	sorted := append([]*Candidate(nil), candidates...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	kept := make([]*Candidate, 0, len(sorted))
	for _, c := range sorted {
		overlap := false
		for _, k := range kept {
			if iou(c, k) >= iouThreshold {
				overlap = true
				break
			}
		}
		if !overlap {
			kept = append(kept, c)
		}
	}
	return kept
}

func iou(a, b *Candidate) float64 {
	start := max64(a.SourceStartMs, b.SourceStartMs)
	end := min64(a.SourceEndMs, b.SourceEndMs)
	if end <= start {
		return 0
	}
	inter := float64(end - start)
	union := float64(a.DurationMs() + b.DurationMs()) - inter
	if union <= 0 {
		return 0
	}
	return inter / union
}

// Rank sorts by score descending and truncates to maxN.
func Rank(candidates []*Candidate, maxN int) []*Candidate {
	if maxN <= 0 {
		maxN = 10
	}
	sorted := Deduplicate(candidates, 0.5)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})
	if len(sorted) > maxN {
		sorted = sorted[:maxN]
	}
	return sorted
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
