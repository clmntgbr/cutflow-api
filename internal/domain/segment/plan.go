package segment

import "math"

const (
	DefaultDurationMs = 5 * 60 * 1000 // 5 minutes
	DefaultOverlapMs  = 2 * 1000      // 2 seconds ASR/silence overlap
	DefaultMaxCount   = 40
)

type PlanConfig struct {
	DurationMs int64
	OverlapMs  int64
	MaxCount   int
}

func NormalizePlanConfig(cfg PlanConfig) PlanConfig {
	if cfg.DurationMs <= 0 {
		cfg.DurationMs = DefaultDurationMs
	}
	if cfg.OverlapMs < 0 {
		cfg.OverlapMs = 0
	}
	if cfg.MaxCount <= 0 {
		cfg.MaxCount = DefaultMaxCount
	}
	return cfg
}

// Window is a logical time range used to parallelize later workers.
// It is not a physical media file.
type Window struct {
	Index              int
	StartMs            int64
	EndMs              int64
	ProcessingStartMs  int64
	ProcessingEndMs    int64
}

// PlanWindows builds temporal windows covering [0, durationMs).
func PlanWindows(durationMs int64, cfg PlanConfig) []Window {
	cfg = NormalizePlanConfig(cfg)
	if durationMs <= 0 {
		return nil
	}

	count := int(math.Ceil(float64(durationMs) / float64(cfg.DurationMs)))
	effective := cfg.DurationMs
	if count > cfg.MaxCount {
		count = cfg.MaxCount
		effective = int64(math.Ceil(float64(durationMs) / float64(count)))
	}
	if count < 1 {
		count = 1
	}

	windows := make([]Window, 0, count)
	for i := 0; i < count; i++ {
		start := int64(i) * effective
		end := start + effective
		if i == count-1 || end > durationMs {
			end = durationMs
		}
		if end <= start {
			break
		}

		procStart := start - cfg.OverlapMs
		if procStart < 0 {
			procStart = 0
		}
		procEnd := end + cfg.OverlapMs
		if procEnd > durationMs {
			procEnd = durationMs
		}

		windows = append(windows, Window{
			Index:             i,
			StartMs:           start,
			EndMs:             end,
			ProcessingStartMs: procStart,
			ProcessingEndMs:   procEnd,
		})
	}
	return windows
}
