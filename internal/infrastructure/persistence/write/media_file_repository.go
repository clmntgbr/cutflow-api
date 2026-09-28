package write

import (
	"context"
	"errors"
	"time"

	domainmediafile "go-api/internal/domain/mediafile"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type mediaFileWriteRepository struct {
	db *gorm.DB
}

func NewMediaFileWriteRepository(db *gorm.DB) domainmediafile.MediaFileWriteRepository {
	return &mediaFileWriteRepository{db: db}
}

func (r *mediaFileWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *mediaFileWriteRepository) Save(ctx context.Context, media *domainmediafile.MediaFile) error {
	return DBWithContext(ctx, r.db).Create(mediaFileModelFromDomain(media)).Error
}

func (r *mediaFileWriteRepository) Update(ctx context.Context, media *domainmediafile.MediaFile) error {
	return DBWithContext(ctx, r.db).Save(mediaFileModelFromDomain(media)).Error
}

func (r *mediaFileWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainmediafile.MediaFile, error) {
	return r.get(ctx, "id = ?", id)
}

func (r *mediaFileWriteRepository) GetByStorageKey(ctx context.Context, storageKey string) (*domainmediafile.MediaFile, error) {
	return r.get(ctx, "storage_key = ?", storageKey)
}

func (r *mediaFileWriteRepository) UpdateThumbnailKey(ctx context.Context, id uuid.UUID, thumbnailKey string) error {
	return DBWithContext(ctx, r.db).
		Model(&MediaFileModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"thumbnail_key": thumbnailKey,
			"updated_at":    time.Now().UTC(),
		}).Error
}

func (r *mediaFileWriteRepository) ListExpiredPending(ctx context.Context, cutoff time.Time) ([]*domainmediafile.MediaFile, error) {
	var models []MediaFileModel
	err := DBWithContext(ctx, r.db).
		Where("status = ? AND created_at <= ?", domainmediafile.StatusPending, cutoff).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	mediaFiles := make([]*domainmediafile.MediaFile, 0, len(models))
	for i := range models {
		mediaFiles = append(mediaFiles, mediaFileDomainFromModel(&models[i]))
	}
	return mediaFiles, nil
}

func (r *mediaFileWriteRepository) get(ctx context.Context, query string, args ...any) (*domainmediafile.MediaFile, error) {
	db := DBWithContext(ctx, r.db)
	if _, ok := ctx.Value(txCtxKey{}).(*gorm.DB); ok {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	var model MediaFileModel
	err := db.Where(query, args...).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mediaFileDomainFromModel(&model), nil
}
