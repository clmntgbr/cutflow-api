package project

import (
	"context"
	"errors"
	"time"

	"go-api/internal/domain/event"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"
	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
)

const DefaultUploadURLTTL = 15 * time.Minute

type RequestUploadURLCommand struct {
	UserID      uuid.UUID
	Filename    string
	ContentType string
	SizeBytes   int64
}

type RequestUploadURLResult struct {
	ProjectID   string
	MediaFileID string
	UploadURL   string
	ExpiresAt   time.Time
}

type RequestUploadURLHandler struct {
	projectRepo domainproject.ProjectWriteRepository
	mediaRepo   domainmediafile.MediaFileWriteRepository
	outbox      port.OutboxRepository
	storage     port.Storage
	ttl         time.Duration
	maxSize     int64
}

func NewRequestUploadURLHandler(
	projectRepo domainproject.ProjectWriteRepository,
	mediaRepo domainmediafile.MediaFileWriteRepository,
	outbox port.OutboxRepository,
	storage port.Storage,
	ttl time.Duration,
	maxSize int64,
) *RequestUploadURLHandler {
	if ttl <= 0 {
		ttl = DefaultUploadURLTTL
	}
	return &RequestUploadURLHandler{
		projectRepo: projectRepo,
		mediaRepo:   mediaRepo,
		outbox:      outbox,
		storage:     storage,
		ttl:         ttl,
		maxSize:     maxSize,
	}
}

func (h *RequestUploadURLHandler) Handle(ctx context.Context, cmd RequestUploadURLCommand) (*RequestUploadURLResult, error) {
	if h.maxSize > 0 && cmd.SizeBytes > h.maxSize {
		return nil, domainmediafile.ErrMediaTooLarge
	}

	project, err := domainproject.NewProject(cmd.UserID, domainproject.NameFromFilename(cmd.Filename))
	if err != nil {
		return nil, err
	}

	media, err := domainmediafile.NewPresignedMediaFile(
		project.ID,
		cmd.UserID,
		cmd.Filename,
		cmd.ContentType,
		cmd.SizeBytes,
	)
	if err != nil {
		return nil, err
	}

	url, err := h.storage.PresignedPutURL(ctx, media.StorageKey, h.ttl)
	if err != nil {
		return nil, err
	}

	err = h.projectRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.projectRepo.Save(txCtx, project); err != nil {
			return err
		}
		if err := h.mediaRepo.Save(txCtx, media); err != nil {
			return err
		}
		events := append([]event.DomainEvent{}, project.PullEvents()...)
		events = append(events, media.PullEvents()...)
		return h.outbox.StoreEvents(txCtx, events)
	})
	if err != nil {
		return nil, errors.New("failed to create project upload")
	}

	return &RequestUploadURLResult{
		ProjectID:   project.ID.String(),
		MediaFileID: media.ID.String(),
		UploadURL:   url,
		ExpiresAt:   media.CreatedAt.Add(h.ttl),
	}, nil
}
