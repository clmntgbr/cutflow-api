package timeline

import (
	"strconv"

	domaintranscriptissue "go-api/internal/domain/transcriptissue"

	"github.com/google/uuid"
)

// EditorDecision is the UI-facing resolved decision (automatic vs effective).
type EditorDecision struct {
	ID              uuid.UUID
	Type            string
	Label           *string
	SourceStartMs   int64
	SourceEndMs     int64
	AutomaticAction *string
	EffectiveAction string
	ModifiedByUser  bool
}

// ResolveEditorDecisions builds the editor decision list from automatic
// proposals and user overrides. Persisted edit_decision rows are effective
// removes only; this reconstructs automatic vs effective for the UI.
func ResolveEditorDecisions(auto []Decision, overrides []Override, labels map[DecisionKey]string) []EditorDecision {
	out := make([]EditorDecision, 0, len(auto)+len(overrides))

	for _, d := range auto {
		if d.Action != ActionRemove {
			continue
		}
		typ := normalizeEditorType(d.Type)
		autoAction := ActionRemove
		effective := ActionRemove
		modified := false
		if hasKeepOverride(overrides, d.SourceStartMs, d.SourceEndMs) {
			effective = ActionKeep
			modified = true
		}
		var label *string
		if text, ok := labels[DecisionKey{Type: typ, Start: d.SourceStartMs, End: d.SourceEndMs}]; ok && text != "" {
			copied := text
			label = &copied
		}
		id := d.ID
		if id == uuid.Nil {
			id = StableDecisionID(d.MediaFileID, typ, d.SourceStartMs, d.SourceEndMs)
		}
		out = append(out, EditorDecision{
			ID:              id,
			Type:            typ,
			Label:           label,
			SourceStartMs:   d.SourceStartMs,
			SourceEndMs:     d.SourceEndMs,
			AutomaticAction: &autoAction,
			EffectiveAction: effective,
			ModifiedByUser:  modified,
		})
	}

	for _, o := range overrides {
		if o.Action != ActionRemove {
			continue
		}
		if coveredByAuto(auto, o.SourceStartMs, o.SourceEndMs) {
			continue
		}
		id := o.ID
		if id == uuid.Nil {
			id = StableDecisionID(o.MediaFileID, DecisionManual, o.SourceStartMs, o.SourceEndMs)
		}
		out = append(out, EditorDecision{
			ID:              id,
			Type:            DecisionManual,
			Label:           nil,
			SourceStartMs:   o.SourceStartMs,
			SourceEndMs:     o.SourceEndMs,
			AutomaticAction: nil,
			EffectiveAction: ActionRemove,
			ModifiedByUser:  true,
		})
	}

	return out
}

// StableDecisionID returns a deterministic ID for editor decisions so overrides
// can target the same decision across GETs.
func StableDecisionID(mediaFileID uuid.UUID, typ string, startMs, endMs int64) uuid.UUID {
	return uuid.NewSHA1(mediaFileID, []byte(typ+":"+itoa(startMs)+":"+itoa(endMs)))
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}

// DecisionKey identifies an automatic proposal for label lookup.
type DecisionKey struct {
	Type  string
	Start int64
	End   int64
}

// LabelsFromIssues maps filler/repetition issue text onto decision keys.
func LabelsFromIssues(issues []*domaintranscriptissue.Issue) map[DecisionKey]string {
	out := make(map[DecisionKey]string, len(issues))
	for _, issue := range issues {
		if issue == nil || issue.Text == "" {
			continue
		}
		out[DecisionKey{
			Type:  normalizeEditorType(issue.Type),
			Start: issue.SourceStartMs,
			End:   issue.SourceEndMs,
		}] = issue.Text
	}
	return out
}

func normalizeEditorType(t string) string {
	switch t {
	case DecisionFalseStart:
		return DecisionRepetition
	default:
		return t
	}
}

func hasKeepOverride(overrides []Override, start, end int64) bool {
	for _, o := range overrides {
		if o.Action != ActionKeep {
			continue
		}
		if rangesOverlap(start, end, o.SourceStartMs, o.SourceEndMs) {
			return true
		}
	}
	return false
}

func coveredByAuto(auto []Decision, start, end int64) bool {
	for _, d := range auto {
		if d.Action != ActionRemove {
			continue
		}
		if rangesOverlap(start, end, d.SourceStartMs, d.SourceEndMs) {
			return true
		}
	}
	return false
}

func rangesOverlap(aStart, aEnd, bStart, bEnd int64) bool {
	return aStart < bEnd && bStart < aEnd
}
