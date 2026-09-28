package handler

import (
	"context"
	"io"

	querymediafile "go-api/internal/application/query/mediafile"
	domainmediafile "go-api/internal/domain/mediafile"
)

type mediaFileGetOwnedHandler interface {
	Handle(ctx context.Context, q querymediafile.GetOwnedMediaFileQuery) (*domainmediafile.MediaFileThumbnailView, error)
}

type mediaFileThumbnailStorage interface {
	Get(ctx context.Context, key string) (io.ReadCloser, error)
}
