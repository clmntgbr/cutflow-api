package handler

import (
	"context"
	"io"

	cmdmediafile "go-api/internal/application/command/mediafile"
	querymediafile "go-api/internal/application/query/mediafile"
	domainmediafile "go-api/internal/domain/mediafile"
)

type mediaFileGetOwnedHandler interface {
	Handle(ctx context.Context, q querymediafile.GetOwnedMediaFileQuery) (*domainmediafile.MediaFileThumbnailView, error)
}

type mediaFileGetEditorHandler interface {
	Handle(ctx context.Context, q querymediafile.GetEditorStateQuery) (*querymediafile.EditorStateView, error)
}

type mediaFileUpdateConfigurationHandler interface {
	Handle(ctx context.Context, cmd cmdmediafile.UpdateEditorConfigurationCommand) (*cmdmediafile.EditorMutationResult, error)
}

type mediaFileDecisionMutationHandler interface {
	Override(ctx context.Context, cmd cmdmediafile.OverrideDecisionCommand) (*cmdmediafile.EditorMutationResult, error)
	ClearOverride(ctx context.Context, cmd cmdmediafile.ClearDecisionOverrideCommand) (*cmdmediafile.EditorMutationResult, error)
	CreateManual(ctx context.Context, cmd cmdmediafile.CreateManualDecisionCommand) (*cmdmediafile.EditorMutationResult, error)
}

type mediaFileFinalizeHandler interface {
	Handle(ctx context.Context, cmd cmdmediafile.FinalizeEditorCommand) (*cmdmediafile.FinalizeEditorResult, error)
}

type mediaFileThumbnailStorage interface {
	Get(ctx context.Context, key string) (io.ReadCloser, error)
}
