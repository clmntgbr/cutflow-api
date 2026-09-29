package fixture

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"go-api/internal/domain/port"
)

// Dump matches fixtures/viral_llm.json (recorded or hand-authored LLM responses).
type Dump struct {
	Provider  string          `json:"provider"`
	Model     string          `json:"model"`
	Responses []ResponseBlock `json:"responses"`
}

type ResponseBlock struct {
	// Schema A (expected structured output)
	Candidates []Candidate `json:"candidates"`
	// Schema B (actual gpt-4o-mini freeform from first dumps)
	Clips []Clip `json:"clips"`
}

type Candidate struct {
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

type Clip struct {
	Start       string `json:"start"`
	End         string `json:"end"`
	Content     string `json:"content"`
	Description string `json:"description"`
}

// Client loads a viral LLM fixture and serves all proposals once (no live API call).
type Client struct {
	path      string
	provider  string
	model     string
	proposals []port.ViralLLMProposal
	served    atomic.Bool
}

func NewClient(path string) (*Client, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read viral fixture %s: %w", path, err)
	}
	var dump Dump
	if err := json.Unmarshal(raw, &dump); err != nil {
		return nil, fmt.Errorf("parse viral fixture %s: %w", path, err)
	}

	proposals := make([]port.ViralLLMProposal, 0)
	for _, block := range dump.Responses {
		proposals = append(proposals, proposalsFromBlock(block)...)
	}
	if len(proposals) == 0 {
		return nil, fmt.Errorf("viral fixture %s has no candidates/clips", path)
	}

	provider := dump.Provider
	if provider == "" {
		provider = "fixture"
	}
	model := dump.Model
	if model == "" {
		model = "fixture"
	}

	return &Client{
		path:      path,
		provider:  provider,
		model:     model,
		proposals: proposals,
	}, nil
}

func (c *Client) Provider() string { return c.provider }
func (c *Client) Model() string    { return c.model }

func (c *Client) Analyze(_ context.Context, _ port.ViralAnalyzeInput) ([]port.ViralLLMProposal, error) {
	// Serve the full fixture once; subsequent chunk calls get nothing (avoids duplicates).
	if c.served.Swap(true) {
		return nil, nil
	}
	out := make([]port.ViralLLMProposal, len(c.proposals))
	copy(out, c.proposals)
	return out, nil
}

func proposalsFromBlock(block ResponseBlock) []port.ViralLLMProposal {
	out := make([]port.ViralLLMProposal, 0, len(block.Candidates)+len(block.Clips))
	for _, c := range block.Candidates {
		start, err1 := toInt64(c.StartMs)
		end, err2 := toInt64(c.EndMs)
		if err1 != nil || err2 != nil || end <= start {
			continue
		}
		score := c.Score
		if score <= 0 {
			score = 0.8
		}
		out = append(out, port.ViralLLMProposal{
			StartMs:         start,
			EndMs:           end,
			Score:           score,
			HookScore:       c.HookScore,
			StandaloneScore: c.StandaloneScore,
			PayoffScore:     c.PayoffScore,
			InterestScore:   c.InterestScore,
			Title:           c.Title,
			Hook:            c.Hook,
			Reason:          c.Reason,
		})
	}
	for _, clip := range block.Clips {
		start, err1 := parseClockMs(clip.Start)
		end, err2 := parseClockMs(clip.End)
		if err1 != nil || err2 != nil || end <= start {
			continue
		}
		hook := clip.Content
		if hook == "" {
			hook = clip.Description
		}
		title := truncateRunes(hook, 80)
		out = append(out, port.ViralLLMProposal{
			StartMs: start,
			EndMs:   end,
			Score:   0.8,
			Title:   title,
			Hook:    hook,
			Reason:  "fixture",
		})
	}
	return out
}

func parseClockMs(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("empty clock")
	}
	parts := strings.Split(raw, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, fmt.Errorf("invalid clock %q", raw)
	}
	var hours, minutes int64
	var seconds float64
	var err error
	if len(parts) == 3 {
		hours, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, err
		}
		minutes, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, err
		}
		seconds, err = strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return 0, err
		}
	} else {
		minutes, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, err
		}
		seconds, err = strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, err
		}
	}
	ms := hours*3_600_000 + minutes*60_000 + int64(seconds*1000)
	return ms, nil
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
		return t.Int64()
	case string:
		return parseClockMs(t)
	default:
		return 0, fmt.Errorf("unsupported number type %T", v)
	}
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
