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
	domainsilence "go-api/internal/domain/silence"
	domaintranscript "go-api/internal/domain/transcript"
	domaintranscriptissue "go-api/internal/domain/transcriptissue"
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
	HasAudio        bool      `json:"hasAudio,omitempty"`
	SilenceCount    int       `json:"silenceCount,omitempty"`
	WordCount       int       `json:"wordCount,omitempty"`
	Language        string    `json:"language,omitempty"`
	SRTKey          string    `json:"srtKey,omitempty"`
	ASSKey          string    `json:"assKey,omitempty"`
	FillerCount     int       `json:"fillerCount,omitempty"`
	RepetitionCount int       `json:"repetitionCount,omitempty"`
	FalseStartCount int       `json:"falseStartCount,omitempty"`
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

func (h *PublishRealtimeHandler) OnSilenceDetected(ctx context.Context, payload []byte) error {
	var evt domainsilence.SilenceDetected
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityMediaFile, realtime.ActionSilenceDetected, evt.UserID, mediaFileRealtimePayload{
		MediaFileID:  evt.MediaFileID,
		ProjectID:    evt.ProjectID,
		SilenceCount: evt.SilenceCount,
		OccurredAt:   evt.Timestamp,
	})
}

func (h *PublishRealtimeHandler) OnTranscriptReady(ctx context.Context, payload []byte) error {
	var evt domaintranscript.TranscriptReady
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityMediaFile, realtime.ActionTranscriptReady, evt.UserID, mediaFileRealtimePayload{
		MediaFileID: evt.MediaFileID,
		ProjectID:   evt.ProjectID,
		Status:      evt.Status,
		Language:    evt.Language,
		WordCount:   evt.WordCount,
		SRTKey:      evt.SRTKey,
		ASSKey:      evt.ASSKey,
		OccurredAt:  evt.Timestamp,
	})
}

func (h *PublishRealtimeHandler) OnTranscriptFailed(ctx context.Context, payload []byte) error {
	var evt domaintranscript.TranscriptFailed
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

func (h *PublishRealtimeHandler) OnTranscriptAnalysisReady(ctx context.Context, payload []byte) error {
	var evt domaintranscriptissue.TranscriptAnalysisReady
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityMediaFile, realtime.ActionTranscriptAnalysisReady, evt.UserID, mediaFileRealtimePayload{
		MediaFileID:     evt.MediaFileID,
		ProjectID:       evt.ProjectID,
		FillerCount:     evt.FillerCount,
		RepetitionCount: evt.RepetitionCount,
		FalseStartCount: evt.FalseStartCount,
		OccurredAt:      evt.Timestamp,
	})
}
