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

func (r *detectedTranscriptIssueWriteRepository) ListByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) ([]*domaintranscriptissue.Issue, error) {
	var models []DetectedTranscriptIssueModel
	if err := DBWithContext(ctx, r.db).
		Where("media_file_id = ?", mediaFileID).
		Order("source_start_ms ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domaintranscriptissue.Issue, 0, len(models))
	for _, m := range models {
		out = append(out, &domaintranscriptissue.Issue{
			ID:             m.ID,
			MediaFileID:    m.MediaFileID,
			TranscriptID:   m.TranscriptID,
			Type:           m.IssueType,
			Text:           m.Text,
			SourceStartMs:  m.SourceStartMs,
			SourceEndMs:    m.SourceEndMs,
			Confidence:     m.Confidence,
			WordStartIndex: m.WordStartIndex,
			WordEndIndex:   m.WordEndIndex,
			CreatedAt:      m.CreatedAt,
		})
	}
	return out, nil
}
