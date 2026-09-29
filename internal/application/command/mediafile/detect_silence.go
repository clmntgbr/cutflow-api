package mediafile

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/event"
	"go-api/internal/domain/port"
	domainsilence "go-api/internal/domain/silence"

	"github.com/google/uuid"
)

type DetectSilenceCommand struct {
	MediaFileID uuid.UUID
	AudioKey    string
	ProjectID   uuid.UUID
	UserID      uuid.UUID
}

type DetectSilenceHandler struct {
	silenceRepo   domainsilence.DetectedSilenceWriteRepository
	storage       port.Storage
	detector      port.SilenceDetector
	outbox        port.OutboxRepository
	thresholdDB   float64
	minDurationMs int64
}

func NewDetectSilenceHandler(
	silenceRepo domainsilence.DetectedSilenceWriteRepository,
	storage port.Storage,
	detector port.SilenceDetector,
	outbox port.OutboxRepository,
	thresholdDB float64,
	minDurationMs int64,
) *DetectSilenceHandler {
	return &DetectSilenceHandler{
		silenceRepo:   silenceRepo,
		storage:       storage,
		detector:      detector,
		outbox:        outbox,
		thresholdDB:   thresholdDB,
		minDurationMs: minDurationMs,
	}
}

func (h *DetectSilenceHandler) Handle(ctx context.Context, cmd DetectSilenceCommand) error {
	existing, err := h.silenceRepo.CountByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if existing > 0 {
		return nil
	}

	tmp, err := os.CreateTemp("", "media-silence-*.opus")
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

	intervals, err := h.detector.Detect(ctx, tmp.Name(), h.thresholdDB, h.minDurationMs)
	if err != nil {
		return messaging.Retryable(err)
	}

	rows := make([]*domainsilence.DetectedSilence, 0, len(intervals))
	for _, interval := range intervals {
		rows = append(rows, domainsilence.NewDetectedSilence(cmd.MediaFileID, interval.StartMs, interval.EndMs))
	}

	evt := domainsilence.SilenceDetected{
		ID:           uuid.New().String(),
		MediaFileID:  cmd.MediaFileID.String(),
		ProjectID:    cmd.ProjectID.String(),
		UserID:       cmd.UserID.String(),
		SilenceCount: len(rows),
		Timestamp:    time.Now().UTC(),
	}

	return h.silenceRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		count, err := h.silenceRepo.CountByMediaFileID(txCtx, cmd.MediaFileID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if count > 0 {
			return nil
		}
		if err := h.silenceRepo.ReplaceForMediaFile(txCtx, cmd.MediaFileID, rows); err != nil {
			return messaging.Retryable(err)
		}
		log.Printf("silence detected mediaFileId=%s count=%d", cmd.MediaFileID, len(rows))
		return h.outbox.StoreEvents(txCtx, []event.DomainEvent{evt})
	})
}
