package mediafile

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MediaFileWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, media *MediaFile) error
	Update(ctx context.Context, media *MediaFile) error
	GetByID(ctx context.Context, id uuid.UUID) (*MediaFile, error)
	GetByStorageKey(ctx context.Context, storageKey string) (*MediaFile, error)
	UpdateThumbnailKey(ctx context.Context, id uuid.UUID, thumbnailKey string) error
	ListExpiredPending(ctx context.Context, cutoff time.Time) ([]*MediaFile, error)
}
