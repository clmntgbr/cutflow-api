package handler

import (
	"context"

	cmdmediafile "go-api/internal/application/command/mediafile"
)

type storageConfirmUploadHandler interface {
	Handle(ctx context.Context, cmd cmdmediafile.ConfirmUploadCommand) error
}
