package media

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	domainmediaaudio "go-api/internal/domain/mediaaudio"
)

type AudioExtractor struct{}

func NewAudioExtractor() *AudioExtractor {
	return &AudioExtractor{}
}

// Extract writes a mono 16 kHz Opus (~32 kbps) file for ASR and silence detection.
func (e *AudioExtractor) Extract(ctx context.Context, videoPath, outputPath string) error {
	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-y",
		"-i", videoPath,
		"-vn",
		"-ac", "1",
		"-ar", "16000",
		"-c:a", "libopus",
		"-b:a", domainmediaaudio.ExtractedBitrate,
		outputPath,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg audio extract failed: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}
