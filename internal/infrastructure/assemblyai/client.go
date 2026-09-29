package assemblyai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go-api/internal/domain/port"
	"go-api/internal/infrastructure/subtitle"
)

const (
	uploadURL     = "https://api.assemblyai.com/v2/upload"
	transcriptURL = "https://api.assemblyai.com/v2/transcript"
	pollInterval  = 3 * time.Second
	charsPerCaption = 32
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

func (c *Client) Transcribe(ctx context.Context, audioPath string) (port.TranscriptResult, error) {
	if c.apiKey == "" {
		return port.TranscriptResult{}, fmt.Errorf("ASSEMBLYAI_API_KEY is not set")
	}

	uploadURLValue, err := c.upload(ctx, audioPath)
	if err != nil {
		return port.TranscriptResult{}, err
	}
	jobID, err := c.start(ctx, uploadURLValue)
	if err != nil {
		return port.TranscriptResult{}, err
	}
	detail, err := c.poll(ctx, jobID)
	if err != nil {
		return port.TranscriptResult{}, err
	}
	srt, err := c.fetchSRT(ctx, jobID)
	if err != nil {
		return port.TranscriptResult{}, err
	}

	words := make([]port.TranscriptWord, 0, len(detail.Words))
	for _, word := range detail.Words {
		w := port.TranscriptWord{
			Text:    word.Text,
			StartMs: word.Start,
			EndMs:   word.End,
		}
		if word.Confidence > 0 {
			conf := word.Confidence
			w.Confidence = &conf
		}
		words = append(words, w)
	}

	return port.TranscriptResult{
		ProviderJobID: jobID,
		Language:      detail.LanguageCode,
		Text:          detail.Text,
		SRT:           srt,
		ASS:           subtitle.SRTToASS(srt),
		Words:         words,
	}, nil
}

func (c *Client) upload(ctx context.Context, audioPath string) (string, error) {
	file, err := os.Open(audioPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, file)
	if err != nil {
		return "", err
	}
	req.Header.Set("authorization", c.apiKey)
	req.Header.Set("content-type", "application/octet-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("assemblyai upload failed: %s", string(body))
	}
	var payload struct {
		UploadURL string `json:"upload_url"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	if payload.UploadURL == "" {
		return "", fmt.Errorf("assemblyai upload missing upload_url")
	}
	return payload.UploadURL, nil
}

func (c *Client) start(ctx context.Context, audioURL string) (string, error) {
	payload, _ := json.Marshal(map[string]any{
		"audio_url":          audioURL,
		"language_detection": true,
		"punctuate":          true,
		"format_text":        true,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, transcriptURL, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("authorization", c.apiKey)
	req.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("assemblyai start failed: %s", string(body))
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if out.ID == "" {
		return "", fmt.Errorf("assemblyai start missing id")
	}
	return out.ID, nil
}

type transcriptDetail struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Text         string `json:"text"`
	LanguageCode string `json:"language_code"`
	Error        string `json:"error"`
	Words        []struct {
		Text       string  `json:"text"`
		Start      int64   `json:"start"`
		End        int64   `json:"end"`
		Confidence float64 `json:"confidence"`
	} `json:"words"`
}

func (c *Client) poll(ctx context.Context, jobID string) (*transcriptDetail, error) {
	url := transcriptURL + "/" + jobID
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("authorization", c.apiKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode >= 300 {
			return nil, fmt.Errorf("assemblyai poll failed: %s", string(body))
		}

		var detail transcriptDetail
		if err := json.Unmarshal(body, &detail); err != nil {
			return nil, err
		}
		switch detail.Status {
		case "completed":
			return &detail, nil
		case "error":
			if detail.Error == "" {
				detail.Error = "unknown transcription error"
			}
			return nil, fmt.Errorf("assemblyai transcription failed: %s", detail.Error)
		}
	}
}

func (c *Client) fetchSRT(ctx context.Context, jobID string) (string, error) {
	url := fmt.Sprintf("%s/%s/srt?chars_per_caption=%d", transcriptURL, jobID, charsPerCaption)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("authorization", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusBadRequest && bytes.Contains(body, []byte("empty")) {
			return "", nil
		}
		return "", fmt.Errorf("assemblyai srt failed: %s", string(body))
	}
	return string(body), nil
}
