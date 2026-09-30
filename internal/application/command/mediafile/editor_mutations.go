package mediafile

import (
	"context"
	"errors"
	"log"
	"time"

	"go-api/internal/application/messaging"
	querymediafile "go-api/internal/application/query/mediafile"
	"go-api/internal/domain/event"
	domainjob "go-api/internal/domain/job"
	domainmediaaudio "go-api/internal/domain/mediaaudio"
	domainmediaconfig "go-api/internal/domain/mediaconfig"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"
	domainsilence "go-api/internal/domain/silence"
	domaintimeline "go-api/internal/domain/timeline"
	domaintranscriptissue "go-api/internal/domain/transcriptissue"

	"github.com/google/uuid"
)

var (
	ErrStaleTimeline      = errors.New("stale timeline")
	ErrDecisionNotFound   = errors.New("decision not found")
	ErrInvalidManualRange = errors.New("invalid manual cut range")
)

type OverrideStore interface {
	ListByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) ([]domaintimeline.Override, error)
	Save(ctx context.Context, o domaintimeline.Override) error
	DeleteMatching(ctx context.Context, mediaFileID uuid.UUID, startMs, endMs int64, action string) error
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type ActiveTimelineVersionFinder interface {
	GetActiveByMediaFileID(ctx context.Context, mediaFileID uuid.UUID) (*domaintimeline.Timeline, error)
}

type UpdateEditorConfigurationCommand struct {
	MediaFileID       uuid.UUID
	UserID            uuid.UUID
	TimelineVersion   *int
	Silence           *SilenceConfigPatch
	FillerEnabled     *bool
	RepetitionEnabled *bool
	SubtitlesEnabled  *bool
	SubtitleMaxWords  *int
	// RebuildTimeline enqueues a timeline rebuild (used for silence config changes).
	RebuildTimeline bool
}

type SilenceConfigPatch struct {
	Enabled         *bool
	ThresholdMode   *string
	DetectionLevel  *string
	MinDurationMs   *int
	PaddingBeforeMs *int
	PaddingAfterMs  *int
	ThresholdDB     *float64
}

// EditorMutationResult is returned when an editor change enqueues a timeline rebuild.
type EditorMutationResult struct {
	JobID                   uuid.UUID
	PreviousTimelineVersion int
}

type UpdateEditorConfigurationHandler struct {
	mediaRepo    domainmediafile.MediaFileWriteRepository
	audioRepo    domainmediaaudio.MediaAudioWriteRepository
	configRepo   domainmediaconfig.MediaConfigurationWriteRepository
	timelineRepo ActiveTimelineVersionFinder
	jobRepo      domainjob.JobWriteRepository
	outbox       port.OutboxRepository
}

func NewUpdateEditorConfigurationHandler(
	mediaRepo domainmediafile.MediaFileWriteRepository,
	audioRepo domainmediaaudio.MediaAudioWriteRepository,
	configRepo domainmediaconfig.MediaConfigurationWriteRepository,
	timelineRepo ActiveTimelineVersionFinder,
	jobRepo domainjob.JobWriteRepository,
	outbox port.OutboxRepository,
) *UpdateEditorConfigurationHandler {
	return &UpdateEditorConfigurationHandler{
		mediaRepo:    mediaRepo,
		audioRepo:    audioRepo,
		configRepo:   configRepo,
		timelineRepo: timelineRepo,
		jobRepo:      jobRepo,
		outbox:       outbox,
	}
}

func (h *UpdateEditorConfigurationHandler) Handle(
	ctx context.Context,
	cmd UpdateEditorConfigurationCommand,
) (*EditorMutationResult, error) {
	media, err := h.requireOwnedMedia(ctx, cmd.MediaFileID, cmd.UserID)
	if err != nil {
		return nil, err
	}
	if err := assertTimelineVersion(ctx, h.timelineRepo, cmd.MediaFileID, cmd.TimelineVersion); err != nil {
		return nil, err
	}

	var result *EditorMutationResult
	err = h.configRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		cfg, err := h.configRepo.GetByMediaFileID(txCtx, cmd.MediaFileID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if cfg == nil {
			cfg = domainmediaconfig.NewDefault(cmd.MediaFileID)
			if err := h.configRepo.Save(txCtx, cfg); err != nil {
				return messaging.Retryable(err)
			}
		}
		needsRedetect := silenceAnalysisParamsChanging(cfg, cmd.Silence)
		applyConfigurationPatch(cfg, cmd)
		cfg.UpdatedAt = time.Now().UTC()
		if err := h.configRepo.Update(txCtx, cfg); err != nil {
			return messaging.Retryable(err)
		}
		if !cmd.RebuildTimeline {
			return nil
		}
		prevVersion := 0
		if cmd.TimelineVersion != nil {
			prevVersion = *cmd.TimelineVersion
		} else {
			tl, err := h.timelineRepo.GetActiveByMediaFileID(txCtx, cmd.MediaFileID)
			if err != nil {
				return messaging.Retryable(err)
			}
			if tl != nil {
				prevVersion = tl.Version
			}
		}

		// Analysis inputs (level / threshold) need a fresh ffmpeg pass first.
		// Rebuild is chained after silence_redetected — do not rebuild on stale silences.
		if needsRedetect {
			jobID, err := enqueueSilenceRedetection(
				txCtx,
				h.audioRepo,
				h.jobRepo,
				h.outbox,
				media.ProjectID,
				media.ID,
				media.UserID,
			)
			if err != nil {
				return err
			}
			result = &EditorMutationResult{JobID: jobID, PreviousTimelineVersion: prevVersion}
			return nil
		}

		jobID, err := enqueueTimelineRebuild(
			txCtx,
			h.jobRepo,
			h.outbox,
			media.ProjectID,
			media.ID,
			media.UserID,
			configurationRebuildReason(cmd),
		)
		if err != nil {
			return err
		}
		result = &EditorMutationResult{JobID: jobID, PreviousTimelineVersion: prevVersion}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (h *UpdateEditorConfigurationHandler) requireOwnedMedia(
	ctx context.Context,
	mediaFileID, userID uuid.UUID,
) (*domainmediafile.MediaFile, error) {
	media, err := h.mediaRepo.GetByID(ctx, mediaFileID)
	if err != nil {
		return nil, messaging.Retryable(err)
	}
	if media == nil || media.UserID != userID {
		return nil, messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
	}
	return media, nil
}

func applyConfigurationPatch(cfg *domainmediaconfig.MediaConfiguration, cmd UpdateEditorConfigurationCommand) {
	if cmd.Silence != nil {
		s := cmd.Silence
		if s.Enabled != nil {
			cfg.SilenceRemovalEnabled = *s.Enabled
		}
		if s.ThresholdMode != nil {
			cfg.SilenceThresholdMode = *s.ThresholdMode
		}
		if s.DetectionLevel != nil {
			cfg.SilenceDetectionLevel = *s.DetectionLevel
		}
		if s.MinDurationMs != nil {
			cfg.SilenceMinDurationMs = *s.MinDurationMs
		}
		if s.PaddingBeforeMs != nil {
			cfg.SilencePaddingBeforeMs = *s.PaddingBeforeMs
		}
		if s.PaddingAfterMs != nil {
			cfg.SilencePaddingAfterMs = *s.PaddingAfterMs
		}
		if s.ThresholdDB != nil {
			cfg.SilenceThresholdDB = s.ThresholdDB
			cfg.SilenceThresholdMode = domainmediaconfig.ThresholdModeManual
		}
	}
	if cmd.FillerEnabled != nil {
		cfg.FillerRemovalEnabled = *cmd.FillerEnabled
	}
	if cmd.RepetitionEnabled != nil {
		cfg.RepetitionRemovalEnabled = *cmd.RepetitionEnabled
	}
	if cmd.SubtitlesEnabled != nil {
		cfg.SubtitlesEnabled = *cmd.SubtitlesEnabled
	}
	if cmd.SubtitleMaxWords != nil {
		cfg.SubtitleMaxWords = *cmd.SubtitleMaxWords
	}
}

// silenceAnalysisParamsChanging reports whether the patch touches ffmpeg analysis
// inputs (threshold / aggressiveness). Those require re-detection; min duration and
// paddings only need a timeline rebuild over existing raw silences.
func silenceAnalysisParamsChanging(cfg *domainmediaconfig.MediaConfiguration, patch *SilenceConfigPatch) bool {
	if patch == nil {
		return false
	}
	if patch.DetectionLevel != nil && *patch.DetectionLevel != cfg.SilenceDetectionLevel {
		return true
	}
	if patch.ThresholdMode != nil && *patch.ThresholdMode != cfg.SilenceThresholdMode {
		return true
	}
	if patch.ThresholdDB != nil {
		if cfg.SilenceThresholdDB == nil || *cfg.SilenceThresholdDB != *patch.ThresholdDB {
			return true
		}
		if cfg.SilenceThresholdMode != domainmediaconfig.ThresholdModeManual {
			return true
		}
	}
	return false
}

func configurationRebuildReason(cmd UpdateEditorConfigurationCommand) string {
	if cmd.Silence != nil {
		return "silence_configuration_updated"
	}
	return "configuration_updated"
}

type OverrideDecisionCommand struct {
	MediaFileID     uuid.UUID
	UserID          uuid.UUID
	DecisionID      uuid.UUID
	Action          string
	TimelineVersion *int
}

type ClearDecisionOverrideCommand struct {
	MediaFileID     uuid.UUID
	UserID          uuid.UUID
	DecisionID      uuid.UUID
	TimelineVersion *int
}

type CreateManualDecisionCommand struct {
	MediaFileID     uuid.UUID
	UserID          uuid.UUID
	Action          string
	SourceStartMs   int64
	SourceEndMs     int64
	TimelineVersion *int
}

type DecisionMutationHandler struct {
	mediaRepo    domainmediafile.MediaFileWriteRepository
	configRepo   domainmediaconfig.MediaConfigurationWriteRepository
	silenceRepo  domainsilence.DetectedSilenceWriteRepository
	issueRepo    domaintranscriptissue.IssueWriteRepository
	overrideRepo OverrideStore
	timelineRepo ActiveTimelineVersionFinder
}

func NewDecisionMutationHandler(
	mediaRepo domainmediafile.MediaFileWriteRepository,
	configRepo domainmediaconfig.MediaConfigurationWriteRepository,
	silenceRepo domainsilence.DetectedSilenceWriteRepository,
	issueRepo domaintranscriptissue.IssueWriteRepository,
	overrideRepo OverrideStore,
	timelineRepo ActiveTimelineVersionFinder,
) *DecisionMutationHandler {
	return &DecisionMutationHandler{
		mediaRepo:    mediaRepo,
		configRepo:   configRepo,
		silenceRepo:  silenceRepo,
		issueRepo:    issueRepo,
		overrideRepo: overrideRepo,
		timelineRepo: timelineRepo,
	}
}

func (h *DecisionMutationHandler) Override(ctx context.Context, cmd OverrideDecisionCommand) error {
	media, err := h.requireOwnedMedia(ctx, cmd.MediaFileID, cmd.UserID)
	if err != nil {
		return err
	}
	if err := assertTimelineVersion(ctx, h.timelineRepo, cmd.MediaFileID, cmd.TimelineVersion); err != nil {
		return err
	}
	decision, err := h.findEditorDecision(ctx, media, cmd.DecisionID)
	if err != nil {
		return err
	}

	return h.overrideRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		_ = h.overrideRepo.DeleteMatching(txCtx, media.ID, decision.SourceStartMs, decision.SourceEndMs, domaintimeline.ActionKeep)
		_ = h.overrideRepo.DeleteMatching(txCtx, media.ID, decision.SourceStartMs, decision.SourceEndMs, domaintimeline.ActionRemove)
		if err := h.overrideRepo.Save(txCtx, domaintimeline.Override{
			ID:            uuid.New(),
			MediaFileID:   media.ID,
			Type:          domaintimeline.DecisionManual,
			SourceStartMs: decision.SourceStartMs,
			SourceEndMs:   decision.SourceEndMs,
			Action:        cmd.Action,
		}); err != nil {
			return messaging.Retryable(err)
		}
		return nil
	})
}

func (h *DecisionMutationHandler) ClearOverride(ctx context.Context, cmd ClearDecisionOverrideCommand) error {
	media, err := h.requireOwnedMedia(ctx, cmd.MediaFileID, cmd.UserID)
	if err != nil {
		return err
	}
	if err := assertTimelineVersion(ctx, h.timelineRepo, cmd.MediaFileID, cmd.TimelineVersion); err != nil {
		return err
	}
	decision, err := h.findEditorDecision(ctx, media, cmd.DecisionID)
	if err != nil {
		return err
	}

	return h.overrideRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.overrideRepo.DeleteMatching(txCtx, media.ID, decision.SourceStartMs, decision.SourceEndMs, domaintimeline.ActionKeep); err != nil {
			return messaging.Retryable(err)
		}
		if err := h.overrideRepo.DeleteMatching(txCtx, media.ID, decision.SourceStartMs, decision.SourceEndMs, domaintimeline.ActionRemove); err != nil {
			return messaging.Retryable(err)
		}
		return nil
	})
}

func (h *DecisionMutationHandler) CreateManual(ctx context.Context, cmd CreateManualDecisionCommand) error {
	media, err := h.requireOwnedMedia(ctx, cmd.MediaFileID, cmd.UserID)
	if err != nil {
		return err
	}
	if err := assertTimelineVersion(ctx, h.timelineRepo, cmd.MediaFileID, cmd.TimelineVersion); err != nil {
		return err
	}
	if cmd.SourceStartMs < 0 || cmd.SourceEndMs <= cmd.SourceStartMs {
		return messaging.NonRetryable(ErrInvalidManualRange)
	}
	if media.DurationMs > 0 && cmd.SourceEndMs > media.DurationMs {
		return messaging.NonRetryable(ErrInvalidManualRange)
	}

	return h.overrideRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.overrideRepo.Save(txCtx, domaintimeline.Override{
			ID:            uuid.New(),
			MediaFileID:   media.ID,
			Type:          domaintimeline.DecisionManual,
			SourceStartMs: cmd.SourceStartMs,
			SourceEndMs:   cmd.SourceEndMs,
			Action:        cmd.Action,
		}); err != nil {
			return messaging.Retryable(err)
		}
		return nil
	})
}

func (h *DecisionMutationHandler) requireOwnedMedia(
	ctx context.Context,
	mediaFileID, userID uuid.UUID,
) (*domainmediafile.MediaFile, error) {
	media, err := h.mediaRepo.GetByID(ctx, mediaFileID)
	if err != nil {
		return nil, messaging.Retryable(err)
	}
	if media == nil || media.UserID != userID {
		return nil, messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
	}
	return media, nil
}

func (h *DecisionMutationHandler) findEditorDecision(
	ctx context.Context,
	media *domainmediafile.MediaFile,
	decisionID uuid.UUID,
) (domaintimeline.EditorDecision, error) {
	cfg, err := h.configRepo.GetByMediaFileID(ctx, media.ID)
	if err != nil {
		return domaintimeline.EditorDecision{}, messaging.Retryable(err)
	}
	if cfg == nil {
		cfg = domainmediaconfig.NewDefault(media.ID)
	}
	silences, err := h.silenceRepo.ListByMediaFileID(ctx, media.ID)
	if err != nil {
		return domaintimeline.EditorDecision{}, messaging.Retryable(err)
	}
	issues, err := h.issueRepo.ListByMediaFileID(ctx, media.ID)
	if err != nil {
		return domaintimeline.EditorDecision{}, messaging.Retryable(err)
	}
	overrides, err := h.overrideRepo.ListByMediaFileID(ctx, media.ID)
	if err != nil {
		return domaintimeline.EditorDecision{}, messaging.Retryable(err)
	}
	auto := domaintimeline.BuildDecisions(domaintimeline.BuildInput{
		MediaFileID:          media.ID,
		MediaDurationMs:      media.DurationMs,
		SilenceRemoval:       cfg.SilenceRemovalEnabled,
		FillerRemoval:        cfg.FillerRemovalEnabled,
		RepetitionRemoval:    cfg.RepetitionRemovalEnabled,
		SilenceMinDurationMs: cfg.SilenceMinDurationMs,
		SilencePadBeforeMs:   cfg.SilencePaddingBeforeMs,
		SilencePadAfterMs:    cfg.SilencePaddingAfterMs,
		SpeechMinDurationMs:  cfg.SpeechMinDurationMs,
		Silences:             domainsilence.IntervalsFromDetected(silences),
		Issues:               issues,
		Overrides:            nil,
	})
	resolved := domaintimeline.ResolveEditorDecisions(auto, overrides, domaintimeline.LabelsFromIssues(issues))
	for _, d := range resolved {
		if d.ID == decisionID {
			return d, nil
		}
	}
	return domaintimeline.EditorDecision{}, messaging.NonRetryable(ErrDecisionNotFound)
}

func assertTimelineVersion(
	ctx context.Context,
	timelineRepo ActiveTimelineVersionFinder,
	mediaFileID uuid.UUID,
	expected *int,
) error {
	tl, err := timelineRepo.GetActiveByMediaFileID(ctx, mediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if tl == nil {
		return messaging.NonRetryable(querymediafile.ErrEditorNotReady)
	}
	if expected != nil && tl.Version != *expected {
		return &StaleTimelineError{CurrentVersion: tl.Version}
	}
	return nil
}

type StaleTimelineError struct {
	CurrentVersion int
}

func (e *StaleTimelineError) Error() string        { return ErrStaleTimeline.Error() }
func (e *StaleTimelineError) Is(target error) bool { return target == ErrStaleTimeline }

func enqueueTimelineRebuild(
	ctx context.Context,
	jobRepo domainjob.JobWriteRepository,
	outbox port.OutboxRepository,
	projectID, mediaFileID, userID uuid.UUID,
	reason string,
) (uuid.UUID, error) {
	timelineJob := domainjob.New(projectID, mediaFileID, userID, domainjob.NameRebuildTimeline)
	if err := jobRepo.Save(ctx, timelineJob); err != nil {
		return uuid.Nil, messaging.Retryable(err)
	}
	events := make([]event.DomainEvent, 0, 2)
	events = append(events, timelineJob.PullEvents()...)
	events = append(events, domaintimeline.RebuildRequested{
		ID:          uuid.New().String(),
		MediaFileID: mediaFileID.String(),
		ProjectID:   projectID.String(),
		UserID:      userID.String(),
		JobID:       timelineJob.ID.String(),
		Reason:      reason,
		Timestamp:   time.Now().UTC(),
	})
	log.Printf("timeline rebuild requested mediaFileId=%s jobId=%s reason=%s", mediaFileID, timelineJob.ID, reason)
	if err := outbox.StoreEvents(ctx, events); err != nil {
		return uuid.Nil, err
	}
	return timelineJob.ID, nil
}

func enqueueSilenceRedetection(
	ctx context.Context,
	audioRepo domainmediaaudio.MediaAudioWriteRepository,
	jobRepo domainjob.JobWriteRepository,
	outbox port.OutboxRepository,
	projectID, mediaFileID, userID uuid.UUID,
) (uuid.UUID, error) {
	audio, err := audioRepo.GetByMediaFileID(ctx, mediaFileID)
	if err != nil {
		return uuid.Nil, messaging.Retryable(err)
	}
	if audio == nil || audio.StorageKey == "" {
		return uuid.Nil, messaging.NonRetryable(querymediafile.ErrEditorNotReady)
	}

	silenceJob := domainjob.New(projectID, mediaFileID, userID, domainjob.NameDetectSilence)
	if err := jobRepo.Save(ctx, silenceJob); err != nil {
		return uuid.Nil, messaging.Retryable(err)
	}
	events := make([]event.DomainEvent, 0, 2)
	events = append(events, silenceJob.PullEvents()...)
	events = append(events, domainmediafile.MediaFileSilenceRequested{
		ID:          uuid.New().String(),
		MediaFileID: mediaFileID.String(),
		ProjectID:   projectID.String(),
		UserID:      userID.String(),
		JobID:       silenceJob.ID.String(),
		AudioKey:    audio.StorageKey,
		Force:       true,
		Timestamp:   time.Now().UTC(),
	})
	log.Printf("silence redetection requested mediaFileId=%s jobId=%s", mediaFileID, silenceJob.ID)
	if err := outbox.StoreEvents(ctx, events); err != nil {
		return uuid.Nil, err
	}
	return silenceJob.ID, nil
}
