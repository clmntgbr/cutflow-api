package mediafile

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/event"
	domainjob "go-api/internal/domain/job"
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
	jobRepo   domainjob.JobWriteRepository
	storage   port.Storage
	extractor port.AudioExtractor
	outbox    port.OutboxRepository
}

func NewExtractAudioHandler(
	mediaRepo domainmediafile.MediaFileWriteRepository,
	audioRepo domainmediaaudio.MediaAudioWriteRepository,
	jobRepo domainjob.JobWriteRepository,
	storage port.Storage,
	extractor port.AudioExtractor,
	outbox port.OutboxRepository,
) *ExtractAudioHandler {
	return &ExtractAudioHandler{
		mediaRepo: mediaRepo,
		audioRepo: audioRepo,
		jobRepo:   jobRepo,
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

	job := domainjob.New(media.ProjectID, media.ID, media.UserID, domainjob.NameExtractAudio)
	job.MarkProcessing()
	if err := h.persistNewJob(ctx, job); err != nil {
		return err
	}

	audio := existing
	if audio == nil {
		audio = domainmediaaudio.NewPending(media.ID, media.ProjectID, media.UserID)
		if err := h.audioRepo.Save(ctx, audio); err != nil {
			_ = h.failJob(ctx, job.ID, "audio extraction failed")
			return messaging.Retryable(err)
		}
	}

	audio.MarkProcessing()
	if err := h.audioRepo.Update(ctx, audio); err != nil {
		_ = h.failJob(ctx, job.ID, "audio extraction failed")
		return messaging.Retryable(err)
	}

	src, err := os.CreateTemp("", "media-audio-src-*")
	if err != nil {
		_ = h.failJob(ctx, job.ID, "audio extraction failed")
		return messaging.Retryable(err)
	}
	defer os.Remove(src.Name())
	defer src.Close()

	out, err := os.CreateTemp("", "media-audio-*.opus")
	if err != nil {
		_ = h.failJob(ctx, job.ID, "audio extraction failed")
		return messaging.Retryable(err)
	}
	outPath := out.Name()
	_ = out.Close()
	defer os.Remove(outPath)

	reader, err := h.storage.Get(ctx, media.StorageKey)
	if err != nil {
		_ = h.failJob(ctx, job.ID, "audio extraction failed")
		return messaging.Retryable(err)
	}
	if _, err := io.Copy(src, reader); err != nil {
		_ = reader.Close()
		_ = h.failJob(ctx, job.ID, "audio extraction failed")
		return messaging.Retryable(err)
	}
	_ = reader.Close()
	if err := src.Close(); err != nil {
		_ = h.failJob(ctx, job.ID, "audio extraction failed")
		return messaging.Retryable(err)
	}

	if err := h.extractor.Extract(ctx, src.Name(), outPath); err != nil {
		log.Printf("extract audio failed mediaFileId=%s: %v", media.ID, err)
		if failErr := h.fail(ctx, media.ID, job.ID, "audio extraction failed"); failErr != nil {
			return failErr
		}
		return messaging.NonRetryable(err)
	}

	info, err := os.Stat(outPath)
	if err != nil {
		_ = h.failJob(ctx, job.ID, "audio extraction failed")
		return messaging.Retryable(err)
	}
	file, err := os.Open(outPath)
	if err != nil {
		_ = h.failJob(ctx, job.ID, "audio extraction failed")
		return messaging.Retryable(err)
	}
	defer file.Close()

	if err := h.storage.Put(ctx, audio.StorageKey, file, info.Size(), domainmediaaudio.ExtractedContentType); err != nil {
		_ = h.failJob(ctx, job.ID, "audio extraction failed")
		return messaging.Retryable(err)
	}

	return h.complete(ctx, media.ID, job.ID, info.Size())
}

func (h *ExtractAudioHandler) persistNewJob(ctx context.Context, job *domainjob.Job) error {
	return h.jobRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.jobRepo.Save(txCtx, job); err != nil {
			return messaging.Retryable(err)
		}
		return h.outbox.StoreEvents(txCtx, job.PullEvents())
	})
}

func (h *ExtractAudioHandler) complete(ctx context.Context, mediaFileID, jobID uuid.UUID, sizeBytes int64) error {
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

		job, err := h.jobRepo.GetByID(txCtx, jobID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if job == nil {
			return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
		}
		job.MarkSuccess()
		if err := h.jobRepo.Update(txCtx, job); err != nil {
			return messaging.Retryable(err)
		}

		silenceJob := domainjob.New(audio.ProjectID, audio.MediaFileID, audio.UserID, domainjob.NameDetectSilence)
		transcriptJob := domainjob.New(audio.ProjectID, audio.MediaFileID, audio.UserID, domainjob.NameTranscribeAudio)
		if err := h.jobRepo.Save(txCtx, silenceJob); err != nil {
			return messaging.Retryable(err)
		}
		if err := h.jobRepo.Save(txCtx, transcriptJob); err != nil {
			return messaging.Retryable(err)
		}

		now := time.Now().UTC()
		events := make([]event.DomainEvent, 0, 8)
		events = append(events, audio.PullEvents()...)
		events = append(events, job.PullEvents()...)
		events = append(events, silenceJob.PullEvents()...)
		events = append(events, transcriptJob.PullEvents()...)
		events = append(events,
			domainmediafile.MediaFileSilenceRequested{
				ID:          uuid.New().String(),
				MediaFileID: audio.MediaFileID.String(),
				ProjectID:   audio.ProjectID.String(),
				UserID:      audio.UserID.String(),
				JobID:       silenceJob.ID.String(),
				AudioKey:    audio.StorageKey,
				Timestamp:   now,
			},
			domainmediafile.MediaFileTranscriptRequested{
				ID:          uuid.New().String(),
				MediaFileID: audio.MediaFileID.String(),
				ProjectID:   audio.ProjectID.String(),
				UserID:      audio.UserID.String(),
				JobID:       transcriptJob.ID.String(),
				AudioKey:    audio.StorageKey,
				Timestamp:   now,
			},
		)
		return h.outbox.StoreEvents(txCtx, events)
	})
}

func (h *ExtractAudioHandler) fail(ctx context.Context, mediaFileID, jobID uuid.UUID, reason string) error {
	return h.audioRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		audio, err := h.audioRepo.GetByMediaFileID(txCtx, mediaFileID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if audio == nil {
			return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
		}
		if audio.Status != domainmediaaudio.StatusFailed {
			audio.MarkFailed(reason)
			if err := h.audioRepo.Update(txCtx, audio); err != nil {
				return messaging.Retryable(err)
			}
		}

		events := audio.PullEvents()
		job, err := h.jobRepo.GetByID(txCtx, jobID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if job != nil && job.Status != domainjob.StatusFailed {
			job.MarkFailed(reason)
			if err := h.jobRepo.Update(txCtx, job); err != nil {
				return messaging.Retryable(err)
			}
			events = append(events, job.PullEvents()...)
		}
		return h.outbox.StoreEvents(txCtx, events)
	})
}

func (h *ExtractAudioHandler) failJob(ctx context.Context, jobID uuid.UUID, reason string) error {
	return h.jobRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		job, err := h.jobRepo.GetByID(txCtx, jobID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if job == nil || job.Status == domainjob.StatusFailed {
			return nil
		}
		job.MarkFailed(reason)
		if err := h.jobRepo.Update(txCtx, job); err != nil {
			return messaging.Retryable(err)
		}
		return h.outbox.StoreEvents(txCtx, job.PullEvents())
	})
}
