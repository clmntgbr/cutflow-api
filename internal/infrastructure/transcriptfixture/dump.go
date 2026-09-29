package fixture

import (
	"encoding/json"
	"fmt"
	"os"
)

// Dump is a portable export of transcript + transcript_word rows (without runtime IDs).
type Dump struct {
	Transcript TranscriptRow `json:"transcript"`
	Words      []WordRow     `json:"transcript_words"`
	// Relative paths next to the JSON file (optional; defaults to subtitles.srt / subtitles.ass).
	SRTFile string `json:"srt_file,omitempty"`
	ASSFile string `json:"ass_file,omitempty"`
}

type TranscriptRow struct {
	Language      string `json:"language"`
	Text          string `json:"text"`
	ProviderJobID string `json:"provider_job_id"`
	Status        string `json:"status"`
}

type WordRow struct {
	WordIndex     int      `json:"word_index"`
	Text          string   `json:"text"`
	SourceStartMs int64    `json:"source_start_ms"`
	SourceEndMs   int64    `json:"source_end_ms"`
	Confidence    *float64 `json:"confidence"`
	Kind          string   `json:"kind"`
}

func LoadDump(path string) (*Dump, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read fixture json %s: %w", path, err)
	}
	var dump Dump
	if err := json.Unmarshal(raw, &dump); err != nil {
		return nil, fmt.Errorf("parse fixture json %s: %w", path, err)
	}
	if dump.Transcript.Status == "" {
		dump.Transcript.Status = "completed"
	}
	if dump.Transcript.ProviderJobID == "" {
		dump.Transcript.ProviderJobID = "fixture"
	}
	if dump.SRTFile == "" {
		dump.SRTFile = "subtitles.srt"
	}
	if dump.ASSFile == "" {
		dump.ASSFile = "subtitles.ass"
	}
	return &dump, nil
}

func WriteDump(path string, dump *Dump) error {
	raw, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o644)
}

// BuildDumpFromSRT builds a re-insertable dump from an SRT file.
func BuildDumpFromSRT(srt string, language string) *Dump {
	words, text := wordsFromSRT(srt)
	rows := make([]WordRow, 0, len(words))
	for i, w := range words {
		rows = append(rows, WordRow{
			WordIndex:     i,
			Text:          w.Text,
			SourceStartMs: w.StartMs,
			SourceEndMs:   w.EndMs,
			Kind:          "speech",
		})
	}
	return &Dump{
		Transcript: TranscriptRow{
			Language:      language,
			Text:          text,
			ProviderJobID: "fixture",
			Status:        "completed",
		},
		Words:   rows,
		SRTFile: "subtitles.srt",
		ASSFile: "subtitles.ass",
	}
}
