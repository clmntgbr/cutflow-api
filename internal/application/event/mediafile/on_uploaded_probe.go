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

type ProbeMediaOnUploadedHandler struct {
	probe *cmdmediafile.ProbeMediaHandler
}

func NewProbeMediaOnUploadedHandler(probe *cmdmediafile.ProbeMediaHandler) *ProbeMediaOnUploadedHandler {
	return &ProbeMediaOnUploadedHandler{probe: probe}
}

func (h *ProbeMediaOnUploadedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileUploaded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	mediaFileID, err := uuid.Parse(evt.MediaFileID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	log.Printf("probe worker received event type=%s mediaFileId=%s", evt.EventType(), evt.MediaFileID)
	if err := h.probe.Handle(ctx, cmdmediafile.ProbeMediaCommand{MediaFileID: mediaFileID}); err != nil {
		log.Printf("probe worker failed mediaFileId=%s: %v", evt.MediaFileID, err)
		return err
	}
	log.Printf("probe worker finished mediaFileId=%s", evt.MediaFileID)
	return nil
}
