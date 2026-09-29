package mediafile

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/event"
	domainjob "go-api/internal/domain/job"
	domainmediaconfig "go-api/internal/domain/mediaconfig"
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
	silenceRepo domainsilence.DetectedSilenceWriteRepository
	configRepo  domainmediaconfig.MediaConfigurationWriteRepository
	jobRepo     domainjob.JobWriteRepository
	storage     port.Storage
	detector    port.SilenceDetector
	noiseFloor  port.NoiseFloorAnalyzer
	outbox      port.OutboxRepository
}

func NewDetectSilenceHandler(
	silenceRepo domainsilence.DetectedSilenceWriteRepository,
	configRepo domainmediaconfig.MediaConfigurationWriteRepository,
	jobRepo domainjob.JobWriteRepository,
	storage port.Storage,
	detector port.SilenceDetector,
	noiseFloor port.NoiseFloorAnalyzer,
	outbox port.OutboxRepository,
) *DetectSilenceHandler {
	return &DetectSilenceHandler{
		silenceRepo: silenceRepo,
		configRepo:  configRepo,
		jobRepo:     jobRepo,
		storage:     storage,
		detector:    detector,
		noiseFloor:  noiseFloor,
		outbox:      outbox,
	}
}

func (h *DetectSilenceHandler) Handle(ctx context.Context, cmd DetectSilenceCommand) error {
	if err := h.markProcessing(ctx, cmd.JobID); err != nil {
		return err
	}

	cfg, err := h.configRepo.GetByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if cfg == nil {
		cfg = domainmediaconfig.NewDefault(cmd.MediaFileID)
		if err := h.configRepo.Save(ctx, cfg); err != nil {
			return messaging.Retryable(err)
		}
	}

	if !cfg.SilenceRemovalEnabled {
		log.Printf("silence detection skipped mediaFileId=%s: disabled in configuration", cmd.MediaFileID)
		return h.persistResults(ctx, cmd, nil)
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

	thresholdDB, noiseFloorDB, err := h.resolveThreshold(ctx, tmp.Name(), cfg)
	if err != nil {
		if failErr := h.fail(ctx, cmd.JobID, "silence detection failed"); failErr != nil {
			return failErr
		}
		return messaging.NonRetryable(err)
	}

	// Capture raw silences with an analysis floor. Config min duration + paddings
	// are applied later via silence.ApplyEditFilters (EditDecision / Timeline).
	intervals, err := h.detector.Detect(ctx, tmp.Name(), thresholdDB, domainsilence.AnalysisMinSilenceMs)
	if err != nil {
		if failErr := h.fail(ctx, cmd.JobID, "silence detection failed"); failErr != nil {
			return failErr
		}
		return messaging.NonRetryable(err)
	}

	if noiseFloorDB != nil {
		cfg.SetNoiseFloor(*noiseFloorDB)
	} else {
		cfg.ClearNoiseFloor()
	}
	cfg.SetCalculatedThreshold(thresholdDB)
	if err := h.configRepo.Update(ctx, cfg); err != nil {
		return messaging.Retryable(err)
	}

	noiseFloorLog := "n/a"
	if noiseFloorDB != nil {
		noiseFloorLog = fmt.Sprintf("%.1f", *noiseFloorDB)
	}
	log.Printf(
		"silence detected mediaFileId=%s noiseFloor=%sdB threshold=%.1fdB analysisMinMs=%d count=%d level=%s",
		cmd.MediaFileID, noiseFloorLog, thresholdDB, domainsilence.AnalysisMinSilenceMs, len(intervals), cfg.SilenceDetectionLevel,
	)
	return h.persistResults(ctx, cmd, intervals)
}

func (h *DetectSilenceHandler) resolveThreshold(
	ctx context.Context,
	audioPath string,
	cfg *domainmediaconfig.MediaConfiguration,
) (float64, *float64, error) {
	if cfg.SilenceThresholdMode == domainmediaconfig.ThresholdModeManual && cfg.SilenceThresholdDB != nil {
		return *cfg.SilenceThresholdDB, nil, nil
	}
	noiseFloor, err := h.noiseFloor.Analyze(ctx, audioPath)
	if err != nil {
		// Fallback when analysis fails: assume -45 dBFS floor so threshold stays
		// reproducible, and still persist that assumed floor.
		log.Printf("noise floor analysis failed, using fallback: %v", err)
		fallback := -45.0
		return cfg.ResolveSilenceThreshold(fallback), &fallback, nil
	}
	nf := noiseFloor
	return cfg.ResolveSilenceThreshold(noiseFloor), &nf, nil
}

func (h *DetectSilenceHandler) persistResults(
	ctx context.Context,
	cmd DetectSilenceCommand,
	intervals []port.SilenceInterval,
) error {
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
