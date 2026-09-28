package port

import "context"

type FrameExtractor interface {
	ExtractThumbnail(ctx context.Context, videoPath string) ([]byte, error)
}
