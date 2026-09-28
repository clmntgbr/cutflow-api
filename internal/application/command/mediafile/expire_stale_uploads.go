package mediafile

import (
	"context"
	"log"
	"time"

	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"
)

type ExpireStaleUploadsHandler struct {
	repo   domainmediafile.MediaFileWriteRepository
	outbox port.OutboxRepository
	ttl    time.Duration
}

func NewExpireStaleUploadsHandler(
	repo domainmediafile.MediaFileWriteRepository,
	outbox port.OutboxRepository,
	ttl time.Duration,
) *ExpireStaleUploadsHandler {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return &ExpireStaleUploadsHandler{repo: repo, outbox: outbox, ttl: ttl}
}

func (h *ExpireStaleUploadsHandler) Handle(ctx context.Context) error {
	now := time.Now().UTC()
	mediaFiles, err := h.repo.ListExpiredPending(ctx, now.Add(-h.ttl))
	if err != nil {
		return err
	}

	for _, media := range mediaFiles {
		err := h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
			current, err := h.repo.GetByID(txCtx, media.ID)
			if err != nil {
				return err
			}
			if current == nil {
				return nil
			}
			if err := current.MarkUploadExpired(); err != nil {
				return nil
			}
			if err := h.repo.Update(txCtx, current); err != nil {
				return err
			}
			return h.outbox.StoreEvents(txCtx, current.PullEvents())
		})
		if err != nil {
			log.Printf("failed to expire media file %s: %v", media.ID, err)
		}
	}
	return nil
}
