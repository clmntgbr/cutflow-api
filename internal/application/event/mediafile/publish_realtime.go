package mediafile

import (
	"context"
	"encoding/json"
	"time"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	domainmediaaudio "go-api/internal/domain/mediaaudio"
	domainmediafile "go-api/internal/domain/mediafile"
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

type mediaFileRealtimePayload struct {
	MediaFileID     string    `json:"mediaFileId"`
	ProjectID       string    `json:"projectId"`
	Status          string    `json:"status"`
	ThumbnailURL    string    `json:"thumbnailUrl,omitempty"`
	DurationMs      int64     `json:"durationMs,omitempty"`
	Width           *int      `json:"width,omitempty"`
	Height          *int      `json:"height,omitempty"`
	FPS             *float64  `json:"fps,omitempty"`
	VideoCodec      string    `json:"videoCodec,omitempty"`
	AudioCodec      string    `json:"audioCodec,omitempty"`
	AudioSampleRate *int      `json:"audioSampleRate,omitempty"`
	AudioChannels   *int      `json:"audioChannels,omitempty"`
	SizeBytes       int64     `json:"sizeBytes,omitempty"`
	SegmentCount    int       `json:"segmentCount,omitempty"`
	HasAudio        bool      `json:"hasAudio,omitempty"`
	Reason          string    `json:"reason,omitempty"`
	OccurredAt      time.Time `json:"occurredAt"`
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

func (h *PublishRealtimeHandler) OnProbing(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileProbing
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityMediaFile, realtime.ActionProbing, evt.UserID, mediaFileRealtimePayload{
		MediaFileID: evt.MediaFileID,
		ProjectID:   evt.ProjectID,
		Status:      evt.Status,
		OccurredAt:  evt.Timestamp,
	})
}

func (h *PublishRealtimeHandler) OnReady(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileReady
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityMediaFile, realtime.ActionReady, evt.UserID, mediaFileRealtimePayload{
		MediaFileID:     evt.MediaFileID,
		ProjectID:       evt.ProjectID,
		Status:          evt.Status,
		DurationMs:      evt.DurationMs,
		Width:           evt.Width,
		Height:          evt.Height,
		FPS:             evt.FPS,
		VideoCodec:      evt.VideoCodec,
		AudioCodec:      evt.AudioCodec,
		AudioSampleRate: evt.AudioSampleRate,
		AudioChannels:   evt.AudioChannels,
		SizeBytes:       evt.SizeBytes,
		OccurredAt:      evt.Timestamp,
	})
}

func (h *PublishRealtimeHandler) OnProbeFailed(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileProbeFailed
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityMediaFile, realtime.ActionFailed, evt.UserID, mediaFileRealtimePayload{
		MediaFileID: evt.MediaFileID,
		ProjectID:   evt.ProjectID,
		Status:      evt.Status,
		Reason:      evt.Reason,
		OccurredAt:  evt.Timestamp,
	})
}

func (h *PublishRealtimeHandler) OnSegmentsReady(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileSegmentsReady
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityMediaFile, realtime.ActionSegmentsReady, evt.UserID, mediaFileRealtimePayload{
		MediaFileID:  evt.MediaFileID,
		ProjectID:    evt.ProjectID,
		Status:       domainmediafile.StatusReady,
		DurationMs:   evt.DurationMs,
		SegmentCount: evt.SegmentCount,
		HasAudio:     evt.HasAudio,
		OccurredAt:   evt.Timestamp,
	})
}

func (h *PublishRealtimeHandler) OnAudioReady(ctx context.Context, payload []byte) error {
	var evt domainmediaaudio.MediaAudioReady
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityMediaFile, realtime.ActionAudioReady, evt.UserID, mediaFileRealtimePayload{
		MediaFileID: evt.MediaFileID,
		ProjectID:   evt.ProjectID,
		Status:      evt.Status,
		SizeBytes:   evt.SizeBytes,
		OccurredAt:  evt.Timestamp,
	})
}

func (h *PublishRealtimeHandler) OnAudioFailed(ctx context.Context, payload []byte) error {
	var evt domainmediaaudio.MediaAudioFailed
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityMediaFile, realtime.ActionFailed, evt.UserID, mediaFileRealtimePayload{
		MediaFileID: evt.MediaFileID,
		ProjectID:   evt.ProjectID,
		Status:      evt.Status,
		Reason:      evt.Reason,
		OccurredAt:  evt.Timestamp,
	})
}
