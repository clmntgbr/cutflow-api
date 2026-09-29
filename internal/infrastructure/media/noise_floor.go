package media

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

type NoiseFloorAnalyzer struct{}

func NewNoiseFloorAnalyzer() *NoiseFloorAnalyzer {
	return &NoiseFloorAnalyzer{}
}

var meanVolumeRe = regexp.MustCompile(`mean_volume:\s*(-?[0-9.]+)\s*dB`)

// Analyze estimates a noise-floor proxy via ffmpeg volumedetect mean_volume.
// Temporary calibration: whole-file mean, not a quiet-percentile floor.
func (a *NoiseFloorAnalyzer) Analyze(ctx context.Context, audioPath string) (float64, error) {
	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-hide_banner",
		"-i", audioPath,
		"-af", "volumedetect",
		"-f", "null",
		"-",
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	_ = cmd.Run()

	match := meanVolumeRe.FindStringSubmatch(stderr.String())
	if len(match) != 2 {
		return 0, fmt.Errorf("volumedetect mean_volume not found")
	}
	value, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, fmt.Errorf("parse mean_volume: %w", err)
	}
	if value > 0 || value < -100 {
		return 0, fmt.Errorf("unexpected mean_volume %.1f", value)
	}
	return value, nil
}
