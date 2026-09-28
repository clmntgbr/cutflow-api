package mediafile

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"

	"go-api/internal/application/messaging"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type GenerateThumbnailCommand struct {
	MediaFileID uuid.UUID
}

type GenerateThumbnailHandler struct {
	mediaRepo domainmediafile.MediaFileWriteRepository
	storage   port.Storage
	extractor port.FrameExtractor
	outbox    port.OutboxRepository
}

func NewGenerateThumbnailHandler(
	mediaRepo domainmediafile.MediaFileWriteRepository,
	storage port.Storage,
	extractor port.FrameExtractor,
	outbox port.OutboxRepository,
) *GenerateThumbnailHandler {
	return &GenerateThumbnailHandler{
		mediaRepo: mediaRepo,
		storage:   storage,
		extractor: extractor,
		outbox:    outbox,
	}
}

func (h *GenerateThumbnailHandler) Handle(ctx context.Context, cmd GenerateThumbnailCommand) error {
	media, err := h.mediaRepo.GetByID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if media == nil {
		return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
	}
	if media.ThumbnailKey != "" {
		return nil
	}

	tmp, err := os.CreateTemp("", "media-thumb-src-*")
	if err != nil {
		return messaging.Retryable(err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	reader, err := h.storage.Get(ctx, media.StorageKey)
	if err != nil {
		return messaging.Retryable(err)
	}
	if _, err := io.Copy(tmp, reader); err != nil {
		_ = reader.Close()
		return messaging.Retryable(err)
	}
	_ = reader.Close()
	if err := tmp.Close(); err != nil {
		return messaging.Retryable(err)
	}

	if err := h.storeThumbnail(ctx, media, tmp.Name()); err != nil {
		log.Printf("failed to store thumbnail for media file %s: %v", media.ID, err)
		return messaging.Retryable(err)
	}
	return nil
}

func (h *GenerateThumbnailHandler) storeThumbnail(
	ctx context.Context,
	media *domainmediafile.MediaFile,
	videoPath string,
) error {
	if h.extractor == nil {
		return nil
	}
	data, err := h.extractor.ExtractThumbnail(ctx, videoPath)
	if err != nil {
		return err
	}
	key := domainmediafile.NewThumbnailStorageKey(media.ID)
	if err := h.storage.Put(ctx, key, bytes.NewReader(data), int64(len(data)), "image/jpeg"); err != nil {
		return err
	}

	media.SetThumbnailKey(key)
	return h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.mediaRepo.UpdateThumbnailKey(txCtx, media.ID, key); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, media.PullEvents())
	})
}
