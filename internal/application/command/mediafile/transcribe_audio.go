package mediafile

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"

	"go-api/internal/application/messaging"
	domainjob "go-api/internal/domain/job"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"
	domaintranscript "go-api/internal/domain/transcript"

	"github.com/google/uuid"
)

type TranscribeAudioCommand struct {
	MediaFileID uuid.UUID
	AudioKey    string
	ProjectID   uuid.UUID
	UserID      uuid.UUID
	JobID       uuid.UUID
}

type TranscribeAudioHandler struct {
	transcriptRepo domaintranscript.TranscriptWriteRepository
	jobRepo        domainjob.JobWriteRepository
	storage        port.Storage
	transcriber    port.SpeechTranscriber
	outbox         port.OutboxRepository
}

func NewTranscribeAudioHandler(
	transcriptRepo domaintranscript.TranscriptWriteRepository,
	jobRepo domainjob.JobWriteRepository,
	storage port.Storage,
	transcriber port.SpeechTranscriber,
	outbox port.OutboxRepository,
) *TranscribeAudioHandler {
	return &TranscribeAudioHandler{
		transcriptRepo: transcriptRepo,
		jobRepo:        jobRepo,
		storage:        storage,
		transcriber:    transcriber,
		outbox:         outbox,
	}
}

func (h *TranscribeAudioHandler) Handle(ctx context.Context, cmd TranscribeAudioCommand) error {
	if err := h.markProcessing(ctx, cmd.JobID); err != nil {
		return err
	}

	existing, err := h.transcriptRepo.GetByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if existing != nil && existing.Status == domaintranscript.StatusCompleted {
		return h.succeedJob(ctx, cmd.JobID)
	}

	transcript := existing
	if transcript == nil {
		transcript = domaintranscript.NewPending(cmd.MediaFileID, cmd.ProjectID, cmd.UserID)
		if err := h.transcriptRepo.Save(ctx, transcript); err != nil {
			return messaging.Retryable(err)
		}
	}

	tmp, err := os.CreateTemp("", "media-transcript-*.opus")
	if err != nil {
		return messaging.Retryable(err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	reader, err := h.storage.Get(ctx, cmd.AudioKey)
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

	result, err := h.transcriber.Transcribe(ctx, tmp.Name())
	if err != nil {
		log.Printf("transcription failed mediaFileId=%s: %v", cmd.MediaFileID, err)
		if failErr := h.fail(ctx, cmd.MediaFileID, cmd.JobID, "transcription failed"); failErr != nil {
			return failErr
		}
		return messaging.NonRetryable(err)
	}

	ass := result.ASS
	if err := h.storage.Put(ctx, transcript.SRTStorageKey, bytes.NewReader([]byte(result.SRT)), int64(len(result.SRT)), "application/x-subrip"); err != nil {
		return messaging.Retryable(err)
	}
	if err := h.storage.Put(ctx, transcript.ASSStorageKey, bytes.NewReader([]byte(ass)), int64(len(ass)), "text/x-ass"); err != nil {
		return messaging.Retryable(err)
	}

	words := make([]domaintranscript.Word, 0, len(result.Words))
	for i, word := range result.Words {
		words = append(words, domaintranscript.Word{
			WordIndex:     i,
			Text:          word.Text,
			SourceStartMs: word.StartMs,
			SourceEndMs:   word.EndMs,
			Confidence:    word.Confidence,
			Kind:          domaintranscript.KindSpeech,
		})
	}

	return h.complete(ctx, cmd.MediaFileID, cmd.JobID, result.ProviderJobID, result.Language, result.Text, words)
}

func (h *TranscribeAudioHandler) markProcessing(ctx context.Context, jobID uuid.UUID) error {
	if jobID == uuid.Nil {
		return nil
	}
	return h.jobRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		job, err := h.jobRepo.GetByID(txCtx, jobID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if job == nil || job.Status == domainjob.StatusSuccess || job.Status == domainjob.StatusProcessing {
			return nil
		}
		job.MarkProcessing()
		if err := h.jobRepo.Update(txCtx, job); err != nil {
			return messaging.Retryable(err)
		}
		return h.outbox.StoreEvents(txCtx, job.PullEvents())
	})
}

func (h *TranscribeAudioHandler) succeedJob(ctx context.Context, jobID uuid.UUID) error {
	if jobID == uuid.Nil {
		return nil
	}
	return h.jobRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		job, err := h.jobRepo.GetByID(txCtx, jobID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if job == nil || job.Status == domainjob.StatusSuccess {
			return nil
		}
		job.MarkSuccess()
		if err := h.jobRepo.Update(txCtx, job); err != nil {
			return messaging.Retryable(err)
		}
		return h.outbox.StoreEvents(txCtx, job.PullEvents())
	})
}

func (h *TranscribeAudioHandler) complete(
	ctx context.Context,
	mediaFileID, jobID uuid.UUID,
	providerJobID, language, text string,
	words []domaintranscript.Word,
) error {
	return h.transcriptRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		transcript, err := h.transcriptRepo.GetByMediaFileID(txCtx, mediaFileID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if transcript == nil {
			return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
		}
		if transcript.Status == domaintranscript.StatusCompleted {
			return h.succeedJobInTx(txCtx, jobID)
		}
		if providerJobID != "" {
			transcript.MarkProcessing(providerJobID)
		}
		transcript.MarkCompleted(language, text, transcript.SRTStorageKey, transcript.ASSStorageKey, words)
		if err := h.transcriptRepo.Update(txCtx, transcript); err != nil {
			return messaging.Retryable(err)
		}
		if err := h.transcriptRepo.ReplaceWords(txCtx, transcript.ID, words); err != nil {
			return messaging.Retryable(err)
		}
		events := transcript.PullEvents()
		if jobID != uuid.Nil {
			job, err := h.jobRepo.GetByID(txCtx, jobID)
			if err != nil {
				return messaging.Retryable(err)
			}
			if job != nil && job.Status != domainjob.StatusSuccess {
				job.MarkSuccess()
				if err := h.jobRepo.Update(txCtx, job); err != nil {
					return messaging.Retryable(err)
				}
				events = append(events, job.PullEvents()...)
			}
		}
		return h.outbox.StoreEvents(txCtx, events)
	})
}

func (h *TranscribeAudioHandler) succeedJobInTx(ctx context.Context, jobID uuid.UUID) error {
	if jobID == uuid.Nil {
		return nil
	}
	job, err := h.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if job == nil || job.Status == domainjob.StatusSuccess {
		return nil
	}
	job.MarkSuccess()
	if err := h.jobRepo.Update(ctx, job); err != nil {
		return messaging.Retryable(err)
	}
	return h.outbox.StoreEvents(ctx, job.PullEvents())
}

func (h *TranscribeAudioHandler) fail(ctx context.Context, mediaFileID, jobID uuid.UUID, reason string) error {
	return h.transcriptRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		transcript, err := h.transcriptRepo.GetByMediaFileID(txCtx, mediaFileID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if transcript == nil {
			return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
		}

		if transcript.Status != domaintranscript.StatusFailed {
			transcript.MarkFailed(reason)
			if err := h.transcriptRepo.Update(txCtx, transcript); err != nil {
				return messaging.Retryable(err)
			}
		}
		domainEvents := transcript.PullEvents()
		if jobID != uuid.Nil {
			job, err := h.jobRepo.GetByID(txCtx, jobID)
			if err != nil {
				return messaging.Retryable(err)
			}
			if job != nil && job.Status != domainjob.StatusFailed {
				job.MarkFailed(reason)
				if err := h.jobRepo.Update(txCtx, job); err != nil {
					return messaging.Retryable(err)
				}
				domainEvents = append(domainEvents, job.PullEvents()...)
			}
		}
		return h.outbox.StoreEvents(txCtx, domainEvents)
	})
}
