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

type CreateSegmentsOnReadyHandler struct {
	create *cmdmediafile.CreateSegmentsHandler
}

func NewCreateSegmentsOnReadyHandler(create *cmdmediafile.CreateSegmentsHandler) *CreateSegmentsOnReadyHandler {
	return &CreateSegmentsOnReadyHandler{create: create}
}

func (h *CreateSegmentsOnReadyHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileReady
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	mediaFileID, err := uuid.Parse(evt.MediaFileID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	log.Printf("create segments received event type=%s mediaFileId=%s", evt.EventType(), evt.MediaFileID)
	if err := h.create.Handle(ctx, cmdmediafile.CreateSegmentsCommand{MediaFileID: mediaFileID}); err != nil {
		log.Printf("create segments failed mediaFileId=%s: %v", evt.MediaFileID, err)
		return err
	}
	log.Printf("create segments finished mediaFileId=%s", evt.MediaFileID)
	return nil
}
