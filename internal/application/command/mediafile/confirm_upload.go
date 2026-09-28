package mediafile

import (
	"context"

	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"
)

type ConfirmUploadCommand struct {
	StorageKey  string
	ContentType string
	SizeBytes   int64
}

type ConfirmUploadHandler struct {
	mediaRepo domainmediafile.MediaFileWriteRepository
	outbox    port.OutboxRepository
	maxSize   int64
}

func NewConfirmUploadHandler(
	mediaRepo domainmediafile.MediaFileWriteRepository,
	outbox port.OutboxRepository,
	maxSize int64,
) *ConfirmUploadHandler {
	return &ConfirmUploadHandler{
		mediaRepo: mediaRepo,
		outbox:    outbox,
		maxSize:   maxSize,
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

		if err := h.mediaRepo.Update(txCtx, media); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, media.PullEvents())
	})
}
