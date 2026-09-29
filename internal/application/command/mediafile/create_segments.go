package mediafile

import (
	"context"
	"time"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/event"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"
	domainsegment "go-api/internal/domain/segment"

	"github.com/google/uuid"
)

type CreateSegmentsCommand struct {
	MediaFileID uuid.UUID
}

type CreateSegmentsHandler struct {
	mediaRepo   domainmediafile.MediaFileWriteRepository
	segmentRepo domainsegment.SegmentWriteRepository
	outbox      port.OutboxRepository
	planCfg     domainsegment.PlanConfig
}

func NewCreateSegmentsHandler(
	mediaRepo domainmediafile.MediaFileWriteRepository,
	segmentRepo domainsegment.SegmentWriteRepository,
	outbox port.OutboxRepository,
	planCfg domainsegment.PlanConfig,
) *CreateSegmentsHandler {
	return &CreateSegmentsHandler{
		mediaRepo:   mediaRepo,
		segmentRepo: segmentRepo,
		outbox:      outbox,
		planCfg:     domainsegment.NormalizePlanConfig(planCfg),
	}
}

func (h *CreateSegmentsHandler) Handle(ctx context.Context, cmd CreateSegmentsCommand) error {
	media, err := h.mediaRepo.GetByID(ctx, cmd.MediaFileID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if media == nil {
		return messaging.NonRetryable(domainmediafile.ErrMediaNotFound)
	}
	if media.Status != domainmediafile.StatusReady &&
		media.Status != domainmediafile.StatusProcessing &&
		media.Status != domainmediafile.StatusCompleted {
		return messaging.NonRetryable(domainmediafile.ErrInvalidTransition)
	}

	existing, err := h.segmentRepo.CountByMediaFileID(ctx, media.ID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if existing > 0 {
		return nil
	}

	windows := domainsegment.PlanWindows(media.DurationMs, h.planCfg)
	if len(windows) == 0 {
		return messaging.NonRetryable(domainmediafile.ErrProbeInvalid)
	}

	segments := make([]*domainsegment.Segment, 0, len(windows))
	for _, window := range windows {
		segments = append(segments, domainsegment.NewFromWindow(media.ID, window))
	}

	evt := domainmediafile.MediaFileSegmentsReady{
		ID:           uuid.New().String(),
		MediaFileID:  media.ID.String(),
		ProjectID:    media.ProjectID.String(),
		UserID:       media.UserID.String(),
		SegmentCount: len(segments),
		DurationMs:   media.DurationMs,
		HasAudio:     media.AudioCodec != "",
		Timestamp:    time.Now().UTC(),
	}

	return h.segmentRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		count, err := h.segmentRepo.CountByMediaFileID(txCtx, media.ID)
		if err != nil {
			return messaging.Retryable(err)
		}
		if count > 0 {
			return nil
		}
		if err := h.segmentRepo.SaveBatch(txCtx, segments); err != nil {
			return messaging.Retryable(err)
		}
		return h.outbox.StoreEvents(txCtx, []event.DomainEvent{evt})
	})
}
