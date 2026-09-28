package write

import (
	"context"
	"errors"

	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectWriteRepository struct {
	db *gorm.DB
}

func NewProjectWriteRepository(db *gorm.DB) domainproject.ProjectWriteRepository {
	return &projectWriteRepository{db: db}
}

func (r *projectWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *projectWriteRepository) Save(ctx context.Context, project *domainproject.Project) error {
	return DBWithContext(ctx, r.db).Create(projectModelFromDomain(project)).Error
}

func (r *projectWriteRepository) Update(ctx context.Context, project *domainproject.Project) error {
	return DBWithContext(ctx, r.db).Save(projectModelFromDomain(project)).Error
}

func (r *projectWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainproject.Project, error) {
	var model ProjectModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return projectDomainFromModel(&model), nil
}
