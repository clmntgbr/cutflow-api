package mediafile

import (
	"context"
	"encoding/json"
	"log"

	cmdmediafile "go-api/internal/application/command/mediafile"
	"go-api/internal/application/messaging"
	domainmediafile "go-api/internal/domain/mediafile"
	domaintimeline "go-api/internal/domain/timeline"

	"github.com/google/uuid"
)

type DetectSilenceOnRequestedHandler struct {
	detect *cmdmediafile.DetectSilenceHandler
}

func NewDetectSilenceOnRequestedHandler(detect *cmdmediafile.DetectSilenceHandler) *DetectSilenceOnRequestedHandler {
	return &DetectSilenceOnRequestedHandler{detect: detect}
}

func (h *DetectSilenceOnRequestedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileSilenceRequested
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	mediaFileID, err := uuid.Parse(evt.MediaFileID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	userID, err := uuid.Parse(evt.UserID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	jobID := uuid.Nil
	if evt.JobID != "" {
		jobID, err = uuid.Parse(evt.JobID)
		if err != nil {
			return messaging.NonRetryable(err)
		}
	}

	log.Printf("silence worker received mediaFileId=%s jobId=%s", evt.MediaFileID, evt.JobID)
	return h.detect.Handle(ctx, cmdmediafile.DetectSilenceCommand{
		MediaFileID: mediaFileID,
		AudioKey:    evt.AudioKey,
		ProjectID:   projectID,
		UserID:      userID,
		JobID:       jobID,
	})
}

type TranscribeAudioOnRequestedHandler struct {
	transcribe *cmdmediafile.TranscribeAudioHandler
}

func NewTranscribeAudioOnRequestedHandler(
	transcribe *cmdmediafile.TranscribeAudioHandler,
) *TranscribeAudioOnRequestedHandler {
	return &TranscribeAudioOnRequestedHandler{transcribe: transcribe}
}

func (h *TranscribeAudioOnRequestedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileTranscriptRequested
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	mediaFileID, err := uuid.Parse(evt.MediaFileID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	userID, err := uuid.Parse(evt.UserID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	jobID := uuid.Nil
	if evt.JobID != "" {
		jobID, err = uuid.Parse(evt.JobID)
		if err != nil {
			return messaging.NonRetryable(err)
		}
	}

	log.Printf("transcript worker received mediaFileId=%s jobId=%s", evt.MediaFileID, evt.JobID)
	return h.transcribe.Handle(ctx, cmdmediafile.TranscribeAudioCommand{
		MediaFileID: mediaFileID,
		AudioKey:    evt.AudioKey,
		ProjectID:   projectID,
		UserID:      userID,
		JobID:       jobID,
	})
}

type AnalyzeTranscriptOnRequestedHandler struct {
	analyze *cmdmediafile.AnalyzeTranscriptHandler
}

func NewAnalyzeTranscriptOnRequestedHandler(
	analyze *cmdmediafile.AnalyzeTranscriptHandler,
) *AnalyzeTranscriptOnRequestedHandler {
	return &AnalyzeTranscriptOnRequestedHandler{analyze: analyze}
}

func (h *AnalyzeTranscriptOnRequestedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileAnalysisRequested
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	mediaFileID, err := uuid.Parse(evt.MediaFileID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	userID, err := uuid.Parse(evt.UserID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	transcriptID := uuid.Nil
	if evt.TranscriptID != "" {
		transcriptID, err = uuid.Parse(evt.TranscriptID)
		if err != nil {
			return messaging.NonRetryable(err)
		}
	}
	jobID := uuid.Nil
	if evt.JobID != "" {
		jobID, err = uuid.Parse(evt.JobID)
		if err != nil {
			return messaging.NonRetryable(err)
		}
	}

	log.Printf("analysis worker received mediaFileId=%s jobId=%s", evt.MediaFileID, evt.JobID)
	return h.analyze.Handle(ctx, cmdmediafile.AnalyzeTranscriptCommand{
		MediaFileID:  mediaFileID,
		TranscriptID: transcriptID,
		ProjectID:    projectID,
		UserID:       userID,
		JobID:        jobID,
	})
}

type AnalyzeViralOnRequestedHandler struct {
	analyze *cmdmediafile.AnalyzeViralHandler
}

func NewAnalyzeViralOnRequestedHandler(analyze *cmdmediafile.AnalyzeViralHandler) *AnalyzeViralOnRequestedHandler {
	return &AnalyzeViralOnRequestedHandler{analyze: analyze}
}

func (h *AnalyzeViralOnRequestedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileViralRequested
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	mediaFileID, err := uuid.Parse(evt.MediaFileID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	userID, err := uuid.Parse(evt.UserID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	transcriptID := uuid.Nil
	if evt.TranscriptID != "" {
		transcriptID, err = uuid.Parse(evt.TranscriptID)
		if err != nil {
			return messaging.NonRetryable(err)
		}
	}
	jobID := uuid.Nil
	if evt.JobID != "" {
		jobID, err = uuid.Parse(evt.JobID)
		if err != nil {
			return messaging.NonRetryable(err)
		}
	}

	log.Printf("viral worker received mediaFileId=%s jobId=%s", evt.MediaFileID, evt.JobID)
	return h.analyze.Handle(ctx, cmdmediafile.AnalyzeViralCommand{
		MediaFileID:  mediaFileID,
		TranscriptID: transcriptID,
		ProjectID:    projectID,
		UserID:       userID,
		JobID:        jobID,
	})
}

type RebuildTimelineOnRequestedHandler struct {
	rebuild *cmdmediafile.RebuildTimelineHandler
}

func NewRebuildTimelineOnRequestedHandler(rebuild *cmdmediafile.RebuildTimelineHandler) *RebuildTimelineOnRequestedHandler {
	return &RebuildTimelineOnRequestedHandler{rebuild: rebuild}
}

func (h *RebuildTimelineOnRequestedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domaintimeline.RebuildRequested
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	mediaFileID, err := uuid.Parse(evt.MediaFileID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	userID, err := uuid.Parse(evt.UserID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	jobID := uuid.Nil
	if evt.JobID != "" {
		jobID, err = uuid.Parse(evt.JobID)
		if err != nil {
			return messaging.NonRetryable(err)
		}
	}

	log.Printf("timeline worker received mediaFileId=%s jobId=%s reason=%s", evt.MediaFileID, evt.JobID, evt.Reason)
	return h.rebuild.Handle(ctx, cmdmediafile.RebuildTimelineCommand{
		MediaFileID: mediaFileID,
		ProjectID:   projectID,
		UserID:      userID,
		JobID:       jobID,
		Reason:      evt.Reason,
	})
}

type ExtractAudioOnReadyHandler struct {
	extract *cmdmediafile.ExtractAudioHandler
}

func NewExtractAudioOnReadyHandler(extract *cmdmediafile.ExtractAudioHandler) *ExtractAudioOnReadyHandler {
	return &ExtractAudioOnReadyHandler{extract: extract}
}

func (h *ExtractAudioOnReadyHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileReady
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	if evt.AudioCodec == "" {
		log.Printf("extract audio skipped mediaFileId=%s: no audio track", evt.MediaFileID)
		return nil
	}
	mediaFileID, err := uuid.Parse(evt.MediaFileID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf("extract audio received event type=%s mediaFileId=%s", evt.EventType(), evt.MediaFileID)
	return h.extract.Handle(ctx, cmdmediafile.ExtractAudioCommand{MediaFileID: mediaFileID})
}
