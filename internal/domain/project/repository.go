package project

import (
	"context"
	"time"

	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type ProjectWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, project *Project) error
	Update(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
}

type ProjectReadRepository interface {
	FindByID(ctx context.Context, id, userID uuid.UUID) (*ProjectDetailView, error)
	List(ctx context.Context, userID uuid.UUID, query paginate.PaginateQuery) ([]ProjectListView, int64, error)
}

// ProjectListView is a lean row for short list calls.
type ProjectListView struct {
	ID           uuid.UUID
	Name         string
	Status       string
	CreatedAt    time.Time
	MediaFileID  uuid.UUID
	ThumbnailKey string
}

type ProjectDetailView struct {
	ID         uuid.UUID
	Name       string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	MediaFiles []ProjectMediaFileView
}

type ProjectMediaFileView struct {
	ID               uuid.UUID
	OriginalFilename string
	MimeType         string
	SizeBytes        int64
	DurationMs       int64
	StorageKey       string
	OriginalURL      string
	ThumbnailKey     string
	Status           string
	CreatedAt        time.Time
}
