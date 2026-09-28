package read

import (
	"context"
	"errors"

	domainmediafile "go-api/internal/domain/mediafile"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mediaFileThumbnailRow struct {
	ID           uuid.UUID
	ThumbnailKey string
}

func (mediaFileThumbnailRow) TableName() string { return "media_file" }

type mediaFileReadRepository struct {
	db *gorm.DB
}

func NewMediaFileReadRepository(db *gorm.DB) domainmediafile.MediaFileReadRepository {
	return &mediaFileReadRepository{db: db}
}

func (r *mediaFileReadRepository) FindOwnedByID(
	ctx context.Context,
	id, userID uuid.UUID,
) (*domainmediafile.MediaFileThumbnailView, error) {
	var row mediaFileThumbnailRow
	err := r.db.WithContext(ctx).
		Select("id", "thumbnail_key").
		Where("id = ? AND user_id = ?", id, userID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domainmediafile.MediaFileThumbnailView{
		ID:           row.ID,
		ThumbnailKey: row.ThumbnailKey,
	}, nil
}
