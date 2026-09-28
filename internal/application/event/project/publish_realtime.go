package project

import (
	"context"
	"encoding/json"
	"time"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	"go-api/internal/domain/port"
	domainproject "go-api/internal/domain/project"
)

type PublishRealtimeHandler struct {
	publisher *realtime.Publisher
}

func NewPublishRealtimeHandler(realtimePublisher port.RealtimePublisher) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{
		publisher: realtime.NewPublisher(realtimePublisher),
	}
}

type projectRealtimePayload struct {
	ProjectID  string    `json:"projectId"`
	Name       string    `json:"name,omitempty"`
	Status     string    `json:"status"`
	OccurredAt time.Time `json:"occurredAt"`
}

func (h *PublishRealtimeHandler) OnUpdated(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityProject, realtime.ActionUpdated, evt.UserID, projectRealtimePayload{
		ProjectID:  evt.ProjectID,
		Name:       evt.Name,
		Status:     evt.Status,
		OccurredAt: evt.Timestamp,
	})
}
