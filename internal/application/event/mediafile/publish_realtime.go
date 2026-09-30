package mediafile

import (
	"context"
	"encoding/json"
	"time"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/domain/port"
	domaintimeline "go-api/internal/domain/timeline"
)

type PublishRealtimeHandler struct {
	publisher *realtime.Publisher
}

func NewPublishRealtimeHandler(realtimePublisher port.RealtimePublisher) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{
		publisher: realtime.NewPublisher(realtimePublisher),
	}
}

type mediaFileRealtimePayload struct {
	MediaFileID  string    `json:"mediaFileId"`
	ProjectID    string    `json:"projectId"`
	Status       string    `json:"status,omitempty"`
	ThumbnailURL string    `json:"thumbnailUrl,omitempty"`
	SizeBytes    int64     `json:"sizeBytes,omitempty"`
	TimelineID   string    `json:"timelineId,omitempty"`
	Version      int       `json:"version,omitempty"`
	DurationMs   int64     `json:"durationMs,omitempty"`
	JobID        string    `json:"jobId,omitempty"`
	Reason       string    `json:"reason,omitempty"`
	OccurredAt   time.Time `json:"occurredAt"`
}

func (h *PublishRealtimeHandler) OnUploaded(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileUploaded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityMediaFile, realtime.ActionUploaded, evt.UserID, mediaFileRealtimePayload{
		MediaFileID: evt.MediaFileID,
		ProjectID:   evt.ProjectID,
		Status:      evt.Status,
		SizeBytes:   evt.SizeBytes,
		OccurredAt:  evt.Timestamp,
	})
}

func (h *PublishRealtimeHandler) OnThumbnailReady(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileThumbnailReady
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(
		ctx,
		realtime.EntityMediaFile,
		realtime.ActionThumbnailReady,
		evt.UserID,
		mediaFileRealtimePayload{
			MediaFileID:  evt.MediaFileID,
			ProjectID:    evt.ProjectID,
			Status:       evt.Status,
			ThumbnailURL: "/api/media-files/" + evt.MediaFileID + "/thumbnail",
			OccurredAt:   evt.Timestamp,
		},
	)
}

func (h *PublishRealtimeHandler) OnTimelineUpdated(ctx context.Context, payload []byte) error {
	var evt domaintimeline.Updated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityMediaFile, realtime.ActionTimelineUpdated, evt.UserID, mediaFileRealtimePayload{
		MediaFileID: evt.MediaFileID,
		ProjectID:   evt.ProjectID,
		TimelineID:  evt.TimelineID,
		Version:     evt.Version,
		DurationMs:  evt.DurationMs,
		OccurredAt:  evt.Timestamp,
	})
}

func (h *PublishRealtimeHandler) OnTimelineFailed(ctx context.Context, payload []byte) error {
	var evt domaintimeline.Failed
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityMediaFile, realtime.ActionTimelineFailed, evt.UserID, mediaFileRealtimePayload{
		MediaFileID: evt.MediaFileID,
		ProjectID:   evt.ProjectID,
		JobID:       evt.JobID,
		Reason:      evt.Reason,
		OccurredAt:  evt.Timestamp,
	})
}
