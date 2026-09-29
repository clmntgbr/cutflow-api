package write

import (
	"context"
	"errors"

	domainmediaaudio "go-api/internal/domain/mediaaudio"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type mediaAudioWriteRepository struct {
	db *gorm.DB
}

func NewMediaAudioWriteRepository(db *gorm.DB) domainmediaaudio.MediaAudioWriteRepository {
	return &mediaAudioWriteRepository{db: db}
}

func (r *mediaAudioWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *mediaAudioWriteRepository) Save(ctx context.Context, audio *domainmediaaudio.MediaAudio) error {
	return DBWithContext(ctx, r.db).Create(mediaAudioModelFromDomain(audio)).Error
}

func (r *mediaAudioWriteRepository) Update(ctx context.Context, audio *domainmediaaudio.MediaAudio) error {
	return DBWithContext(ctx, r.db).Save(mediaAudioModelFromDomain(audio)).Error
}

func (r *mediaAudioWriteRepository) GetByMediaFileID(
	ctx context.Context,
	mediaFileID uuid.UUID,
) (*domainmediaaudio.MediaAudio, error) {
	db := DBWithContext(ctx, r.db)
	if _, ok := ctx.Value(txCtxKey{}).(*gorm.DB); ok {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	var model MediaAudioModel
	err := db.Where("media_file_id = ?", mediaFileID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mediaAudioDomainFromModel(&model), nil
}
