package fixture

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go-api/internal/domain/port"
)

// Client loads a JSON dump (transcript + transcript_words) and companion source-time SRT.
type Client struct {
	jsonPath string
}

func NewClient(jsonPath string) *Client {
	return &Client{jsonPath: jsonPath}
}

func (c *Client) Transcribe(_ context.Context, _ string) (port.TranscriptResult, error) {
	dump, err := LoadDump(c.jsonPath)
	if err != nil {
		return port.TranscriptResult{}, err
	}

	words := make([]port.TranscriptWord, 0, len(dump.Words))
	for _, w := range dump.Words {
		word := port.TranscriptWord{
			Text:    w.Text,
			StartMs: w.SourceStartMs,
			EndMs:   w.SourceEndMs,
		}
		if w.Confidence != nil {
			conf := *w.Confidence
			word.Confidence = &conf
		}
		words = append(words, word)
	}

	if dump.SRTFile == "" {
		return port.TranscriptResult{}, fmt.Errorf("fixture dump missing srt_file")
	}
	srtPath := dump.SRTFile
	if !filepath.IsAbs(srtPath) {
		srtPath = filepath.Join(filepath.Dir(c.jsonPath), srtPath)
	}
	srtBytes, err := os.ReadFile(srtPath)
	if err != nil {
		return port.TranscriptResult{}, fmt.Errorf("read fixture srt %s: %w", srtPath, err)
	}
	if len(srtBytes) == 0 {
		return port.TranscriptResult{}, fmt.Errorf("fixture srt %s is empty", srtPath)
	}

	return port.TranscriptResult{
		ProviderJobID: dump.Transcript.ProviderJobID,
		Language:      dump.Transcript.Language,
		Text:          dump.Transcript.Text,
		SRT:           string(srtBytes),
		Words:         words,
	}, nil
}
