package write

import (
	"context"
	"time"

	domaintimeline "go-api/internal/domain/timeline"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userOverrideRepository struct {
	db *gorm.DB
}

func NewUserOverrideRepository(db *gorm.DB) *userOverrideRepository {
	return &userOverrideRepository{db: db}
}

func (r *userOverrideRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *userOverrideRepository) ListByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) ([]domaintimeline.Override, error) {
	var models []UserOverrideModel
	if err := DBWithContext(ctx, r.db).
		Where("media_file_id = ?", mediaFileID).
		Order("source_start_ms ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domaintimeline.Override, 0, len(models))
	for _, m := range models {
		out = append(out, domaintimeline.Override{
			ID:            m.ID,
			MediaFileID:   m.MediaFileID,
			Type:          m.OverrideType,
			SourceStartMs: m.SourceStartMs,
			SourceEndMs:   m.SourceEndMs,
			Action:        m.Action,
		})
	}
	return out, nil
}

func (r *userOverrideRepository) Save(ctx context.Context, o domaintimeline.Override) error {
	now := time.Now().UTC()
	id := o.ID
	if id == uuid.Nil {
		id = uuid.New()
	}
	model := &UserOverrideModel{
		ID:            id,
		MediaFileID:   o.MediaFileID,
		OverrideType:  o.Type,
		SourceStartMs: o.SourceStartMs,
		SourceEndMs:   o.SourceEndMs,
		Action:        o.Action,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if model.OverrideType == "" {
		model.OverrideType = domaintimeline.DecisionManual
	}
	return DBWithContext(ctx, r.db).Create(model).Error
}

func (r *userOverrideRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return DBWithContext(ctx, r.db).Where("id = ?", id).Delete(&UserOverrideModel{}).Error
}

func (r *userOverrideRepository) DeleteMatching(
	ctx context.Context,
	mediaFileID uuid.UUID,
	startMs, endMs int64,
	action string,
) error {
	return DBWithContext(ctx, r.db).
		Where(
			"media_file_id = ? AND source_start_ms = ? AND source_end_ms = ? AND action = ?",
			mediaFileID, startMs, endMs, action,
		).
		Delete(&UserOverrideModel{}).Error
}
