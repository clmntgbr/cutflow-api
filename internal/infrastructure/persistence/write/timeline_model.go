package write

import (
	"encoding/json"
	"time"

	domaintimeline "go-api/internal/domain/timeline"

	"github.com/google/uuid"
)

type TimelineModel struct {
	ID            uuid.UUID `gorm:"column:id;primaryKey"`
	ProjectID     uuid.UUID `gorm:"column:project_id"`
	MediaFileID   uuid.UUID `gorm:"column:media_file_id"`
	Version       int       `gorm:"column:version"`
	DurationMs    int64     `gorm:"column:duration_ms"`
	Fingerprint   string    `gorm:"column:fingerprint"`
	EngineVersion string    `gorm:"column:engine_version"`
	IsActive      bool      `gorm:"column:is_active"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (TimelineModel) TableName() string { return "timeline" }

type TimelineSegmentModel struct {
	ID            uuid.UUID `gorm:"column:id;primaryKey"`
	TimelineID    uuid.UUID `gorm:"column:timeline_id"`
	MediaFileID   uuid.UUID `gorm:"column:media_file_id"`
	SegmentIndex  int       `gorm:"column:segment_index"`
	SourceStartMs int64     `gorm:"column:source_start_ms"`
	SourceEndMs   int64     `gorm:"column:source_end_ms"`
	OutputStartMs int64     `gorm:"column:output_start_ms"`
	OutputEndMs   int64     `gorm:"column:output_end_ms"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (TimelineSegmentModel) TableName() string { return "timeline_segment" }

type EditDecisionModel struct {
	ID            uuid.UUID       `gorm:"column:id;primaryKey"`
	MediaFileID   uuid.UUID       `gorm:"column:media_file_id"`
	TimelineID    *uuid.UUID      `gorm:"column:timeline_id"`
	DecisionType  string          `gorm:"column:decision_type"`
	SourceStartMs int64           `gorm:"column:source_start_ms"`
	SourceEndMs   int64           `gorm:"column:source_end_ms"`
	Action        string          `gorm:"column:action"`
	Source        string          `gorm:"column:source"`
	Confidence    *float64        `gorm:"column:confidence"`
	Reasons       json.RawMessage `gorm:"column:reasons;type:jsonb"`
	CreatedAt     time.Time       `gorm:"column:created_at"`
}

func (EditDecisionModel) TableName() string { return "edit_decision" }

type UserOverrideModel struct {
	ID            uuid.UUID `gorm:"column:id;primaryKey"`
	MediaFileID   uuid.UUID `gorm:"column:media_file_id"`
	OverrideType  string    `gorm:"column:override_type"`
	SourceStartMs int64     `gorm:"column:source_start_ms"`
	SourceEndMs   int64     `gorm:"column:source_end_ms"`
	Action        string    `gorm:"column:action"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (UserOverrideModel) TableName() string { return "user_override" }

func timelineModelFromDomain(tl *domaintimeline.Timeline) *TimelineModel {
	return &TimelineModel{
		ID:            tl.ID,
		ProjectID:     tl.ProjectID,
		MediaFileID:   tl.MediaFileID,
		Version:       tl.Version,
		DurationMs:    tl.DurationMs,
		Fingerprint:   tl.Fingerprint,
		EngineVersion: tl.EngineVersion,
		IsActive:      tl.IsActive,
		CreatedAt:     tl.CreatedAt,
		UpdatedAt:     tl.UpdatedAt,
	}
}

func timelineDomainFromModel(m *TimelineModel, segments []domaintimeline.Segment, decisions []domaintimeline.Decision) *domaintimeline.Timeline {
	return &domaintimeline.Timeline{
		ID:            m.ID,
		ProjectID:     m.ProjectID,
		MediaFileID:   m.MediaFileID,
		Version:       m.Version,
		DurationMs:    m.DurationMs,
		Fingerprint:   m.Fingerprint,
		EngineVersion: m.EngineVersion,
		IsActive:      m.IsActive,
		Segments:      segments,
		Decisions:     decisions,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
