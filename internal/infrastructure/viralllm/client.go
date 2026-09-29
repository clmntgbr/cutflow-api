package viralllm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go-api/internal/domain/port"
)

const systemPrompt = `You are analyzing the timestamped transcript of a video.

Identify the strongest self-contained passages that could work as
short-form social media clips (TikTok, Reels, Shorts).

A strong clip should:
- have a compelling opening or hook;
- make sense without the rest of the video;
- contain useful, surprising, emotional or entertaining content;
- have a clear progression;
- end with a conclusion, payoff or natural stopping point.

Prefer complete ideas over arbitrary fixed durations.
Do not invent content that does not appear in the transcript.
All timestamps must correspond to the provided transcript.
Return only JSON matching the schema.`

// FixtureDump stores exact LLM message contents for later replay.
type FixtureDump struct {
	Provider  string            `json:"provider"`
	Model     string            `json:"model"`
	Responses []json.RawMessage `json:"responses"`
}

type Client struct {
	provider   string
	model      string
	apiKey     string
	baseURL    string
	dumpPath   string
	httpClient *http.Client

	mu    sync.Mutex
	dumps []json.RawMessage
}

func New(provider, model, apiKey, baseURL, dumpPath string) *Client {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if baseURL == "" {
		switch provider {
		case "deepseek":
			baseURL = "https://api.deepseek.com/v1"
		default:
			provider = "openai"
			baseURL = "https://api.openai.com/v1"
		}
	}
	if model == "" {
		switch provider {
		case "deepseek":
			model = "deepseek-chat"
		default:
			model = "gpt-4o-mini"
		}
	}
	return &Client{
		provider: provider,
		model:    model,
		apiKey:   apiKey,
		baseURL:  strings.TrimRight(baseURL, "/"),
		dumpPath: strings.TrimSpace(dumpPath),
		httpClient: &http.Client{
			Timeout: 3 * time.Minute,
		},
	}
}

func (c *Client) Provider() string { return c.provider }
func (c *Client) Model() string    { return c.model }

func (c *Client) Analyze(ctx context.Context, input port.ViralAnalyzeInput) ([]port.ViralLLMProposal, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("viral LLM API key is not set")
	}
	if strings.TrimSpace(input.TranscriptText) == "" {
		return nil, nil
	}

	user := fmt.Sprintf(
		"Minimum duration: %d seconds\nMaximum duration: %d seconds\nMaximum candidates: %d\nLanguage: %s\n\nTRANSCRIPT:\n\n%s",
		input.MinDurationMs/1000,
		input.MaxDurationMs/1000,
		input.MaxCandidates,
		input.Language,
		input.TranscriptText,
	)

	payload := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": user},
		},
		"temperature": 0.2,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s chat completions failed (%d): %s", c.provider, resp.StatusCode, truncate(string(raw), 500))
	}

	var envelope struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("parse llm envelope: %w", err)
	}
	if len(envelope.Choices) == 0 {
		return nil, fmt.Errorf("llm returned no choices")
	}

	content := envelope.Choices[0].Message.Content
	if err := c.dumpResponse(content); err != nil {
		log.Printf("viral llm fixture dump failed path=%s: %v", c.dumpPath, err)
	}

	return parseProposals(content)
}

func (c *Client) dumpResponse(content string) error {
	if c.dumpPath == "" {
		return nil
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	raw := json.RawMessage(content)
	if !json.Valid(raw) {
		wrapped, err := json.Marshal(map[string]string{"raw": content})
		if err != nil {
			return err
		}
		raw = wrapped
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.dumps = append(c.dumps, raw)

	dump := FixtureDump{
		Provider:  c.provider,
		Model:     c.model,
		Responses: append([]json.RawMessage(nil), c.dumps...),
	}
	pretty, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		return err
	}
	pretty = append(pretty, '\n')

	if dir := filepath.Dir(c.dumpPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	if err := os.WriteFile(c.dumpPath, pretty, 0o644); err != nil {
		return err
	}
	log.Printf("viral llm fixture dumped path=%s responses=%d", c.dumpPath, len(c.dumps))
	return nil
}

type llmResponse struct {
	Candidates []llmCandidate `json:"candidates"`
}

type llmCandidate struct {
	StartMs         any      `json:"start_ms"`
	EndMs           any      `json:"end_ms"`
	Score           float64  `json:"score"`
	HookScore       *float64 `json:"hook_score"`
	StandaloneScore *float64 `json:"standalone_score"`
	PayoffScore     *float64 `json:"payoff_score"`
	InterestScore   *float64 `json:"interest_score"`
	Title           string   `json:"title"`
	Hook            string   `json:"hook"`
	Reason          string   `json:"reason"`
}

func parseProposals(content string) ([]port.ViralLLMProposal, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, nil
	}
	var parsed llmResponse
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("parse llm json: %w", err)
	}
	out := make([]port.ViralLLMProposal, 0, len(parsed.Candidates))
	for _, c := range parsed.Candidates {
		start, err1 := toInt64(c.StartMs)
		end, err2 := toInt64(c.EndMs)
		if err1 != nil || err2 != nil {
			continue
		}
		out = append(out, port.ViralLLMProposal{
			StartMs:         start,
			EndMs:           end,
			Score:           c.Score,
			HookScore:       c.HookScore,
			StandaloneScore: c.StandaloneScore,
			PayoffScore:     c.PayoffScore,
			InterestScore:   c.InterestScore,
			Title:           c.Title,
			Hook:            c.Hook,
			Reason:          c.Reason,
		})
	}
	return out, nil
}

func toInt64(v any) (int64, error) {
	switch t := v.(type) {
	case float64:
		return int64(t), nil
	case int64:
		return t, nil
	case int:
		return int64(t), nil
	case json.Number:
		i, err := t.Int64()
		return i, err
	case string:
		var f float64
		if _, err := fmt.Sscanf(strings.TrimSpace(t), "%f", &f); err != nil {
			return 0, err
		}
		return int64(f), nil
	default:
		return 0, fmt.Errorf("unsupported number type %T", v)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
