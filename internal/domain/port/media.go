package port

import "context"

type FrameExtractor interface {
	ExtractThumbnail(ctx context.Context, videoPath string) ([]byte, error)
}

// MediaProbeResult is the technical fingerprint of a media file (ffprobe).
type MediaProbeResult struct {
	DurationMs      int64
	Width           *int
	Height          *int
	FPS             *float64
	VideoCodec      string
	AudioCodec      string
	AudioSampleRate *int
	AudioChannels   *int
	SizeBytes       int64
}

type MediaProber interface {
	Probe(ctx context.Context, videoPath string) (MediaProbeResult, error)
}

type AudioExtractor interface {
	Extract(ctx context.Context, videoPath, outputPath string) error
}
