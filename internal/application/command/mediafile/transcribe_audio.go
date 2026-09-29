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
	domaintranscript "go-api/internal/domain/transcript"

	"github.com/google/uuid"
)

type TranscribeAudioCommand struct {
	MediaFileID uuid.UUID
	AudioKey    string
	ProjectID   uuid.UUID
	UserID      uuid.UUID
}

type TranscribeAudioHandler struct {
	transcriptRepo domaintranscript.TranscriptWriteRepository
	storage        port.Storage
	transcriber    port.SpeechTranscriber
	outbox         port.OutboxRepository
}

func NewTranscribeAudioHandler(
	transcriptRepo domaintranscript.TranscriptWriteRepository,
	storage port.Storage,
	transcriber port.SpeechTranscriber,
	outbox port.OutboxRepository,
) *TranscribeAudioHandler {
	return &TranscribeAudioHandler{
		transcriptRepo: transcriptRepo,
		storage:        storage,
		transcriber:    transcriber,
		outbox:         outbox,
	}
}

func (h *TranscribeAudioHandler) Handle(ctx context.Context, cmd TranscribeAudioCommand) error {
	existing, err := h.transcriptRepo.GetByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if existing != nil && existing.Status == domaintranscript.StatusCompleted {
		return nil
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
		if failErr := h.fail(ctx, cmd.MediaFileID, "transcription failed"); failErr != nil {
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

	return h.complete(ctx, cmd.MediaFileID, result.ProviderJobID, result.Language, result.Text, words)
}

func (h *TranscribeAudioHandler) complete(
	ctx context.Context,
	mediaFileID uuid.UUID,
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
			return nil
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
		return h.outbox.StoreEvents(txCtx, transcript.PullEvents())
	})
}

func (h *TranscribeAudioHandler) fail(ctx context.Context, mediaFileID uuid.UUID, reason string) error {
	return h.transcriptRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		transcript, err := h.transcriptRepo.GetByMediaFileID(txCtx, mediaFileID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if transcript == nil {
			return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
		}
		if transcript.Status == domaintranscript.StatusFailed {
			return nil
		}
		transcript.MarkFailed(reason)
		if err := h.transcriptRepo.Update(txCtx, transcript); err != nil {
			return messaging.Retryable(err)
		}
		return h.outbox.StoreEvents(txCtx, transcript.PullEvents())
	})
}
