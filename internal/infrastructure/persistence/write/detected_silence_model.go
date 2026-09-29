package write

import (
	"time"

	domainsilence "go-api/internal/domain/silence"

	"github.com/google/uuid"
)

type DetectedSilenceModel struct {
	ID          uuid.UUID `gorm:"column:id;primaryKey"`
	MediaFileID uuid.UUID `gorm:"column:media_file_id"`
	StartMs     int64     `gorm:"column:start_ms"`
	EndMs       int64     `gorm:"column:end_ms"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (DetectedSilenceModel) TableName() string { return "detected_silence" }

func detectedSilenceModelFromDomain(s *domainsilence.DetectedSilence) *DetectedSilenceModel {
	return &DetectedSilenceModel{
		ID:          s.ID,
		MediaFileID: s.MediaFileID,
		StartMs:     s.StartMs,
		EndMs:       s.EndMs,
		CreatedAt:   s.CreatedAt,
	}
}
