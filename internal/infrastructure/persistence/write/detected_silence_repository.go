package write

import (
	"context"

	domainsilence "go-api/internal/domain/silence"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type detectedSilenceWriteRepository struct {
	db *gorm.DB
}

func NewDetectedSilenceWriteRepository(db *gorm.DB) domainsilence.DetectedSilenceWriteRepository {
	return &detectedSilenceWriteRepository{db: db}
}

func (r *detectedSilenceWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *detectedSilenceWriteRepository) ReplaceForMediaFile(
	ctx context.Context,
	mediaFileID uuid.UUID,
	rows []*domainsilence.DetectedSilence,
) error {
	db := DBWithContext(ctx, r.db)
	if err := db.Where("media_file_id = ?", mediaFileID).Delete(&DetectedSilenceModel{}).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	models := make([]*DetectedSilenceModel, 0, len(rows))
	for _, row := range rows {
		models = append(models, detectedSilenceModelFromDomain(row))
	}
	return db.Create(&models).Error
}

func (r *detectedSilenceWriteRepository) CountByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (int64, error) {
	var total int64
	err := DBWithContext(ctx, r.db).
		Model(&DetectedSilenceModel{}).
		Where("media_file_id = ?", mediaFileID).
		Count(&total).Error
	return total, err
}
