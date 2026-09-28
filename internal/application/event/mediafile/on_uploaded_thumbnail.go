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

type GenerateThumbnailOnUploadedHandler struct {
	generate *cmdmediafile.GenerateThumbnailHandler
}

func NewGenerateThumbnailOnUploadedHandler(
	generate *cmdmediafile.GenerateThumbnailHandler,
) *GenerateThumbnailOnUploadedHandler {
	return &GenerateThumbnailOnUploadedHandler{generate: generate}
}

func (h *GenerateThumbnailOnUploadedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainmediafile.MediaFileUploaded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	mediaFileID, err := uuid.Parse(evt.MediaFileID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	log.Printf("thumbnail worker received event type=%s mediaFileId=%s", evt.EventType(), evt.MediaFileID)
	if err := h.generate.Handle(ctx, cmdmediafile.GenerateThumbnailCommand{MediaFileID: mediaFileID}); err != nil {
		log.Printf("thumbnail worker failed mediaFileId=%s: %v", evt.MediaFileID, err)
		return err
	}
	log.Printf("thumbnail worker finished mediaFileId=%s", evt.MediaFileID)
	return nil
}
