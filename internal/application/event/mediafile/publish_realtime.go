package mediafile

import (
	"context"
	"encoding/json"
	"time"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	"go-api/internal/domain/port"
	domainmediafile "go-api/internal/domain/mediafile"
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
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	Status      string    `json:"status"`
	OccurredAt  time.Time `json:"occurredAt"`
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
		OccurredAt:  evt.Timestamp,
	})
}
