package write

import (
	"context"

	domaintranscriptissue "go-api/internal/domain/transcriptissue"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type detectedTranscriptIssueWriteRepository struct {
	db *gorm.DB
}

func NewDetectedTranscriptIssueWriteRepository(db *gorm.DB) domaintranscriptissue.IssueWriteRepository {
	return &detectedTranscriptIssueWriteRepository{db: db}
}

func (r *detectedTranscriptIssueWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *detectedTranscriptIssueWriteRepository) ReplaceForMediaFile(
	ctx context.Context,
	mediaFileID uuid.UUID,
	rows []*domaintranscriptissue.Issue,
) error {
	db := DBWithContext(ctx, r.db)
	if err := db.Where("media_file_id = ?", mediaFileID).Delete(&DetectedTranscriptIssueModel{}).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	models := make([]*DetectedTranscriptIssueModel, 0, len(rows))
	for _, row := range rows {
		models = append(models, detectedTranscriptIssueModelFromDomain(row))
	}
	return db.Create(&models).Error
}

func (r *detectedTranscriptIssueWriteRepository) CountByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (int64, error) {
	var total int64
	err := DBWithContext(ctx, r.db).
		Model(&DetectedTranscriptIssueModel{}).
		Where("media_file_id = ?", mediaFileID).
		Count(&total).Error
	return total, err
}
