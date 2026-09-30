package read

import (
	"context"
	"errors"

	domainmediafile "go-api/internal/domain/mediafile"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mediaFileEditorRow struct {
	ID               uuid.UUID
	OriginalFilename *string `gorm:"column:original_filename"`
	StorageKey       string  `gorm:"column:storage_key"`
	DurationMs       int64   `gorm:"column:duration_ms"`
	Width            *int
	Height           *int
}

func (mediaFileEditorRow) TableName() string { return "media_file" }

func (r *mediaFileReadRepository) FindOwnedForEditor(
	ctx context.Context,
	id, userID uuid.UUID,
) (*domainmediafile.MediaFileEditorView, error) {
	var row mediaFileEditorRow
	err := r.db.WithContext(ctx).
		Select("id", "original_filename", "storage_key", "duration_ms", "width", "height").
		Where("id = ? AND user_id = ?", id, userID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	out := &domainmediafile.MediaFileEditorView{
		ID:         row.ID,
		StorageKey: row.StorageKey,
		DurationMs: row.DurationMs,
		Width:      row.Width,
		Height:     row.Height,
	}
	if row.OriginalFilename != nil {
		out.OriginalFilename = *row.OriginalFilename
	}
	return out, nil
}
