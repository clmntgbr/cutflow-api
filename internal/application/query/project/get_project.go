package project

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-api/internal/domain/port"
	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
)

const mediaURLTTL = time.Hour

type GetProjectByIDQuery struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

type GetProjectByIDHandler struct {
	readRepo domainproject.ProjectReadRepository
	storage  port.Storage
}

func NewGetProjectByIDHandler(
	readRepo domainproject.ProjectReadRepository,
	storage port.Storage,
) *GetProjectByIDHandler {
	return &GetProjectByIDHandler{readRepo: readRepo, storage: storage}
}

func (h *GetProjectByIDHandler) Handle(
	ctx context.Context,
	q GetProjectByIDQuery,
) (*domainproject.ProjectDetailView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID, q.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if view == nil {
		return nil, domainproject.ErrProjectNotFound
	}

	for i := range view.MediaFiles {
		media := &view.MediaFiles[i]
		media.OriginalURL = presignMediaURL(ctx, h.storage, media.ID, media.StorageKey)
	}
	return view, nil
}

func presignMediaURL(ctx context.Context, storage port.Storage, mediaFileID uuid.UUID, key string) string {
	if key == "" || storage == nil {
		return ""
	}
	url, err := storage.PresignedGetURL(ctx, key, mediaURLTTL)
	if err != nil {
		log.Printf("failed to presign media url for media file %s: %v", mediaFileID, err)
		return ""
	}
	return url
}
