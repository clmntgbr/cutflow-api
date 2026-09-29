package mediaconfig

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	ThresholdModeAuto   = "auto"
	ThresholdModeManual = "manual"

	DetectionLevelLow             = "low"
	DetectionLevelModerate        = "moderate"
	DetectionLevelAggressive      = "aggressive"
	DetectionLevelVeryAggressive  = "very_aggressive"

	DefaultPaddingBeforeMs  = 50
	DefaultPaddingAfterMs   = 150
	DefaultMinSilenceMs     = 500
	DefaultMinSpeechMs      = 300
	DefaultSubtitleMaxWords = 3
	DefaultViralMinMs       = 20_000
	DefaultViralMaxMs       = 90_000
)

// LevelOffsetDB is added to the estimated noise floor to get the silence threshold.
// Higher offset → more aggressive (more audio counted as silence). Temporary calibration.
func LevelOffsetDB(level string) float64 {
	switch level {
	case DetectionLevelLow:
		return 6
	case DetectionLevelModerate:
		return 8
	case DetectionLevelVeryAggressive:
		return 14
	default: // aggressive
		return 10
	}
}

type MediaConfiguration struct {
	ID        uuid.UUID
	MediaFileID uuid.UUID

	SilenceRemovalEnabled         bool
	SilenceThresholdMode          string
	SilenceThresholdDB            *float64
	CalculatedSilenceThresholdDB  *float64
	SilenceDetectionLevel         string
	SilencePaddingBeforeMs        int
	SilencePaddingAfterMs         int
	SilenceMinDurationMs          int
	SpeechMinDurationMs           int

	FillerRemovalEnabled      bool
	RepetitionRemovalEnabled  bool

	SubtitlesEnabled  bool
	SubtitleMaxWords  int

	ViralDetectionEnabled   bool
	ViralClipMinDurationMs  int
	ViralClipMaxDurationMs  int

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewDefault(mediaFileID uuid.UUID) *MediaConfiguration {
	now := time.Now().UTC()
	return &MediaConfiguration{
		ID:                    uuid.New(),
		MediaFileID:           mediaFileID,
		SilenceRemovalEnabled: true,
		SilenceThresholdMode:  ThresholdModeAuto,
		SilenceDetectionLevel: DetectionLevelAggressive,
		SilencePaddingBeforeMs: DefaultPaddingBeforeMs,
		SilencePaddingAfterMs:  DefaultPaddingAfterMs,
		SilenceMinDurationMs:   DefaultMinSilenceMs,
		SpeechMinDurationMs:    DefaultMinSpeechMs,
		FillerRemovalEnabled:     true,
		RepetitionRemovalEnabled: true,
		SubtitlesEnabled:         true,
		SubtitleMaxWords:         DefaultSubtitleMaxWords,
		ViralDetectionEnabled:    true,
		ViralClipMinDurationMs:   DefaultViralMinMs,
		ViralClipMaxDurationMs:   DefaultViralMaxMs,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ResolveSilenceThreshold returns the dB threshold to use for silencedetect.
// In auto mode, noiseFloorDB comes from audio analysis (mean/noise estimate).
func (c *MediaConfiguration) ResolveSilenceThreshold(noiseFloorDB float64) float64 {
	if c.SilenceThresholdMode == ThresholdModeManual && c.SilenceThresholdDB != nil {
		return *c.SilenceThresholdDB
	}
	threshold := noiseFloorDB + LevelOffsetDB(c.SilenceDetectionLevel)
	if threshold < -60 {
		threshold = -60
	}
	if threshold > -20 {
		threshold = -20
	}
	return threshold
}

func (c *MediaConfiguration) SetCalculatedThreshold(db float64) {
	c.CalculatedSilenceThresholdDB = &db
	c.UpdatedAt = time.Now().UTC()
}

type MediaConfigurationWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, cfg *MediaConfiguration) error
	Update(ctx context.Context, cfg *MediaConfiguration) error
	GetByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (*MediaConfiguration, error)
}
