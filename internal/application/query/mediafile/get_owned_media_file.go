package mediafile

import (
	"context"
	"fmt"

	domainmediafile "go-api/internal/domain/mediafile"

	"github.com/google/uuid"
)

type GetOwnedMediaFileQuery struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

type GetOwnedMediaFileHandler struct {
	readRepo domainmediafile.MediaFileReadRepository
}

func NewGetOwnedMediaFileHandler(readRepo domainmediafile.MediaFileReadRepository) *GetOwnedMediaFileHandler {
	return &GetOwnedMediaFileHandler{readRepo: readRepo}
}

func (h *GetOwnedMediaFileHandler) Handle(
	ctx context.Context,
	q GetOwnedMediaFileQuery,
) (*domainmediafile.MediaFileThumbnailView, error) {
	view, err := h.readRepo.FindOwnedByID(ctx, q.ID, q.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get media file: %w", err)
	}
	if view == nil {
		return nil, domainmediafile.ErrMediaNotFound
	}
	return view, nil
}
