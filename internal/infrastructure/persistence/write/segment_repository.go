package write

import (
	"context"

	domainsegment "go-api/internal/domain/segment"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type segmentWriteRepository struct {
	db *gorm.DB
}

func NewSegmentWriteRepository(db *gorm.DB) domainsegment.SegmentWriteRepository {
	return &segmentWriteRepository{db: db}
}

func (r *segmentWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *segmentWriteRepository) SaveBatch(ctx context.Context, segments []*domainsegment.Segment) error {
	if len(segments) == 0 {
		return nil
	}
	models := make([]*SegmentModel, 0, len(segments))
	for _, segment := range segments {
		models = append(models, segmentModelFromDomain(segment))
	}
	return DBWithContext(ctx, r.db).Create(&models).Error
}

func (r *segmentWriteRepository) CountByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (int64, error) {
	var total int64
	err := DBWithContext(ctx, r.db).
		Model(&SegmentModel{}).
		Where("media_file_id = ?", mediaFileID).
		Count(&total).Error
	return total, err
}
