package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/persistence/write"
	"go-api/internal/infrastructure/storage"
	"go-api/internal/infrastructure/transcriptfixture"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func NewExportTranscriptCommand() *cobra.Command {
	var mediaFileID string
	var outPath string

	cmd := &cobra.Command{
		Use:   "export-transcript",
		Short: "Export transcript + transcript_word rows to a JSON fixture",
		Long:  "Writes a portable JSON dump (no media/project/user IDs) plus the companion source-time SRT next to it.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if mediaFileID == "" {
				return fmt.Errorf("--media-file-id is required")
			}
			id, err := uuid.Parse(mediaFileID)
			if err != nil {
				return fmt.Errorf("invalid media-file-id: %w", err)
			}
			if outPath == "" {
				outPath = "fixtures/transcript.json"
			}

			env := config.Load()
			db := config.ConnectDatabase(env)
			repo := write.NewTranscriptWriteRepository(db)

			transcript, err := repo.GetByMediaFileID(context.Background(), id)
			if err != nil {
				return err
			}
			if transcript == nil {
				return fmt.Errorf("transcript not found for media_file_id=%s", mediaFileID)
			}
			if transcript.SRTStorageKey == "" {
				return fmt.Errorf("transcript has no SRT storage key")
			}
			wordRows, err := repo.ListWords(context.Background(), transcript.ID)
			if err != nil {
				return fmt.Errorf("load words: %w", err)
			}

			words := make([]fixture.WordRow, 0, len(wordRows))
			for _, w := range wordRows {
				words = append(words, fixture.WordRow{
					WordIndex:     w.WordIndex,
					Text:          w.Text,
					SourceStartMs: w.SourceStartMs,
					SourceEndMs:   w.SourceEndMs,
					Confidence:    w.Confidence,
					Kind:          w.Kind,
				})
			}

			dump := &fixture.Dump{
				Transcript: fixture.TranscriptRow{
					Language:      transcript.Language,
					Text:          transcript.Text,
					ProviderJobID: transcript.ProviderJobID,
					Status:        transcript.Status,
				},
				Words:   words,
				SRTFile: "subtitles.srt",
			}

			if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil && filepath.Dir(outPath) != "." {
				return err
			}
			if err := fixture.WriteDump(outPath, dump); err != nil {
				return err
			}

			store, err := storage.NewMinIOStorage(env)
			if err != nil {
				return fmt.Errorf("storage: %w", err)
			}
			baseDir := filepath.Dir(outPath)
			if err := copyStorageObject(context.Background(), store, transcript.SRTStorageKey, filepath.Join(baseDir, dump.SRTFile)); err != nil {
				return fmt.Errorf("copy SRT: %w", err)
			}

			fmt.Printf("Exported transcript fixture to %s (%d words)\n", outPath, len(words))
			return nil
		},
	}

	cmd.Flags().StringVar(&mediaFileID, "media-file-id", "", "Media file UUID")
	cmd.Flags().StringVar(&outPath, "out", "fixtures/transcript.json", "Output JSON path")
	return cmd
}

func NewBuildTranscriptFixtureCommand() *cobra.Command {
	var srtPath string
	var outPath string
	var language string

	cmd := &cobra.Command{
		Use:   "build-transcript-fixture",
		Short: "Build transcript JSON fixture from local SRT (source-time words)",
		RunE: func(cmd *cobra.Command, args []string) error {
			srtBytes, err := os.ReadFile(srtPath)
			if err != nil {
				return err
			}
			if len(srtBytes) == 0 {
				return fmt.Errorf("SRT file is empty: %s", srtPath)
			}
			dump := fixture.BuildDumpFromSRT(string(srtBytes), language)
			dump.SRTFile = filepath.Base(srtPath)

			if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil && filepath.Dir(outPath) != "." {
				return err
			}
			if err := fixture.WriteDump(outPath, dump); err != nil {
				return err
			}

			info, _ := os.Stat(outPath)
			raw, _ := json.Marshal(map[string]any{
				"out":   outPath,
				"words": len(dump.Words),
				"bytes": info.Size(),
			})
			fmt.Println(string(raw))
			return nil
		},
	}

	cmd.Flags().StringVar(&srtPath, "srt", "subtitles.srt", "Source SRT path")
	cmd.Flags().StringVar(&outPath, "out", "fixtures/transcript.json", "Output JSON path")
	cmd.Flags().StringVar(&language, "language", "fr", "Language code")
	return cmd
}

func copyStorageObject(ctx context.Context, store *storage.MinIOStorage, key, dest string) error {
	if key == "" {
		return nil
	}
	reader, err := store.Get(ctx, key)
	if err != nil {
		return err
	}
	defer reader.Close()
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.ReadFrom(reader)
	return err
}
