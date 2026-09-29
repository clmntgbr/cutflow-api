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

type SilenceInterval struct {
	StartMs int64
	EndMs   int64
}

type SilenceDetector interface {
	Detect(ctx context.Context, audioPath string, thresholdDB float64, minDurationMs int64) ([]SilenceInterval, error)
}

// NoiseFloorAnalyzer estimates background noise level in dBFS for auto silence threshold.
type NoiseFloorAnalyzer interface {
	Analyze(ctx context.Context, audioPath string) (float64, error)
}

type TranscriptWord struct {
	Text      string
	StartMs   int64
	EndMs     int64
	Confidence *float64
}

type TranscriptResult struct {
	ProviderJobID string
	Language      string
	Text          string
	SRT           string
	// ASS, when non-empty, is uploaded as-is (fixture / pre-rendered). Otherwise the handler generates it.
	ASS   string
	Words []TranscriptWord
}

type SpeechTranscriber interface {
	Transcribe(ctx context.Context, audioPath string) (TranscriptResult, error)
}
