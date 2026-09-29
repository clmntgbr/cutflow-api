package fixture

import (
	"testing"

	"go-api/internal/domain/port"
)

func TestParseClockMs(t *testing.T) {
	ms, err := parseClockMs("01:40.745")
	if err != nil {
		t.Fatal(err)
	}
	if ms != 100745 {
		t.Fatalf("got %d want 100745", ms)
	}
	ms, err = parseClockMs("28:35.870")
	if err != nil {
		t.Fatal(err)
	}
	if ms != 1715870 {
		t.Fatalf("got %d want 1715870", ms)
	}
}

func TestProposalsFromClipsBlock(t *testing.T) {
	got := proposalsFromBlock(ResponseBlock{
		Clips: []Clip{
			{Start: "00:05.339", End: "00:21.052", Content: "Hello world"},
		},
	})
	if len(got) != 1 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].StartMs != 5339 || got[0].EndMs != 21052 {
		t.Fatalf("%+v", got[0])
	}
	if got[0].Hook != "Hello world" || got[0].Score != 0.8 {
		t.Fatalf("%+v", got[0])
	}
}

func TestClientServesOnce(t *testing.T) {
	c := &Client{
		provider: "fixture",
		model:    "fixture",
		proposals: []port.ViralLLMProposal{
			{StartMs: 0, EndMs: 1000, Score: 0.9},
		},
	}
	first, err := c.Analyze(nil, port.ViralAnalyzeInput{})
	if err != nil || len(first) != 1 {
		t.Fatalf("first=%v err=%v", first, err)
	}
	second, err := c.Analyze(nil, port.ViralAnalyzeInput{})
	if err != nil || len(second) != 0 {
		t.Fatalf("second=%v err=%v", second, err)
	}
}
