package write

import (
	"context"

	domainviral "go-api/internal/domain/viral"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type viralCandidateWriteRepository struct {
	db *gorm.DB
}

func NewViralCandidateWriteRepository(db *gorm.DB) domainviral.CandidateWriteRepository {
	return &viralCandidateWriteRepository{db: db}
}

func (r *viralCandidateWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *viralCandidateWriteRepository) ReplaceForMediaFile(
	ctx context.Context,
	mediaFileID uuid.UUID,
	rows []*domainviral.Candidate,
) error {
	db := DBWithContext(ctx, r.db)
	if err := db.Where("media_file_id = ?", mediaFileID).Delete(&ViralCandidateModel{}).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	models := make([]*ViralCandidateModel, 0, len(rows))
	for _, row := range rows {
		models = append(models, viralCandidateModelFromDomain(row))
	}
	return db.Create(&models).Error
}

func (r *viralCandidateWriteRepository) CountByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (int64, error) {
	var total int64
	err := DBWithContext(ctx, r.db).
		Model(&ViralCandidateModel{}).
		Where("media_file_id = ?", mediaFileID).
		Count(&total).Error
	return total, err
}
