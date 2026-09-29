package viral

import "fmt"

// ValidateConfig holds product constraints for candidate acceptance.
type ValidateConfig struct {
	MediaDurationMs int64
	MinDurationMs   int64
	MaxDurationMs   int64
	MinScore        float64
}

// ValidateCandidate checks SOURCE TIME bounds and configured duration/score.
func ValidateCandidate(c *Candidate, cfg ValidateConfig) error {
	if c == nil {
		return fmt.Errorf("candidate is nil")
	}
	if c.SourceStartMs < 0 {
		return fmt.Errorf("start_ms < 0")
	}
	if c.SourceEndMs <= c.SourceStartMs {
		return fmt.Errorf("end_ms <= start_ms")
	}
	if cfg.MediaDurationMs > 0 {
		if c.SourceStartMs >= cfg.MediaDurationMs {
			return fmt.Errorf("start beyond media duration")
		}
		if c.SourceEndMs > cfg.MediaDurationMs {
			return fmt.Errorf("end beyond media duration")
		}
	}
	dur := c.DurationMs()
	if cfg.MinDurationMs > 0 && dur < cfg.MinDurationMs {
		return fmt.Errorf("duration %dms < min %dms", dur, cfg.MinDurationMs)
	}
	if cfg.MaxDurationMs > 0 && dur > cfg.MaxDurationMs {
		return fmt.Errorf("duration %dms > max %dms", dur, cfg.MaxDurationMs)
	}
	if c.Score < 0 || c.Score > 1 {
		return fmt.Errorf("score out of range")
	}
	if cfg.MinScore > 0 && c.Score < cfg.MinScore {
		return fmt.Errorf("score %.3f < min %.3f", c.Score, cfg.MinScore)
	}
	return nil
}

// FilterValid keeps candidates that pass validation (after boundary resolve).
func FilterValid(candidates []*Candidate, cfg ValidateConfig) []*Candidate {
	out := make([]*Candidate, 0, len(candidates))
	for _, c := range candidates {
		if err := ValidateCandidate(c, cfg); err != nil {
			continue
		}
		out = append(out, c)
	}
	return out
}
