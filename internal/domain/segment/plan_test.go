package segment

import "testing"

func TestPlanWindows_FiveMinuteChunks(t *testing.T) {
	windows := PlanWindows(30*60*1000, PlanConfig{
		DurationMs: 5 * 60 * 1000,
		OverlapMs:  2 * 1000,
		MaxCount:   40,
	})
	if len(windows) != 6 {
		t.Fatalf("count: got %d want 6", len(windows))
	}
	if windows[0].StartMs != 0 || windows[0].EndMs != 300000 {
		t.Fatalf("first window: %+v", windows[0])
	}
	if windows[0].ProcessingStartMs != 0 || windows[0].ProcessingEndMs != 302000 {
		t.Fatalf("first processing: %+v", windows[0])
	}
	last := windows[len(windows)-1]
	if last.EndMs != 30*60*1000 {
		t.Fatalf("last end: got %d", last.EndMs)
	}
	if last.ProcessingStartMs != 25*60*1000-2000 {
		t.Fatalf("last processing start: got %d", last.ProcessingStartMs)
	}
}

func TestPlanWindows_ShortVideoSingleWindow(t *testing.T) {
	windows := PlanWindows(90_000, PlanConfig{DurationMs: 300_000, OverlapMs: 2000})
	if len(windows) != 1 {
		t.Fatalf("count: got %d want 1", len(windows))
	}
	if windows[0].StartMs != 0 || windows[0].EndMs != 90_000 {
		t.Fatalf("window: %+v", windows[0])
	}
}
