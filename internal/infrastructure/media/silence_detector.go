package media

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"go-api/internal/domain/port"
)

type SilenceDetector struct{}

func NewSilenceDetector() *SilenceDetector {
	return &SilenceDetector{}
}

var (
	silenceStartRe = regexp.MustCompile(`silence_start:\s*([0-9.]+)`)
	silenceEndRe   = regexp.MustCompile(`silence_end:\s*([0-9.]+)`)
)

func (d *SilenceDetector) Detect(
	ctx context.Context,
	audioPath string,
	thresholdDB float64,
	minDurationMs int64,
) ([]port.SilenceInterval, error) {
	if thresholdDB == 0 {
		thresholdDB = -35
	}
	minDurationSec := float64(minDurationMs) / 1000
	if minDurationSec <= 0 {
		minDurationSec = 0.4
	}

	filter := fmt.Sprintf("silencedetect=noise=%.1fdB:d=%.3f", thresholdDB, minDurationSec)
	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-hide_banner",
		"-i", audioPath,
		"-af", filter,
		"-f", "null",
		"-",
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	_ = cmd.Run() // silencedetect logs to stderr; non-zero exit is common for null mux

	return parseSilenceLog(stderr.String()), nil
}

func parseSilenceLog(log string) []port.SilenceInterval {
	var out []port.SilenceInterval
	var currentStart *float64

	for _, line := range strings.Split(log, "\n") {
		if match := silenceStartRe.FindStringSubmatch(line); len(match) == 2 {
			if sec, err := strconv.ParseFloat(match[1], 64); err == nil {
				v := sec
				currentStart = &v
			}
			continue
		}
		if match := silenceEndRe.FindStringSubmatch(line); len(match) == 2 && currentStart != nil {
			if endSec, err := strconv.ParseFloat(match[1], 64); err == nil {
				startMs := int64(*currentStart * 1000)
				endMs := int64(endSec * 1000)
				if endMs > startMs {
					out = append(out, port.SilenceInterval{StartMs: startMs, EndMs: endMs})
				}
			}
			currentStart = nil
		}
	}
	return out
}
