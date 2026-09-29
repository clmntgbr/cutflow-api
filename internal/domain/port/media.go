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
	// SRT is the source-time caption file; always produced and stored. Not used for final render ASS.
	SRT   string
	Words []TranscriptWord
}

type SpeechTranscriber interface {
	Transcribe(ctx context.Context, audioPath string) (TranscriptResult, error)
}

// ViralLLMProposal is a raw LLM suggestion before boundary resolve / validation.
type ViralLLMProposal struct {
	StartMs         int64
	EndMs           int64
	Score           float64
	HookScore       *float64
	StandaloneScore *float64
	PayoffScore     *float64
	InterestScore   *float64
	Title           string
	Hook            string
	Reason          string
}

type ViralAnalyzeInput struct {
	TranscriptText string
	Language       string
	MinDurationMs  int64
	MaxDurationMs  int64
	MaxCandidates  int
}

// ViralAnalyzer calls an LLM provider to propose short-form clip windows.
type ViralAnalyzer interface {
	Analyze(ctx context.Context, input ViralAnalyzeInput) ([]ViralLLMProposal, error)
	Provider() string
	Model() string
}
