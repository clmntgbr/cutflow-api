package fixture

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go-api/internal/domain/port"
)

// Client loads a JSON dump (transcript + transcript_words) and companion SRT/ASS files.
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

	baseDir := filepath.Dir(c.jsonPath)
	srtPath := dump.SRTFile
	assPath := dump.ASSFile
	if !filepath.IsAbs(srtPath) {
		srtPath = filepath.Join(baseDir, srtPath)
	}
	if !filepath.IsAbs(assPath) {
		assPath = filepath.Join(baseDir, assPath)
	}

	srtBytes, err := os.ReadFile(srtPath)
	if err != nil {
		return port.TranscriptResult{}, fmt.Errorf("read fixture srt %s: %w", srtPath, err)
	}
	assBytes, err := os.ReadFile(assPath)
	if err != nil {
		return port.TranscriptResult{}, fmt.Errorf("read fixture ass %s: %w", assPath, err)
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

	return port.TranscriptResult{
		ProviderJobID: dump.Transcript.ProviderJobID,
		Language:      dump.Transcript.Language,
		Text:          dump.Transcript.Text,
		SRT:           string(srtBytes),
		ASS:           string(assBytes),
		Words:         words,
	}, nil
}
