package write

import (
	"context"

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
			MediaFileID:   m.MediaFileID,
			Type:          m.OverrideType,
			SourceStartMs: m.SourceStartMs,
			SourceEndMs:   m.SourceEndMs,
			Action:        m.Action,
		})
	}
	return out, nil
}
