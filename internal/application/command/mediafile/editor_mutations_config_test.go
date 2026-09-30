package mediafile

import (
	"testing"

	domainmediaconfig "go-api/internal/domain/mediaconfig"

	"github.com/google/uuid"
)

func TestSilenceAnalysisParamsChanging(t *testing.T) {
	cfg := domainmediaconfig.NewDefault(uuid.New())
	cfg.SilenceDetectionLevel = domainmediaconfig.DetectionLevelAggressive

	level := domainmediaconfig.DetectionLevelLow
	if !silenceAnalysisParamsChanging(cfg, &SilenceConfigPatch{DetectionLevel: &level}) {
		t.Fatal("expected redetect when detection level changes")
	}

	same := domainmediaconfig.DetectionLevelAggressive
	minMs := 1200
	pad := 240
	if silenceAnalysisParamsChanging(cfg, &SilenceConfigPatch{
		DetectionLevel:  &same,
		MinDurationMs:   &minMs,
		PaddingBeforeMs: &pad,
		PaddingAfterMs:  &pad,
	}) {
		t.Fatal("edit filters alone must not redetect")
	}
}

func TestTimelineRelevantConfigChanged_SilenceDrawerPayload(t *testing.T) {
	cfg := domainmediaconfig.NewDefault(uuid.New())
	before := snapshotTimelineRelevantConfig(cfg)

	cfg.SilenceDetectionLevel = domainmediaconfig.DetectionLevelLow
	cfg.SilenceMinDurationMs = 1200
	cfg.SilencePaddingBeforeMs = 240
	cfg.SilencePaddingAfterMs = 280
	if !timelineRelevantConfigChanged(before, cfg) {
		t.Fatal("expected timeline rebuild when silence drawer fields change")
	}

	afterSame := snapshotTimelineRelevantConfig(cfg)
	if timelineRelevantConfigChanged(afterSame, cfg) {
		t.Fatal("unchanged config must not rebuild")
	}
}
