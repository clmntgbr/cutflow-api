package mediafile

import (
	"context"
	"io"
	"log"
	"os"

	"go-api/internal/application/messaging"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type ProbeMediaCommand struct {
	MediaFileID uuid.UUID
}

type ProbeMediaHandler struct {
	mediaRepo domainmediafile.MediaFileWriteRepository
	storage   port.Storage
	prober    port.MediaProber
	outbox    port.OutboxRepository
}

func NewProbeMediaHandler(
	mediaRepo domainmediafile.MediaFileWriteRepository,
	storage port.Storage,
	prober port.MediaProber,
	outbox port.OutboxRepository,
) *ProbeMediaHandler {
	return &ProbeMediaHandler{
		mediaRepo: mediaRepo,
		storage:   storage,
		prober:    prober,
		outbox:    outbox,
	}
}

func (h *ProbeMediaHandler) Handle(ctx context.Context, cmd ProbeMediaCommand) error {
	media, err := h.mediaRepo.GetByID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if media == nil {
		return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
	}
	switch media.Status {
	case domainmediafile.StatusReady, domainmediafile.StatusProcessing, domainmediafile.StatusCompleted:
		return nil
	case domainmediafile.StatusFailed:
		return nil
	case domainmediafile.StatusUploaded, domainmediafile.StatusProbing:
	default:
		return messaging.NonRetryable(domainmediafile.ErrInvalidTransition)
	}

	if err := h.markProbing(ctx, cmd.MediaFileID); err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", "media-probe-src-*")
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

	result, err := h.prober.Probe(ctx, tmp.Name())
	if err != nil {
		log.Printf("ffprobe failed mediaFileId=%s: %v", media.ID, err)
		if failErr := h.failProbe(ctx, cmd.MediaFileID, domainmediafile.ErrProbeInvalid.Error()); failErr != nil {
			return failErr
		}
		return messaging.NonRetryable(domainmediafile.ErrProbeInvalid)
	}

	meta := domainmediafile.ProbeMetadata{
		DurationMs:      result.DurationMs,
		Width:           result.Width,
		Height:          result.Height,
		FPS:             result.FPS,
		VideoCodec:      result.VideoCodec,
		AudioCodec:      result.AudioCodec,
		AudioSampleRate: result.AudioSampleRate,
		AudioChannels:   result.AudioChannels,
		SizeBytes:       result.SizeBytes,
	}
	if err := domainmediafile.ValidateProbeMetadata(meta); err != nil {
		if failErr := h.failProbe(ctx, cmd.MediaFileID, err.Error()); failErr != nil {
			return failErr
		}
		return messaging.NonRetryable(err)
	}

	return h.persistReady(ctx, cmd.MediaFileID, meta)
}

func (h *ProbeMediaHandler) markProbing(ctx context.Context, mediaFileID uuid.UUID) error {
	return h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		media, err := h.mediaRepo.GetByID(txCtx, mediaFileID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if media == nil {
			return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
		}
		switch media.Status {
		case domainmediafile.StatusProbing, domainmediafile.StatusReady, domainmediafile.StatusProcessing, domainmediafile.StatusCompleted:
			return nil
		case domainmediafile.StatusFailed:
			return messaging.NonRetryable(domainmediafile.ErrInvalidTransition)
		}

		if err := media.MarkProbing(); err != nil {
			return messaging.NonRetryable(err)
		}
		if err := h.mediaRepo.Update(txCtx, media); err != nil {
			return messaging.Retryable(err)
		}
		return h.outbox.StoreEvents(txCtx, media.PullEvents())
	})
}

func (h *ProbeMediaHandler) persistReady(
	ctx context.Context,
	mediaFileID uuid.UUID,
	meta domainmediafile.ProbeMetadata,
) error {
	return h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		media, err := h.mediaRepo.GetByID(txCtx, mediaFileID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if media == nil {
			return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
		}
		if media.Status == domainmediafile.StatusReady {
			return nil
		}
		if err := media.ApplyProbe(meta); err != nil {
			return messaging.NonRetryable(err)
		}
		if err := h.mediaRepo.Update(txCtx, media); err != nil {
			return messaging.Retryable(err)
		}
		return h.outbox.StoreEvents(txCtx, media.PullEvents())
	})
}

func (h *ProbeMediaHandler) failProbe(ctx context.Context, mediaFileID uuid.UUID, reason string) error {
	return h.mediaRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		media, err := h.mediaRepo.GetByID(txCtx, mediaFileID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if media == nil {
			return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
		}
		if media.Status == domainmediafile.StatusFailed {
			return nil
		}
		if err := media.MarkProbeFailed(reason); err != nil {
			return messaging.NonRetryable(err)
		}
		if err := h.mediaRepo.Update(txCtx, media); err != nil {
			return messaging.Retryable(err)
		}
		return h.outbox.StoreEvents(txCtx, media.PullEvents())
	})
}
