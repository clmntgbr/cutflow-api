package write

import (
	"time"

	domainsegment "go-api/internal/domain/segment"

	"github.com/google/uuid"
)

type SegmentModel struct {
	ID                uuid.UUID `gorm:"column:id;primaryKey"`
	MediaFileID       uuid.UUID `gorm:"column:media_file_id"`
	SegmentIndex      int       `gorm:"column:segment_index"`
	StartMs           int64     `gorm:"column:start_ms"`
	EndMs             int64     `gorm:"column:end_ms"`
	ProcessingStartMs int64     `gorm:"column:processing_start_ms"`
	ProcessingEndMs   int64     `gorm:"column:processing_end_ms"`
	Status            string    `gorm:"column:status"`
	CreatedAt         time.Time `gorm:"column:created_at"`
}

func (SegmentModel) TableName() string { return "segment" }

func segmentModelFromDomain(s *domainsegment.Segment) *SegmentModel {
	return &SegmentModel{
		ID:                s.ID,
		MediaFileID:       s.MediaFileID,
		SegmentIndex:      s.SegmentIndex,
		StartMs:           s.StartMs,
		EndMs:             s.EndMs,
		ProcessingStartMs: s.ProcessingStartMs,
		ProcessingEndMs:   s.ProcessingEndMs,
		Status:            s.Status,
		CreatedAt:         s.CreatedAt,
	}
}
