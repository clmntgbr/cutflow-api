package timeline

import (
	"testing"

	domainsilence "go-api/internal/domain/silence"
	domaintranscriptissue "go-api/internal/domain/transcriptissue"

	"github.com/google/uuid"
)

func TestNormalizeRanges_Overlap(t *testing.T) {
	got := NormalizeRanges([]Range{
		{StartMs: 10000, EndMs: 15000, Reasons: []string{DecisionSilence}},
		{StartMs: 14500, EndMs: 15400, Reasons: []string{DecisionFiller}},
	})
	if len(got) != 1 || got[0].StartMs != 10000 || got[0].EndMs != 15400 {
		t.Fatalf("got %+v", got)
	}
	if len(got[0].Reasons) != 2 {
		t.Fatalf("reasons=%v", got[0].Reasons)
	}
}

func TestBuildTimeline_ExampleFromSpec(t *testing.T) {
	mediaID := uuid.New()
	projectID := uuid.New()
	decisions := []Decision{
		{Action: ActionRemove, SourceStartMs: 8000, SourceEndMs: 10000, Reasons: []string{DecisionSilence}},
		{Action: ActionRemove, SourceStartMs: 14000, SourceEndMs: 15000, Reasons: []string{DecisionFiller}},
		{Action: ActionRemove, SourceStartMs: 25000, SourceEndMs: 28000, Reasons: []string{DecisionSilence}},
		{Action: ActionRemove, SourceStartMs: 39000, SourceEndMs: 42000, Reasons: []string{DecisionRepetition}},
		{Action: ActionRemove, SourceStartMs: 45000, SourceEndMs: 47000, Reasons: []string{DecisionSilence}},
	}
	tl, err := BuildTimeline(projectID, mediaID, 60000, decisions, 0)
	if err != nil {
		t.Fatal(err)
	}
	if tl.DurationMs != 49000 {
		t.Fatalf("duration=%d want 49000", tl.DurationMs)
	}
	want := [][4]int64{
		{0, 8000, 0, 8000},
		{10000, 14000, 8000, 12000},
		{15000, 25000, 12000, 22000},
		{28000, 39000, 22000, 33000},
		{42000, 45000, 33000, 36000},
		{47000, 60000, 36000, 49000},
	}
	if len(tl.Segments) != len(want) {
		t.Fatalf("segments=%d", len(tl.Segments))
	}
	for i, w := range want {
		s := tl.Segments[i]
		if s.SourceStartMs != w[0] || s.SourceEndMs != w[1] || s.OutputStartMs != w[2] || s.OutputEndMs != w[3] {
			t.Fatalf("seg[%d]=%+v want %v", i, s, w)
		}
	}
}

func TestMapper_SourceInCut(t *testing.T) {
	tl, err := BuildTimeline(uuid.New(), uuid.New(), 30000, []Decision{
		{Action: ActionRemove, SourceStartMs: 10000, SourceEndMs: 15000},
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	m := NewMapper(tl)
	if _, ok := m.MapSourceToOutput(11000); ok {
		t.Fatal("expected removed")
	}
	out, ok := m.MapSourceToOutput(20000)
	if !ok || out != 15000 {
		t.Fatalf("got %d ok=%v", out, ok)
	}
}

func TestBuildDecisions_SilencePaddingAndOverrideKeep(t *testing.T) {
	mediaID := uuid.New()
	decisions := BuildDecisions(BuildInput{
		MediaFileID:          mediaID,
		MediaDurationMs:      60000,
		SilenceRemoval:       true,
		SilenceMinDurationMs: 500,
		SilencePadBeforeMs:   50,
		SilencePadAfterMs:    150,
		Silences: []domainsilence.Interval{
			{StartMs: 10000, EndMs: 12000},
		},
		Overrides: []Override{
			{MediaFileID: mediaID, Action: ActionKeep, SourceStartMs: 10150, SourceEndMs: 11950},
		},
	})
	if len(decisions) != 0 {
		t.Fatalf("expected keep override to cancel silence, got %+v", decisions)
	}
}

func TestBuildDecisions_WithFiller(t *testing.T) {
	mediaID := uuid.New()
	conf := 0.98
	decisions := BuildDecisions(BuildInput{
		MediaFileID:     mediaID,
		FillerRemoval:   true,
		SilenceRemoval:  false,
		Issues: []*domaintranscriptissue.Issue{
			{Type: domaintranscriptissue.TypeFiller, SourceStartMs: 14100, SourceEndMs: 14600, Confidence: &conf},
		},
	})
	if len(decisions) != 1 || decisions[0].Type != DecisionFiller {
		t.Fatalf("got %+v", decisions)
	}
}

func TestFingerprintStable(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	in := FingerprintInput{
		MediaFileID:     id,
		MediaDurationMs: 60000,
		EngineVersion:   EngineVersion,
		Silences:        []domainsilence.Interval{{StartMs: 1, EndMs: 2}},
	}
	a := ComputeFingerprint(in)
	b := ComputeFingerprint(in)
	if a != b || a == "" {
		t.Fatalf("unstable fingerprint")
	}
}
