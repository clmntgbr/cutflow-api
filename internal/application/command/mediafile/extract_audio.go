package mediafile

import (
	"context"
	"io"
	"log"
	"os"

	"go-api/internal/application/messaging"
	domainmediaaudio "go-api/internal/domain/mediaaudio"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type ExtractAudioCommand struct {
	MediaFileID uuid.UUID
}

type ExtractAudioHandler struct {
	mediaRepo domainmediafile.MediaFileWriteRepository
	audioRepo domainmediaaudio.MediaAudioWriteRepository
	storage   port.Storage
	extractor port.AudioExtractor
	outbox    port.OutboxRepository
}

func NewExtractAudioHandler(
	mediaRepo domainmediafile.MediaFileWriteRepository,
	audioRepo domainmediaaudio.MediaAudioWriteRepository,
	storage port.Storage,
	extractor port.AudioExtractor,
	outbox port.OutboxRepository,
) *ExtractAudioHandler {
	return &ExtractAudioHandler{
		mediaRepo: mediaRepo,
		audioRepo: audioRepo,
		storage:   storage,
		extractor: extractor,
		outbox:    outbox,
	}
}

func (h *ExtractAudioHandler) Handle(ctx context.Context, cmd ExtractAudioCommand) error {
	media, err := h.mediaRepo.GetByID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if media == nil {
		return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
	}
	if media.AudioCodec == "" {
		log.Printf("extract audio skipped mediaFileId=%s: no audio track", media.ID)
		return nil
	}

	existing, err := h.audioRepo.GetByMediaFileID(ctx, media.ID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if existing != nil && existing.Status == domainmediaaudio.StatusCompleted {
		return nil
	}

	audio := existing
	if audio == nil {
		audio = domainmediaaudio.NewPending(media.ID, media.ProjectID, media.UserID)
		if err := h.audioRepo.Save(ctx, audio); err != nil {
			return messaging.Retryable(err)
		}
	}

	audio.MarkProcessing()
	if err := h.audioRepo.Update(ctx, audio); err != nil {
		return messaging.Retryable(err)
	}

	src, err := os.CreateTemp("", "media-audio-src-*")
	if err != nil {
		return messaging.Retryable(err)
	}
	defer os.Remove(src.Name())
	defer src.Close()

	out, err := os.CreateTemp("", "media-audio-*.wav")
	if err != nil {
		return messaging.Retryable(err)
	}
	outPath := out.Name()
	_ = out.Close()
	defer os.Remove(outPath)

	reader, err := h.storage.Get(ctx, media.StorageKey)
	if err != nil {
		return messaging.Retryable(err)
	}
	if _, err := io.Copy(src, reader); err != nil {
		_ = reader.Close()
		return messaging.Retryable(err)
	}
	_ = reader.Close()
	if err := src.Close(); err != nil {
		return messaging.Retryable(err)
	}

	if err := h.extractor.Extract(ctx, src.Name(), outPath); err != nil {
		log.Printf("extract audio failed mediaFileId=%s: %v", media.ID, err)
		if failErr := h.fail(ctx, audio.ID, media.ID, "audio extraction failed"); failErr != nil {
			return failErr
		}
		return messaging.NonRetryable(err)
	}

	info, err := os.Stat(outPath)
	if err != nil {
		return messaging.Retryable(err)
	}
	file, err := os.Open(outPath)
	if err != nil {
		return messaging.Retryable(err)
	}
	defer file.Close()

	if err := h.storage.Put(ctx, audio.StorageKey, file, info.Size(), "audio/wav"); err != nil {
		return messaging.Retryable(err)
	}

	return h.complete(ctx, media.ID, info.Size())
}

func (h *ExtractAudioHandler) complete(ctx context.Context, mediaFileID uuid.UUID, sizeBytes int64) error {
	return h.audioRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		audio, err := h.audioRepo.GetByMediaFileID(txCtx, mediaFileID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if audio == nil {
			return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
		}
		if audio.Status == domainmediaaudio.StatusCompleted {
			return nil
		}
		audio.MarkCompleted(sizeBytes)
		if err := h.audioRepo.Update(txCtx, audio); err != nil {
			return messaging.Retryable(err)
		}
		return h.outbox.StoreEvents(txCtx, audio.PullEvents())
	})
}

func (h *ExtractAudioHandler) fail(ctx context.Context, _ uuid.UUID, mediaFileID uuid.UUID, reason string) error {
	return h.audioRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		audio, err := h.audioRepo.GetByMediaFileID(txCtx, mediaFileID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if audio == nil {
			return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
		}
		if audio.Status == domainmediaaudio.StatusFailed {
			return nil
		}
		audio.MarkFailed(reason)
		if err := h.audioRepo.Update(txCtx, audio); err != nil {
			return messaging.Retryable(err)
		}
		return h.outbox.StoreEvents(txCtx, audio.PullEvents())
	})
}
