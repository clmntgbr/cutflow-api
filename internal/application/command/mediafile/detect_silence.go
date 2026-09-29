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
	"go-api/internal/domain/port"
	domainsilence "go-api/internal/domain/silence"

	"github.com/google/uuid"
)

type DetectSilenceCommand struct {
	MediaFileID uuid.UUID
	AudioKey    string
	ProjectID   uuid.UUID
	UserID      uuid.UUID
	JobID       uuid.UUID
}

type DetectSilenceHandler struct {
	silenceRepo   domainsilence.DetectedSilenceWriteRepository
	jobRepo       domainjob.JobWriteRepository
	storage       port.Storage
	detector      port.SilenceDetector
	outbox        port.OutboxRepository
	thresholdDB   float64
	minDurationMs int64
}

func NewDetectSilenceHandler(
	silenceRepo domainsilence.DetectedSilenceWriteRepository,
	jobRepo domainjob.JobWriteRepository,
	storage port.Storage,
	detector port.SilenceDetector,
	outbox port.OutboxRepository,
	thresholdDB float64,
	minDurationMs int64,
) *DetectSilenceHandler {
	return &DetectSilenceHandler{
		silenceRepo:   silenceRepo,
		jobRepo:       jobRepo,
		storage:       storage,
		detector:      detector,
		outbox:        outbox,
		thresholdDB:   thresholdDB,
		minDurationMs: minDurationMs,
	}
}

func (h *DetectSilenceHandler) Handle(ctx context.Context, cmd DetectSilenceCommand) error {
	if err := h.markProcessing(ctx, cmd.JobID); err != nil {
		return err
	}

	existing, err := h.silenceRepo.CountByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if existing > 0 {
		return h.markJobSuccess(ctx, cmd.JobID)
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
		if failErr := h.fail(ctx, cmd.JobID, "silence detection failed"); failErr != nil {
			return failErr
		}
		return messaging.NonRetryable(err)
	}

	rows := make([]*domainsilence.DetectedSilence, 0, len(intervals))
	for _, interval := range intervals {
		rows = append(rows, domainsilence.NewDetectedSilence(cmd.MediaFileID, interval.StartMs, interval.EndMs))
	}

	return h.silenceRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		count, err := h.silenceRepo.CountByMediaFileID(txCtx, cmd.MediaFileID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if count > 0 {
			return h.markJobSuccessInTx(txCtx, cmd.JobID)
		}
		if err := h.silenceRepo.ReplaceForMediaFile(txCtx, cmd.MediaFileID, rows); err != nil {
			return messaging.Retryable(err)
		}
		log.Printf("silence detected mediaFileId=%s count=%d", cmd.MediaFileID, len(rows))

		events := []event.DomainEvent{
			domainsilence.SilenceDetected{
				ID:           uuid.New().String(),
				MediaFileID:  cmd.MediaFileID.String(),
				ProjectID:    cmd.ProjectID.String(),
				UserID:       cmd.UserID.String(),
				SilenceCount: len(rows),
				Timestamp:    time.Now().UTC(),
			},
		}
		if cmd.JobID != uuid.Nil {
			job, err := h.jobRepo.GetByID(txCtx, cmd.JobID)
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

func (h *DetectSilenceHandler) markProcessing(ctx context.Context, jobID uuid.UUID) error {
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

func (h *DetectSilenceHandler) markJobSuccess(ctx context.Context, jobID uuid.UUID) error {
	return h.jobRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		return h.markJobSuccessInTx(txCtx, jobID)
	})
}

func (h *DetectSilenceHandler) markJobSuccessInTx(ctx context.Context, jobID uuid.UUID) error {
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

func (h *DetectSilenceHandler) fail(ctx context.Context, jobID uuid.UUID, reason string) error {
	if jobID == uuid.Nil {
		return nil
	}
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
