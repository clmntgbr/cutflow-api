package read

import (
	"context"
	"errors"

	querymediafile "go-api/internal/application/query/mediafile"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type editorTimelineRow struct {
	ID         uuid.UUID
	Version    int
	DurationMs int64 `gorm:"column:duration_ms"`
}

func (editorTimelineRow) TableName() string { return "timeline" }

type editorSegmentRow struct {
	ID            uuid.UUID
	SegmentIndex  int   `gorm:"column:segment_index"`
	SourceStartMs int64 `gorm:"column:source_start_ms"`
	SourceEndMs   int64 `gorm:"column:source_end_ms"`
	OutputStartMs int64 `gorm:"column:output_start_ms"`
	OutputEndMs   int64 `gorm:"column:output_end_ms"`
}

func (editorSegmentRow) TableName() string { return "timeline_segment" }

type editorTimelineRepository struct {
	db *gorm.DB
}

func NewEditorTimelineRepository(db *gorm.DB) *editorTimelineRepository {
	return &editorTimelineRepository{db: db}
}

func (r *editorTimelineRepository) GetActiveWithSegments(
	ctx context.Context,
	mediaFileID uuid.UUID,
) (*querymediafile.EditorTimelineView, error) {
	var row editorTimelineRow
	err := r.db.WithContext(ctx).
		Select("id", "version", "duration_ms").
		Where("media_file_id = ? AND is_active = TRUE", mediaFileID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var segs []editorSegmentRow
	if err := r.db.WithContext(ctx).
		Select("id", "segment_index", "source_start_ms", "source_end_ms", "output_start_ms", "output_end_ms").
		Where("timeline_id = ?", row.ID).
		Order("segment_index ASC").
		Find(&segs).Error; err != nil {
		return nil, err
	}

	segments := make([]querymediafile.EditorTimelineSegmentView, 0, len(segs))
	for _, s := range segs {
		segments = append(segments, querymediafile.EditorTimelineSegmentView{
			ID:            s.ID,
			Index:         s.SegmentIndex,
			SourceStartMs: s.SourceStartMs,
			SourceEndMs:   s.SourceEndMs,
			OutputStartMs: s.OutputStartMs,
			OutputEndMs:   s.OutputEndMs,
		})
	}

	return &querymediafile.EditorTimelineView{
		ID:         row.ID,
		Version:    row.Version,
		DurationMs: row.DurationMs,
		Segments:   segments,
	}, nil
}
