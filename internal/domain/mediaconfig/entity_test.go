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
	// aggressive offset = +12 → -36
	if got != -36 {
		t.Fatalf("got %.1f want -36", got)
	}
}

func TestResolveSilenceThreshold_AutoLevelsStayDistinctForTypicalFloor(t *testing.T) {
	cfg := NewDefault(mustParseUUID("11111111-1111-1111-1111-111111111111"))
	cfg.SilenceThresholdMode = ThresholdModeAuto

	floor := -30.0
	cfg.SilenceDetectionLevel = DetectionLevelLow
	low := cfg.ResolveSilenceThreshold(floor)
	cfg.SilenceDetectionLevel = DetectionLevelModerate
	moderate := cfg.ResolveSilenceThreshold(floor)
	cfg.SilenceDetectionLevel = DetectionLevelAggressive
	aggressive := cfg.ResolveSilenceThreshold(floor)
	cfg.SilenceDetectionLevel = DetectionLevelVeryAggressive
	very := cfg.ResolveSilenceThreshold(floor)

	if !(low < moderate && moderate < aggressive && aggressive < very) {
		t.Fatalf("expected distinct ascending thresholds, got low=%.1f moderate=%.1f aggressive=%.1f very=%.1f",
			low, moderate, aggressive, very)
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
