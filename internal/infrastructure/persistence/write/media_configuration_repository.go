package write

import (
	"context"
	"errors"

	domainmediaconfig "go-api/internal/domain/mediaconfig"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type mediaConfigurationWriteRepository struct {
	db *gorm.DB
}

func NewMediaConfigurationWriteRepository(db *gorm.DB) domainmediaconfig.MediaConfigurationWriteRepository {
	return &mediaConfigurationWriteRepository{db: db}
}

func (r *mediaConfigurationWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *mediaConfigurationWriteRepository) Save(ctx context.Context, cfg *domainmediaconfig.MediaConfiguration) error {
	return DBWithContext(ctx, r.db).Create(mediaConfigurationModelFromDomain(cfg)).Error
}

func (r *mediaConfigurationWriteRepository) Update(ctx context.Context, cfg *domainmediaconfig.MediaConfiguration) error {
	return DBWithContext(ctx, r.db).Save(mediaConfigurationModelFromDomain(cfg)).Error
}

func (r *mediaConfigurationWriteRepository) GetByMediaFileID(
	ctx context.Context,
	mediaFileID uuid.UUID,
) (*domainmediaconfig.MediaConfiguration, error) {
	db := DBWithContext(ctx, r.db)
	if _, ok := ctx.Value(txCtxKey{}).(*gorm.DB); ok {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var model MediaConfigurationModel
	err := db.Where("media_file_id = ?", mediaFileID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mediaConfigurationDomainFromModel(&model), nil
}
