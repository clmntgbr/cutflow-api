package job

import (
	"context"
	"encoding/json"
	"time"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	domainjob "go-api/internal/domain/job"
	"go-api/internal/domain/port"
)

type PublishRealtimeHandler struct {
	publisher *realtime.Publisher
}

func NewPublishRealtimeHandler(realtimePublisher port.RealtimePublisher) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{
		publisher: realtime.NewPublisher(realtimePublisher),
	}
}

type jobRealtimePayload struct {
	JobID        string    `json:"jobId"`
	ProjectID    string    `json:"projectId"`
	MediaFileID  string    `json:"mediaFileId"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
	OccurredAt   time.Time `json:"occurredAt"`
}

func (h *PublishRealtimeHandler) OnCreated(ctx context.Context, payload []byte) error {
	var evt domainjob.JobCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityJob, realtime.ActionCreated, evt.UserID, jobRealtimePayload{
		JobID:       evt.JobID,
		ProjectID:   evt.ProjectID,
		MediaFileID: evt.MediaFileID,
		Name:        evt.Name,
		Status:      evt.Status,
		OccurredAt:  evt.Timestamp,
	})
}

func (h *PublishRealtimeHandler) OnStatusChanged(ctx context.Context, payload []byte) error {
	var evt domainjob.JobStatusChanged
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityJob, realtime.ActionUpdated, evt.UserID, jobRealtimePayload{
		JobID:        evt.JobID,
		ProjectID:    evt.ProjectID,
		MediaFileID:  evt.MediaFileID,
		Name:         evt.Name,
		Status:       evt.Status,
		ErrorMessage: evt.ErrorMessage,
		OccurredAt:   evt.Timestamp,
	})
}
