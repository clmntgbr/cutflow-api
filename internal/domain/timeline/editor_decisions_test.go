package timeline

import (
	"testing"

	"github.com/google/uuid"
)

func TestResolveEditorDecisions_AutoRemove(t *testing.T) {
	mediaID := uuid.New()
	auto := []Decision{{
		ID:            uuid.New(),
		MediaFileID:   mediaID,
		Type:          DecisionSilence,
		SourceStartMs: 1000,
		SourceEndMs:   2000,
		Action:        ActionRemove,
		Source:        SourceAutomatic,
	}}
	got := ResolveEditorDecisions(auto, nil, nil)
	if len(got) != 1 {
		t.Fatalf("len: got %d", len(got))
	}
	if got[0].EffectiveAction != ActionRemove || got[0].ModifiedByUser {
		t.Fatalf("got %#v", got[0])
	}
	if got[0].AutomaticAction == nil || *got[0].AutomaticAction != ActionRemove {
		t.Fatalf("automatic: %#v", got[0].AutomaticAction)
	}
}

func TestResolveEditorDecisions_KeepOverride(t *testing.T) {
	mediaID := uuid.New()
	auto := []Decision{{
		ID:            uuid.New(),
		MediaFileID:   mediaID,
		Type:          DecisionFiller,
		SourceStartMs: 1000,
		SourceEndMs:   2000,
		Action:        ActionRemove,
		Source:        SourceAutomatic,
	}}
	overrides := []Override{{
		MediaFileID:   mediaID,
		Type:          DecisionFiller,
		SourceStartMs: 1000,
		SourceEndMs:   2000,
		Action:        ActionKeep,
	}}
	got := ResolveEditorDecisions(auto, overrides, map[DecisionKey]string{
		{Type: DecisionFiller, Start: 1000, End: 2000}: "euh",
	})
	if len(got) != 1 {
		t.Fatalf("len: got %d", len(got))
	}
	if got[0].EffectiveAction != ActionKeep || !got[0].ModifiedByUser {
		t.Fatalf("got %#v", got[0])
	}
	if got[0].Label == nil || *got[0].Label != "euh" {
		t.Fatalf("label: %#v", got[0].Label)
	}
}

func TestResolveEditorDecisions_SilenceKeepDoesNotMarkOverlappingFiller(t *testing.T) {
	mediaID := uuid.New()
	auto := []Decision{
		{
			MediaFileID:   mediaID,
			Type:          DecisionSilence,
			SourceStartMs: 1000,
			SourceEndMs:   3000,
			Action:        ActionRemove,
		},
		{
			MediaFileID:   mediaID,
			Type:          DecisionFiller,
			SourceStartMs: 1500,
			SourceEndMs:   1800,
			Action:        ActionRemove,
		},
	}
	overrides := []Override{{
		MediaFileID:   mediaID,
		Type:          DecisionSilence,
		SourceStartMs: 1000,
		SourceEndMs:   3000,
		Action:        ActionKeep,
	}}
	got := ResolveEditorDecisions(auto, overrides, nil)
	if len(got) != 2 {
		t.Fatalf("len: got %d", len(got))
	}
	byType := map[string]EditorDecision{}
	for _, d := range got {
		byType[d.Type] = d
	}
	if byType[DecisionSilence].EffectiveAction != ActionKeep || !byType[DecisionSilence].ModifiedByUser {
		t.Fatalf("silence: %#v", byType[DecisionSilence])
	}
	if byType[DecisionFiller].EffectiveAction != ActionRemove || byType[DecisionFiller].ModifiedByUser {
		t.Fatalf("filler should stay automatic remove: %#v", byType[DecisionFiller])
	}
}

func TestResolveEditorDecisions_ManualRemove(t *testing.T) {
	mediaID := uuid.New()
	overrides := []Override{{
		MediaFileID:   mediaID,
		Type:          DecisionManual,
		SourceStartMs: 5000,
		SourceEndMs:   6000,
		Action:        ActionRemove,
	}}
	got := ResolveEditorDecisions(nil, overrides, nil)
	if len(got) != 1 {
		t.Fatalf("len: got %d", len(got))
	}
	if got[0].Type != DecisionManual || got[0].AutomaticAction != nil {
		t.Fatalf("got %#v", got[0])
	}
	if got[0].EffectiveAction != ActionRemove || !got[0].ModifiedByUser {
		t.Fatalf("got %#v", got[0])
	}
}
