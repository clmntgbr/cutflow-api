package mediaconfig

import (
	"testing"

	"github.com/google/uuid"
)

func TestResolveSilenceThreshold_AutoUsesLevelOffset(t *testing.T) {
	cfg := NewDefault(mustParseUUID("11111111-1111-1111-1111-111111111111"))
	cfg.SilenceThresholdMode = ThresholdModeAuto
	cfg.SilenceDetectionLevel = DetectionLevelAggressive

	got := cfg.ResolveSilenceThreshold(-48)
	// aggressive offset = +10 → -38
	if got != -38 {
		t.Fatalf("got %.1f want -38", got)
	}
}

func TestResolveSilenceThreshold_Manual(t *testing.T) {
	manual := -32.0
	cfg := NewDefault(mustParseUUID("11111111-1111-1111-1111-111111111111"))
	cfg.SilenceThresholdMode = ThresholdModeManual
	cfg.SilenceThresholdDB = &manual

	got := cfg.ResolveSilenceThreshold(-48)
	if got != -32 {
		t.Fatalf("got %.1f want -32", got)
	}
}

func mustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}
