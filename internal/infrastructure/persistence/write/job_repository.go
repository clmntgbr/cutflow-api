package write

import (
	"context"
	"errors"

	domainjob "go-api/internal/domain/job"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type jobWriteRepository struct {
	db *gorm.DB
}

func NewJobWriteRepository(db *gorm.DB) domainjob.JobWriteRepository {
	return &jobWriteRepository{db: db}
}

func (r *jobWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *jobWriteRepository) Save(ctx context.Context, job *domainjob.Job) error {
	return DBWithContext(ctx, r.db).Create(jobModelFromDomain(job)).Error
}

func (r *jobWriteRepository) Update(ctx context.Context, job *domainjob.Job) error {
	return DBWithContext(ctx, r.db).Save(jobModelFromDomain(job)).Error
}

func (r *jobWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainjob.Job, error) {
	db := DBWithContext(ctx, r.db)
	if _, ok := ctx.Value(txCtxKey{}).(*gorm.DB); ok {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var model JobModel
	err := db.Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return jobDomainFromModel(&model), nil
}

func (r *jobWriteRepository) HasSuccessful(ctx context.Context, mediaFileID uuid.UUID, name string) (bool, error) {
	var count int64
	err := DBWithContext(ctx, r.db).
		Model(&JobModel{}).
		Where("media_file_id = ? AND name = ? AND status = ?", mediaFileID, name, domainjob.StatusSuccess).
		Limit(1).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
