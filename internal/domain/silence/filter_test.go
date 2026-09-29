package silence

import "testing"

func TestApplyEditFilters_MinDurationAndPadding(t *testing.T) {
	raw := []Interval{
		{StartMs: 8200, EndMs: 10300},  // 2100 raw → 1900 after pads
		{StartMs: 12000, EndMs: 12600}, // 600 raw → 400 after pads — dropped by min 500
		{StartMs: 25100, EndMs: 27800}, // 2700 raw → 2500 after pads
	}

	got := ApplyEditFilters(raw, 500, 50, 150)
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
	// start += paddingAfter(150), end -= paddingBefore(50)
	if got[0].StartMs != 8350 || got[0].EndMs != 10250 {
		t.Fatalf("first=%+v", got[0])
	}
	if got[1].StartMs != 25250 || got[1].EndMs != 27750 {
		t.Fatalf("second=%+v", got[1])
	}
}

func TestApplyEditFilters_PaddingConsumesInterval(t *testing.T) {
	raw := []Interval{{StartMs: 0, EndMs: 180}}
	got := ApplyEditFilters(raw, 0, 50, 150)
	if len(got) != 0 {
		t.Fatalf("expected empty, got %+v", got)
	}
}

func TestIntervalDurationMs(t *testing.T) {
	if (Interval{StartMs: 8200, EndMs: 10300}).DurationMs() != 2100 {
		t.Fatal("duration")
	}
}
