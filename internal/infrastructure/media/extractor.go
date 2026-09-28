package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type FrameExtractor struct{}

func NewFrameExtractor() *FrameExtractor {
	return &FrameExtractor{}
}

func (e *FrameExtractor) ExtractThumbnail(ctx context.Context, videoPath string) ([]byte, error) {
	tmp, err := os.CreateTemp("", "media-thumb-*.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp thumbnail: %w", err)
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	defer os.Remove(tmpPath)

	var lastErr error
	for _, seek := range []string{"1", "0"} {
		cmd := exec.CommandContext(
			ctx,
			"ffmpeg",
			"-y",
			"-ss", seek,
			"-i", videoPath,
			"-frames:v", "1",
			"-update", "1",
			"-vf", "scale='min(640,iw)':-2",
			"-q:v", "3",
			tmpPath,
		)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			lastErr = fmt.Errorf("ffmpeg thumbnail failed: %w (%s)", err, strings.TrimSpace(stderr.String()))
			continue
		}

		data, err := os.ReadFile(tmpPath)
		if err != nil {
			lastErr = err
			continue
		}
		if len(data) == 0 {
			lastErr = errors.New("ffmpeg produced an empty thumbnail")
			continue
		}
		return data, nil
	}

	if lastErr == nil {
		lastErr = errors.New("ffmpeg produced no thumbnail")
	}
	return nil, lastErr
}
