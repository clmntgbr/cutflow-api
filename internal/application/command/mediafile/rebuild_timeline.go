package mediafile

import (
	"context"
	"log"
	"time"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/event"
	domainjob "go-api/internal/domain/job"
	domainmediaconfig "go-api/internal/domain/mediaconfig"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"
	domainsilence "go-api/internal/domain/silence"
	domaintimeline "go-api/internal/domain/timeline"
	domaintranscriptissue "go-api/internal/domain/transcriptissue"

	"github.com/google/uuid"
)

type RebuildTimelineCommand struct {
	MediaFileID uuid.UUID
	ProjectID   uuid.UUID
	UserID      uuid.UUID
	JobID       uuid.UUID
	Reason      string
}

type UserOverrideLister interface {
	ListByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) ([]domaintimeline.Override, error)
}

type RebuildTimelineHandler struct {
	mediaRepo     domainmediafile.MediaFileWriteRepository
	configRepo    domainmediaconfig.MediaConfigurationWriteRepository
	silenceRepo   domainsilence.DetectedSilenceWriteRepository
	issueRepo     domaintranscriptissue.IssueWriteRepository
	overrideRepo  UserOverrideLister
	timelineRepo  domaintimeline.TimelineWriteRepository
	jobRepo       domainjob.JobWriteRepository
	outbox        port.OutboxRepository
}

func NewRebuildTimelineHandler(
	mediaRepo domainmediafile.MediaFileWriteRepository,
	configRepo domainmediaconfig.MediaConfigurationWriteRepository,
	silenceRepo domainsilence.DetectedSilenceWriteRepository,
	issueRepo domaintranscriptissue.IssueWriteRepository,
	overrideRepo UserOverrideLister,
	timelineRepo domaintimeline.TimelineWriteRepository,
	jobRepo domainjob.JobWriteRepository,
	outbox port.OutboxRepository,
) *RebuildTimelineHandler {
	return &RebuildTimelineHandler{
		mediaRepo:    mediaRepo,
		configRepo:   configRepo,
		silenceRepo:  silenceRepo,
		issueRepo:    issueRepo,
		overrideRepo: overrideRepo,
		timelineRepo: timelineRepo,
		jobRepo:      jobRepo,
		outbox:       outbox,
	}
}

func (h *RebuildTimelineHandler) Handle(ctx context.Context, cmd RebuildTimelineCommand) error {
	if err := h.markProcessing(ctx, cmd.JobID); err != nil {
		return err
	}

	media, err := h.mediaRepo.GetByID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if media == nil || media.DurationMs <= 0 {
		if failErr := h.fail(ctx, cmd, "media not ready"); failErr != nil {
			return failErr
		}
		return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
	}

	cfg, err := h.configRepo.GetByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if cfg == nil {
		cfg = domainmediaconfig.NewDefault(cmd.MediaFileID)
	}

	silences, err := h.silenceRepo.ListByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	issues, err := h.issueRepo.ListByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	overrides, err := h.overrideRepo.ListByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}

	silenceIntervals := domainsilence.IntervalsFromDetected(silences)
	fp := domaintimeline.ComputeFingerprint(domaintimeline.FingerprintInput{
		MediaFileID:          cmd.MediaFileID,
		MediaDurationMs:      media.DurationMs,
		SilenceRemoval:       cfg.SilenceRemovalEnabled,
		FillerRemoval:        cfg.FillerRemovalEnabled,
		RepetitionRemoval:    cfg.RepetitionRemovalEnabled,
		SilenceMinDurationMs: cfg.SilenceMinDurationMs,
		SilencePadBeforeMs:   cfg.SilencePaddingBeforeMs,
		SilencePadAfterMs:    cfg.SilencePaddingAfterMs,
		SpeechMinDurationMs:  cfg.SpeechMinDurationMs,
		Silences:             silenceIntervals,
		Issues:               issues,
		Overrides:            overrides,
		EngineVersion:        domaintimeline.EngineVersion,
	})

	existing, err := h.timelineRepo.FindByFingerprint(ctx, cmd.MediaFileID, fp)
	if err != nil {
		return messaging.Retryable(err)
	}
	if existing != nil {
		log.Printf("timeline rebuild reuse mediaFileId=%s version=%d fingerprint=%s reason=%s",
			cmd.MediaFileID, existing.Version, fp[:12], cmd.Reason)
		return h.reuse(ctx, cmd, existing)
	}

	decisions := domaintimeline.BuildDecisions(domaintimeline.BuildInput{
		MediaFileID:          cmd.MediaFileID,
		MediaDurationMs:      media.DurationMs,
		SilenceRemoval:       cfg.SilenceRemovalEnabled,
		FillerRemoval:        cfg.FillerRemovalEnabled,
		RepetitionRemoval:    cfg.RepetitionRemovalEnabled,
		SilenceMinDurationMs: cfg.SilenceMinDurationMs,
		SilencePadBeforeMs:   cfg.SilencePaddingBeforeMs,
		SilencePadAfterMs:    cfg.SilencePaddingAfterMs,
		SpeechMinDurationMs:  cfg.SpeechMinDurationMs,
		Silences:             silenceIntervals,
		Issues:               issues,
		Overrides:            overrides,
	})

	tl, err := domaintimeline.BuildTimeline(
		cmd.ProjectID,
		cmd.MediaFileID,
		media.DurationMs,
		decisions,
		cfg.SpeechMinDurationMs,
	)
	if err != nil {
		log.Printf("timeline build failed mediaFileId=%s: %v", cmd.MediaFileID, err)
		if failErr := h.fail(ctx, cmd, "timeline build failed"); failErr != nil {
			return failErr
		}
		return messaging.NonRetryable(err)
	}

	version, err := h.timelineRepo.NextVersion(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	tl.Version = version
	tl.Fingerprint = fp
	tl.IsActive = true

	log.Printf(
		"timeline rebuilt mediaFileId=%s version=%d durationMs=%d segments=%d decisions=%d reason=%s",
		cmd.MediaFileID, tl.Version, tl.DurationMs, len(tl.Segments), len(tl.Decisions), cmd.Reason,
	)

	return h.persist(ctx, cmd, tl)
}

func (h *RebuildTimelineHandler) reuse(ctx context.Context, cmd RebuildTimelineCommand, tl *domaintimeline.Timeline) error {
	return h.timelineRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		return h.reuseInTx(txCtx, cmd, tl)
	})
}

func (h *RebuildTimelineHandler) persist(ctx context.Context, cmd RebuildTimelineCommand, tl *domaintimeline.Timeline) error {
	return h.timelineRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Race: another worker may have saved same fingerprint.
		existing, err := h.timelineRepo.FindByFingerprint(txCtx, cmd.MediaFileID, tl.Fingerprint)
		if err != nil {
			return messaging.Retryable(err)
		}
		if existing != nil {
			return h.reuseInTx(txCtx, cmd, existing)
		}

		if err := h.timelineRepo.DeactivateOthers(txCtx, cmd.MediaFileID, tl.ID); err != nil {
			return messaging.Retryable(err)
		}
		if err := h.timelineRepo.Save(txCtx, tl); err != nil {
			return messaging.Retryable(err)
		}

		events := []event.DomainEvent{
			domaintimeline.Updated{
				ID:          uuid.New().String(),
				TimelineID:  tl.ID.String(),
				MediaFileID: cmd.MediaFileID.String(),
				ProjectID:   cmd.ProjectID.String(),
				UserID:      cmd.UserID.String(),
				Version:     tl.Version,
				DurationMs:  tl.DurationMs,
				Fingerprint: tl.Fingerprint,
				Timestamp:   time.Now().UTC(),
			},
		}
		if err := h.succeedJobInTx(txCtx, cmd.JobID, &events); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, events)
	})
}

func (h *RebuildTimelineHandler) reuseInTx(ctx context.Context, cmd RebuildTimelineCommand, tl *domaintimeline.Timeline) error {
	if err := h.timelineRepo.DeactivateOthers(ctx, cmd.MediaFileID, tl.ID); err != nil {
		return messaging.Retryable(err)
	}
	if err := h.timelineRepo.Activate(ctx, tl.ID); err != nil {
		return messaging.Retryable(err)
	}
	events := []event.DomainEvent{
		domaintimeline.Updated{
			ID:          uuid.New().String(),
			TimelineID:  tl.ID.String(),
			MediaFileID: cmd.MediaFileID.String(),
			ProjectID:   cmd.ProjectID.String(),
			UserID:      cmd.UserID.String(),
			Version:     tl.Version,
			DurationMs:  tl.DurationMs,
			Fingerprint: tl.Fingerprint,
			Timestamp:   time.Now().UTC(),
		},
	}
	if err := h.succeedJobInTx(ctx, cmd.JobID, &events); err != nil {
		return err
	}
	return h.outbox.StoreEvents(ctx, events)
}

func (h *RebuildTimelineHandler) succeedJobInTx(ctx context.Context, jobID uuid.UUID, events *[]event.DomainEvent) error {
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
	*events = append(*events, job.PullEvents()...)
	return nil
}

func (h *RebuildTimelineHandler) markProcessing(ctx context.Context, jobID uuid.UUID) error {
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

func (h *RebuildTimelineHandler) fail(ctx context.Context, cmd RebuildTimelineCommand, reason string) error {
	return h.jobRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		events := []event.DomainEvent{
			domaintimeline.Failed{
				ID:          uuid.New().String(),
				MediaFileID: cmd.MediaFileID.String(),
				ProjectID:   cmd.ProjectID.String(),
				UserID:      cmd.UserID.String(),
				JobID:       cmd.JobID.String(),
				Reason:      reason,
				Timestamp:   time.Now().UTC(),
			},
		}
		if cmd.JobID != uuid.Nil {
			job, err := h.jobRepo.GetByID(txCtx, cmd.JobID)
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
		}
		return h.outbox.StoreEvents(txCtx, events)
	})
}
