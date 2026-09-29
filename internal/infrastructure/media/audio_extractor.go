package media

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type AudioExtractor struct{}

func NewAudioExtractor() *AudioExtractor {
	return &AudioExtractor{}
}

// Extract writes a mono 16 kHz PCM WAV suitable for silence detection and ASR.
func (e *AudioExtractor) Extract(ctx context.Context, videoPath, outputPath string) error {
	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-y",
		"-i", videoPath,
		"-vn",
		"-ac", "1",
		"-ar", "16000",
		"-c:a", "pcm_s16le",
		outputPath,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg audio extract failed: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}
