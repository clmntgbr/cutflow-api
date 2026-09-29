package timeline

// Mapper converts SOURCE TIME ↔ OUTPUT TIME for a built Timeline.
type Mapper struct {
	segments []Segment
}

func NewMapper(tl *Timeline) *Mapper {
	if tl == nil {
		return &Mapper{}
	}
	return &Mapper{segments: append([]Segment(nil), tl.Segments...)}
}

// MapSourceToOutput returns output time and whether the source point is kept.
func (m *Mapper) MapSourceToOutput(sourceMs int64) (outputMs int64, kept bool) {
	for _, s := range m.segments {
		if sourceMs < s.SourceStartMs || sourceMs > s.SourceEndMs {
			continue
		}
		offset := sourceMs - s.SourceStartMs
		return s.OutputStartMs + offset, true
	}
	return 0, false
}

// MapRange maps a SOURCE range that may cross cuts into contiguous OUTPUT coverage.
// Removed gaps inside the range are collapsed (viral clip style).
func (m *Mapper) MapRange(sourceStartMs, sourceEndMs int64) (outputStartMs, outputEndMs int64, ok bool) {
	if sourceEndMs <= sourceStartMs {
		return 0, 0, false
	}
	var firstOut *int64
	var lastOut int64
	covered := false
	for _, s := range m.segments {
		overlapStart := max64(sourceStartMs, s.SourceStartMs)
		overlapEnd := min64(sourceEndMs, s.SourceEndMs)
		if overlapEnd <= overlapStart {
			continue
		}
		outStart := s.OutputStartMs + (overlapStart - s.SourceStartMs)
		outEnd := s.OutputStartMs + (overlapEnd - s.SourceStartMs)
		if firstOut == nil {
			v := outStart
			firstOut = &v
		}
		lastOut = outEnd
		covered = true
	}
	if !covered || firstOut == nil {
		return 0, 0, false
	}
	return *firstOut, lastOut, true
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
