package mediafile

import (
	"context"

	"go-api/internal/domain/event"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"
	domainproject "go-api/internal/domain/project"
)

type ConfirmUploadCommand struct {
	StorageKey  string
	ContentType string
	SizeBytes   int64
}

type ConfirmUploadHandler struct {
	mediaRepo   domainmediafile.MediaFileWriteRepository
	projectRepo domainproject.ProjectWriteRepository
	outbox      port.OutboxRepository
	maxSize     int64
}

func NewConfirmUploadHandler(
	mediaRepo domainmediafile.MediaFileWriteRepository,
	projectRepo domainproject.ProjectWriteRepository,
	outbox port.OutboxRepository,
	maxSize int64,
) *ConfirmUploadHandler {
	return &ConfirmUploadHandler{
		mediaRepo:   mediaRepo,
		projectRepo: projectRepo,
		outbox:      outbox,
		maxSize:     maxSize,
	}
}

func (h *ConfirmUploadHandler) Handle(ctx context.Context, cmd ConfirmUploadCommand) error {
	if h.maxSize > 0 && cmd.SizeBytes > h.maxSize {
		return domainmediafile.ErrMediaTooLarge
	}

	return h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		media, err := h.mediaRepo.GetByStorageKey(txCtx, cmd.StorageKey)
		if err != nil {
			return err
		}
		if media == nil {
			return domainmediafile.ErrMediaNotFound
		}

		if err := media.MarkUploaded(cmd.ContentType, cmd.SizeBytes); err != nil {
			return err
		}

		project, err := h.projectRepo.GetByID(txCtx, media.ProjectID)
		if err != nil {
			return err
		}
		if project == nil {
			return domainproject.ErrProjectNotFound
		}
		if err := project.MarkProcessing(); err != nil {
			return err
		}

		if err := h.mediaRepo.Update(txCtx, media); err != nil {
			return err
		}
		if err := h.projectRepo.Update(txCtx, project); err != nil {
			return err
		}

		events := append([]event.DomainEvent{}, media.PullEvents()...)
		events = append(events, project.PullEvents()...)
		return h.outbox.StoreEvents(txCtx, events)
	})
}
