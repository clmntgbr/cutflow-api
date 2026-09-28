package project

import (
	"context"

	"github.com/google/uuid"
)

type ProjectWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, project *Project) error
	Update(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
}
