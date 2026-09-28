package handler

import (
	"context"

	cmdproject "go-api/internal/application/command/project"
)

type projectRequestUploadURLHandler interface {
	Handle(ctx context.Context, cmd cmdproject.RequestUploadURLCommand) (*cmdproject.RequestUploadURLResult, error)
}
