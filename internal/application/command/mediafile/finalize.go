package mediafile

import (
	"context"

	"go-api/internal/application/messaging"
	querymediafile "go-api/internal/application/query/mediafile"
	domainjob "go-api/internal/domain/job"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

// FinalizeEditorCommand validates the active timeline then enqueues a single rebuild
// that applies all editor mutations accumulated during the session.
// RenderPlan persistence is deferred until the render schema exists.
type FinalizeEditorCommand struct {
	MediaFileID     uuid.UUID
	UserID          uuid.UUID
	TimelineID      uuid.UUID
	TimelineVersion int
}

type FinalizeEditorResult struct {
	Status                  string
	JobID                   uuid.UUID
	TimelineID              uuid.UUID
	PreviousTimelineVersion int
}

type FinalizeEditorHandler struct {
	mediaRepo    domainmediafile.MediaFileWriteRepository
	timelineRepo ActiveTimelineVersionFinder
	jobRepo      domainjob.JobWriteRepository
	outbox       port.OutboxRepository
}

func NewFinalizeEditorHandler(
	mediaRepo domainmediafile.MediaFileWriteRepository,
	timelineRepo ActiveTimelineVersionFinder,
	jobRepo domainjob.JobWriteRepository,
	outbox port.OutboxRepository,
) *FinalizeEditorHandler {
	return &FinalizeEditorHandler{
		mediaRepo:    mediaRepo,
		timelineRepo: timelineRepo,
		jobRepo:      jobRepo,
		outbox:       outbox,
	}
}

func (h *FinalizeEditorHandler) Handle(
	ctx context.Context,
	cmd FinalizeEditorCommand,
) (*FinalizeEditorResult, error) {
	media, err := h.mediaRepo.GetByID(ctx, cmd.MediaFileID)
	if err != nil {
		return nil, messaging.Retryable(err)
	}
	if media == nil || media.UserID != cmd.UserID {
		return nil, messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
	}

	tl, err := h.timelineRepo.GetActiveByMediaFileID(ctx, cmd.MediaFileID)
	if err != nil {
		return nil, messaging.Retryable(err)
	}
	if tl == nil {
		return nil, messaging.NonRetryable(querymediafile.ErrEditorNotReady)
	}
	if tl.ID != cmd.TimelineID || tl.Version != cmd.TimelineVersion {
		return nil, &StaleTimelineError{CurrentVersion: tl.Version}
	}

	jobID, err := enqueueTimelineRebuild(
		ctx,
		h.jobRepo,
		h.outbox,
		media.ProjectID,
		media.ID,
		media.UserID,
		"finalize",
	)
	if err != nil {
		return nil, err
	}

	return &FinalizeEditorResult{
		Status:                  "accepted",
		JobID:                   jobID,
		TimelineID:              tl.ID,
		PreviousTimelineVersion: tl.Version,
	}, nil
}
