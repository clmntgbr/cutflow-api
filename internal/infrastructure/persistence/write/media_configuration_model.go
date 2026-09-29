package write

import (
	"time"

	domainmediaconfig "go-api/internal/domain/mediaconfig"

	"github.com/google/uuid"
)

type MediaConfigurationModel struct {
	ID                           uuid.UUID `gorm:"column:id;primaryKey"`
	MediaFileID                  uuid.UUID `gorm:"column:media_file_id"`
	SilenceRemovalEnabled        bool      `gorm:"column:silence_removal_enabled"`
	SilenceThresholdMode         string    `gorm:"column:silence_threshold_mode"`
	SilenceThresholdDB           *float64  `gorm:"column:silence_threshold_db"`
	NoiseFloorDB                 *float64  `gorm:"column:noise_floor_db"`
	CalculatedSilenceThresholdDB *float64  `gorm:"column:calculated_silence_threshold_db"`
	SilenceDetectionLevel        string    `gorm:"column:silence_detection_level"`
	SilencePaddingBeforeMs       int       `gorm:"column:silence_padding_before_ms"`
	SilencePaddingAfterMs        int       `gorm:"column:silence_padding_after_ms"`
	SilenceMinDurationMs         int       `gorm:"column:silence_min_duration_ms"`
	SpeechMinDurationMs          int       `gorm:"column:speech_min_duration_ms"`
	FillerRemovalEnabled         bool      `gorm:"column:filler_removal_enabled"`
	RepetitionRemovalEnabled     bool      `gorm:"column:repetition_removal_enabled"`
	SubtitlesEnabled             bool      `gorm:"column:subtitles_enabled"`
	SubtitleMaxWords             int       `gorm:"column:subtitle_max_words"`
	ViralDetectionEnabled        bool      `gorm:"column:viral_detection_enabled"`
	ViralClipMinDurationMs       int       `gorm:"column:viral_clip_min_duration_ms"`
	ViralClipMaxDurationMs       int       `gorm:"column:viral_clip_max_duration_ms"`
	CreatedAt                    time.Time `gorm:"column:created_at"`
	UpdatedAt                    time.Time `gorm:"column:updated_at"`
}

func (MediaConfigurationModel) TableName() string { return "media_configuration" }

func mediaConfigurationModelFromDomain(c *domainmediaconfig.MediaConfiguration) *MediaConfigurationModel {
	return &MediaConfigurationModel{
		ID:                           c.ID,
		MediaFileID:                  c.MediaFileID,
		SilenceRemovalEnabled:        c.SilenceRemovalEnabled,
		SilenceThresholdMode:         c.SilenceThresholdMode,
		SilenceThresholdDB:           c.SilenceThresholdDB,
		NoiseFloorDB:                 c.NoiseFloorDB,
		CalculatedSilenceThresholdDB: c.CalculatedSilenceThresholdDB,
		SilenceDetectionLevel:        c.SilenceDetectionLevel,
		SilencePaddingBeforeMs:       c.SilencePaddingBeforeMs,
		SilencePaddingAfterMs:        c.SilencePaddingAfterMs,
		SilenceMinDurationMs:         c.SilenceMinDurationMs,
		SpeechMinDurationMs:          c.SpeechMinDurationMs,
		FillerRemovalEnabled:         c.FillerRemovalEnabled,
		RepetitionRemovalEnabled:     c.RepetitionRemovalEnabled,
		SubtitlesEnabled:             c.SubtitlesEnabled,
		SubtitleMaxWords:             c.SubtitleMaxWords,
		ViralDetectionEnabled:        c.ViralDetectionEnabled,
		ViralClipMinDurationMs:       c.ViralClipMinDurationMs,
		ViralClipMaxDurationMs:       c.ViralClipMaxDurationMs,
		CreatedAt:                    c.CreatedAt,
		UpdatedAt:                    c.UpdatedAt,
	}
}

func mediaConfigurationDomainFromModel(m *MediaConfigurationModel) *domainmediaconfig.MediaConfiguration {
	return &domainmediaconfig.MediaConfiguration{
		ID:                           m.ID,
		MediaFileID:                  m.MediaFileID,
		SilenceRemovalEnabled:        m.SilenceRemovalEnabled,
		SilenceThresholdMode:         m.SilenceThresholdMode,
		SilenceThresholdDB:           m.SilenceThresholdDB,
		NoiseFloorDB:                 m.NoiseFloorDB,
		CalculatedSilenceThresholdDB: m.CalculatedSilenceThresholdDB,
		SilenceDetectionLevel:        m.SilenceDetectionLevel,
		SilencePaddingBeforeMs:       m.SilencePaddingBeforeMs,
		SilencePaddingAfterMs:        m.SilencePaddingAfterMs,
		SilenceMinDurationMs:         m.SilenceMinDurationMs,
		SpeechMinDurationMs:          m.SpeechMinDurationMs,
		FillerRemovalEnabled:         m.FillerRemovalEnabled,
		RepetitionRemovalEnabled:     m.RepetitionRemovalEnabled,
		SubtitlesEnabled:             m.SubtitlesEnabled,
		SubtitleMaxWords:             m.SubtitleMaxWords,
		ViralDetectionEnabled:        m.ViralDetectionEnabled,
		ViralClipMinDurationMs:       m.ViralClipMinDurationMs,
		ViralClipMaxDurationMs:       m.ViralClipMaxDurationMs,
		CreatedAt:                    m.CreatedAt,
		UpdatedAt:                    m.UpdatedAt,
	}
}
