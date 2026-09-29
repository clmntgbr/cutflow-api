package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"go-api/internal/domain/port"
)

type MediaProber struct{}

func NewMediaProber() *MediaProber {
	return &MediaProber{}
}

type ffprobeOutput struct {
	Format struct {
		Duration string `json:"duration"`
		Size     string `json:"size"`
	} `json:"format"`
	Streams []ffprobeStream `json:"streams"`
}

type ffprobeStream struct {
	CodecType    string `json:"codec_type"`
	CodecName    string `json:"codec_name"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	AvgFrameRate string `json:"avg_frame_rate"`
	RFrameRate   string `json:"r_frame_rate"`
	SampleRate   string `json:"sample_rate"`
	Channels     int    `json:"channels"`
}

func (p *MediaProber) Probe(ctx context.Context, videoPath string) (port.MediaProbeResult, error) {
	cmd := exec.CommandContext(
		ctx,
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoPath,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return port.MediaProbeResult{}, fmt.Errorf(
			"ffprobe failed: %w (%s)",
			err,
			strings.TrimSpace(stderr.String()),
		)
	}

	var raw ffprobeOutput
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		return port.MediaProbeResult{}, fmt.Errorf("ffprobe json: %w", err)
	}

	result := port.MediaProbeResult{}
	if raw.Format.Duration != "" {
		seconds, err := strconv.ParseFloat(raw.Format.Duration, 64)
		if err != nil {
			return port.MediaProbeResult{}, fmt.Errorf("parse duration %q: %w", raw.Format.Duration, err)
		}
		if seconds < 0 {
			seconds = 0
		}
		result.DurationMs = int64(seconds * 1000)
	}
	if raw.Format.Size != "" {
		size, err := strconv.ParseInt(raw.Format.Size, 10, 64)
		if err == nil && size > 0 {
			result.SizeBytes = size
		}
	}

	for _, stream := range raw.Streams {
		switch stream.CodecType {
		case "video":
			if result.VideoCodec != "" {
				continue
			}
			result.VideoCodec = stream.CodecName
			if stream.Width > 0 {
				w := stream.Width
				result.Width = &w
			}
			if stream.Height > 0 {
				h := stream.Height
				result.Height = &h
			}
			if fps := parseFrameRate(stream.AvgFrameRate); fps != nil {
				result.FPS = fps
			} else if fps := parseFrameRate(stream.RFrameRate); fps != nil {
				result.FPS = fps
			}
		case "audio":
			if result.AudioCodec != "" {
				continue
			}
			result.AudioCodec = stream.CodecName
			if stream.SampleRate != "" {
				if rate, err := strconv.Atoi(stream.SampleRate); err == nil && rate > 0 {
					result.AudioSampleRate = &rate
				}
			}
			if stream.Channels > 0 {
				ch := stream.Channels
				result.AudioChannels = &ch
			}
		}
	}

	return result, nil
}

func parseFrameRate(raw string) *float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "0/0" {
		return nil
	}
	parts := strings.SplitN(raw, "/", 2)
	if len(parts) == 1 {
		value, err := strconv.ParseFloat(parts[0], 64)
		if err != nil || value <= 0 {
			return nil
		}
		return &value
	}
	num, errNum := strconv.ParseFloat(parts[0], 64)
	den, errDen := strconv.ParseFloat(parts[1], 64)
	if errNum != nil || errDen != nil || den == 0 {
		return nil
	}
	value := num / den
	if value <= 0 {
		return nil
	}
	return &value
}
