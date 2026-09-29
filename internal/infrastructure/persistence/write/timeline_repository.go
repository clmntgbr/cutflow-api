package write

import (
	"context"
	"encoding/json"
	"time"

	domaintimeline "go-api/internal/domain/timeline"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type timelineWriteRepository struct {
	db *gorm.DB
}

func NewTimelineWriteRepository(db *gorm.DB) domaintimeline.TimelineWriteRepository {
	return &timelineWriteRepository{db: db}
}

func (r *timelineWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *timelineWriteRepository) GetActiveByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (*domaintimeline.Timeline, error) {
	db := DBWithContext(ctx, r.db)
	var m TimelineModel
	err := db.Where("media_file_id = ? AND is_active = TRUE", mediaFileID).First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.loadTimeline(ctx, &m)
}

func (r *timelineWriteRepository) FindByFingerprint(ctx context.Context, mediaFileID uuid.UUID, fingerprint string) (*domaintimeline.Timeline, error) {
	db := DBWithContext(ctx, r.db)
	var m TimelineModel
	err := db.Where("media_file_id = ? AND fingerprint = ?", mediaFileID, fingerprint).First(&m).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.loadTimeline(ctx, &m)
}

func (r *timelineWriteRepository) ExistsByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (bool, error) {
	var count int64
	err := DBWithContext(ctx, r.db).
		Model(&TimelineModel{}).
		Where("media_file_id = ?", mediaFileID).
		Limit(1).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *timelineWriteRepository) NextVersion(ctx context.Context, mediaFileID uuid.UUID) (int, error) {
	db := DBWithContext(ctx, r.db)
	var max *int
	if err := db.Model(&TimelineModel{}).
		Select("MAX(version)").
		Where("media_file_id = ?", mediaFileID).
		Scan(&max).Error; err != nil {
		return 0, err
	}
	if max == nil {
		return 1, nil
	}
	return *max + 1, nil
}

func (r *timelineWriteRepository) Save(ctx context.Context, tl *domaintimeline.Timeline) error {
	db := DBWithContext(ctx, r.db)
	now := time.Now().UTC()
	if tl.CreatedAt.IsZero() {
		tl.CreatedAt = now
	}
	tl.UpdatedAt = now

	if err := db.Create(timelineModelFromDomain(tl)).Error; err != nil {
		return err
	}

	segModels := make([]*TimelineSegmentModel, 0, len(tl.Segments))
	for _, s := range tl.Segments {
		segModels = append(segModels, &TimelineSegmentModel{
			ID:            uuid.New(),
			TimelineID:    tl.ID,
			MediaFileID:   s.MediaFileID,
			SegmentIndex:  s.Index,
			SourceStartMs: s.SourceStartMs,
			SourceEndMs:   s.SourceEndMs,
			OutputStartMs: s.OutputStartMs,
			OutputEndMs:   s.OutputEndMs,
			CreatedAt:     now,
		})
	}
	if len(segModels) > 0 {
		if err := db.Create(&segModels).Error; err != nil {
			return err
		}
	}

	if err := db.Where("media_file_id = ?", tl.MediaFileID).Delete(&EditDecisionModel{}).Error; err != nil {
		return err
	}
	decModels := make([]*EditDecisionModel, 0, len(tl.Decisions))
	tlID := tl.ID
	for _, d := range tl.Decisions {
		reasons, _ := json.Marshal(d.Reasons)
		id := d.ID
		if id == uuid.Nil {
			id = uuid.New()
		}
		decModels = append(decModels, &EditDecisionModel{
			ID:            id,
			MediaFileID:   tl.MediaFileID,
			TimelineID:    &tlID,
			DecisionType:  d.Type,
			SourceStartMs: d.SourceStartMs,
			SourceEndMs:   d.SourceEndMs,
			Action:        d.Action,
			Source:        d.Source,
			Confidence:    d.Confidence,
			Reasons:       reasons,
			CreatedAt:     now,
		})
	}
	if len(decModels) > 0 {
		if err := db.Create(&decModels).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *timelineWriteRepository) DeactivateOthers(ctx context.Context, mediaFileID, keepID uuid.UUID) error {
	return DBWithContext(ctx, r.db).
		Model(&TimelineModel{}).
		Where("media_file_id = ? AND id <> ?", mediaFileID, keepID).
		Update("is_active", false).Error
}

func (r *timelineWriteRepository) Activate(ctx context.Context, id uuid.UUID) error {
	return DBWithContext(ctx, r.db).
		Model(&TimelineModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_active":  true,
			"updated_at": time.Now().UTC(),
		}).Error
}

func (r *timelineWriteRepository) loadTimeline(ctx context.Context, m *TimelineModel) (*domaintimeline.Timeline, error) {
	db := DBWithContext(ctx, r.db)
	var segModels []TimelineSegmentModel
	if err := db.Where("timeline_id = ?", m.ID).Order("segment_index ASC").Find(&segModels).Error; err != nil {
		return nil, err
	}
	segments := make([]domaintimeline.Segment, 0, len(segModels))
	for _, s := range segModels {
		segments = append(segments, domaintimeline.Segment{
			Index:         s.SegmentIndex,
			MediaFileID:   s.MediaFileID,
			SourceStartMs: s.SourceStartMs,
			SourceEndMs:   s.SourceEndMs,
			OutputStartMs: s.OutputStartMs,
			OutputEndMs:   s.OutputEndMs,
		})
	}

	var decModels []EditDecisionModel
	if err := db.Where("timeline_id = ?", m.ID).Order("source_start_ms ASC").Find(&decModels).Error; err != nil {
		return nil, err
	}
	decisions := make([]domaintimeline.Decision, 0, len(decModels))
	for _, d := range decModels {
		var reasons []string
		_ = json.Unmarshal(d.Reasons, &reasons)
		decisions = append(decisions, domaintimeline.Decision{
			ID:            d.ID,
			MediaFileID:   d.MediaFileID,
			Type:          d.DecisionType,
			SourceStartMs: d.SourceStartMs,
			SourceEndMs:   d.SourceEndMs,
			Action:        d.Action,
			Source:        d.Source,
			Confidence:    d.Confidence,
			Reasons:       reasons,
		})
	}
	return timelineDomainFromModel(m, segments, decisions), nil
}
