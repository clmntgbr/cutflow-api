package mediafile

import (
	"context"
	"log"
	"time"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/event"
	domainjob "go-api/internal/domain/job"
	domainmediafile "go-api/internal/domain/mediafile"
	domaintimeline "go-api/internal/domain/timeline"

	"github.com/google/uuid"
)

const ReasonInitialAnalysisReady = "initial_analysis_ready"

// DecideInitialTimelineEnqueue chooses whether the initial pipeline barrier
// should fire a timeline rebuild.
//
// After a timeline already exists, callers rebuild immediately (not via this helper).
func DecideInitialTimelineEnqueue(silenceReady, textReady bool) bool {
	return silenceReady && textReady
}

type timelineRebuildDeps struct {
	mediaRepo    domainmediafile.MediaFileWriteRepository
	jobRepo      domainjob.JobWriteRepository
	timelineRepo domaintimeline.TimelineWriteRepository
}

// appendTimelineRebuildIfReady enforces the initial analysis barrier:
//   - no timeline yet → wait until silence + text-analysis jobs succeeded, then enqueue once
//   - timeline already exists → enqueue immediately (post READY_TO_EDIT rebuilds)
//
// Must run inside an open DB transaction. Locks media_file to serialize concurrent completions.
func appendTimelineRebuildIfReady(
	ctx context.Context,
	deps timelineRebuildDeps,
	events *[]event.DomainEvent,
	projectID, mediaFileID, userID uuid.UUID,
	triggerReason string,
) error {
	if _, err := deps.mediaRepo.GetByID(ctx, mediaFileID); err != nil {
		return messaging.Retryable(err)
	}

	hasTimeline, err := deps.timelineRepo.ExistsByMediaFileID(ctx, mediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}

	reason := triggerReason
	if !hasTimeline {
		silenceReady, err := deps.jobRepo.HasSuccessful(ctx, mediaFileID, domainjob.NameDetectSilence)
		if err != nil {
			return messaging.Retryable(err)
		}
		textReady, err := deps.jobRepo.HasSuccessful(ctx, mediaFileID, domainjob.NameAnalyzeTranscript)
		if err != nil {
			return messaging.Retryable(err)
		}
		if !DecideInitialTimelineEnqueue(silenceReady, textReady) {
			log.Printf(
				"timeline barrier waiting mediaFileId=%s silenceReady=%t textReady=%t trigger=%s",
				mediaFileID, silenceReady, textReady, triggerReason,
			)
			return nil
		}
		reason = ReasonInitialAnalysisReady
	}

	timelineJob := domainjob.New(projectID, mediaFileID, userID, domainjob.NameRebuildTimeline)
	if err := deps.jobRepo.Save(ctx, timelineJob); err != nil {
		return messaging.Retryable(err)
	}
	*events = append(*events, timelineJob.PullEvents()...)
	*events = append(*events, domaintimeline.RebuildRequested{
		ID:          uuid.New().String(),
		MediaFileID: mediaFileID.String(),
		ProjectID:   projectID.String(),
		UserID:      userID.String(),
		JobID:       timelineJob.ID.String(),
		Reason:      reason,
		Timestamp:   time.Now().UTC(),
	})
	log.Printf(
		"timeline rebuild requested mediaFileId=%s jobId=%s reason=%s initial=%t",
		mediaFileID, timelineJob.ID, reason, !hasTimeline,
	)
	return nil
}
