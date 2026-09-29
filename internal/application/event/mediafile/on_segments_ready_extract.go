package mediafile

import (
	"context"
	"encoding/json"
	"log"

	cmdmediafile "go-api/internal/application/command/mediafile"
	"go-api/internal/application/messaging"
	domainmediafile "go-api/internal/domain/mediafile"

	"github.com/google/uuid"
)

type ExtractAudioOnSegmentsReadyHandler struct {
	extract *cmdmediafile.ExtractAudioHandler
}

func NewExtractAudioOnSegmentsReadyHandler(
	extract *cmdmediafile.ExtractAudioHandler,
) *ExtractAudioOnSegmentsReadyHandler {
	return &ExtractAudioOnSegmentsReadyHandler{extract: extract}
}

func (h *ExtractAudioOnSegmentsReadyHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileSegmentsReady
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	if !evt.HasAudio {
		log.Printf("extract audio skipped mediaFileId=%s: hasAudio=false", evt.MediaFileID)
		return nil
	}
	mediaFileID, err := uuid.Parse(evt.MediaFileID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	log.Printf("extract audio received event type=%s mediaFileId=%s", evt.EventType(), evt.MediaFileID)
	if err := h.extract.Handle(ctx, cmdmediafile.ExtractAudioCommand{MediaFileID: mediaFileID}); err != nil {
		log.Printf("extract audio failed mediaFileId=%s: %v", evt.MediaFileID, err)
		return err
	}
	log.Printf("extract audio finished mediaFileId=%s", evt.MediaFileID)
	return nil
}
