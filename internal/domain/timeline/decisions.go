package timeline

import (
	domainsilence "go-api/internal/domain/silence"
	domaintranscriptissue "go-api/internal/domain/transcriptissue"

	"github.com/google/uuid"
)

// BuildInput gathers SOURCE TIME analyses + config for decision building.
type BuildInput struct {
	MediaFileID          uuid.UUID
	MediaDurationMs      int64
	SilenceRemoval       bool
	FillerRemoval        bool
	RepetitionRemoval    bool
	SilenceMinDurationMs int
	SilencePadBeforeMs   int
	SilencePadAfterMs    int
	SpeechMinDurationMs  int
	Silences             []domainsilence.Interval
	Issues               []*domaintranscriptissue.Issue
	Overrides            []Override
}

// BuildDecisions converts analyses + config + overrides into effective REMOVE decisions.
func BuildDecisions(in BuildInput) []Decision {
	var removes []Decision

	if in.SilenceRemoval {
		filtered := domainsilence.ApplyEditFilters(
			in.Silences,
			in.SilenceMinDurationMs,
			in.SilencePadBeforeMs,
			in.SilencePadAfterMs,
		)
		for _, iv := range filtered {
			removes = append(removes, Decision{
				MediaFileID:   in.MediaFileID,
				Type:          DecisionSilence,
				SourceStartMs: iv.StartMs,
				SourceEndMs:   iv.EndMs,
				Action:        ActionRemove,
				Source:        SourceAutomatic,
				Reasons:       []string{DecisionSilence},
			})
		}
	}

	for _, issue := range in.Issues {
		if issue == nil {
			continue
		}
		switch issue.Type {
		case domaintranscriptissue.TypeFiller:
			if !in.FillerRemoval {
				continue
			}
			removes = append(removes, Decision{
				MediaFileID:   in.MediaFileID,
				Type:          DecisionFiller,
				SourceStartMs: issue.SourceStartMs,
				SourceEndMs:   issue.SourceEndMs,
				Action:        ActionRemove,
				Source:        SourceAutomatic,
				Confidence:    issue.Confidence,
				Reasons:       []string{DecisionFiller},
			})
		case domaintranscriptissue.TypeRepetition:
			if !in.RepetitionRemoval {
				continue
			}
			removes = append(removes, Decision{
				MediaFileID:   in.MediaFileID,
				Type:          DecisionRepetition,
				SourceStartMs: issue.SourceStartMs,
				SourceEndMs:   issue.SourceEndMs,
				Action:        ActionRemove,
				Source:        SourceAutomatic,
				Confidence:    issue.Confidence,
				Reasons:       []string{DecisionRepetition},
			})
		case domaintranscriptissue.TypeFalseStart:
			if !in.RepetitionRemoval {
				continue
			}
			removes = append(removes, Decision{
				MediaFileID:   in.MediaFileID,
				Type:          DecisionFalseStart,
				SourceStartMs: issue.SourceStartMs,
				SourceEndMs:   issue.SourceEndMs,
				Action:        ActionRemove,
				Source:        SourceAutomatic,
				Confidence:    issue.Confidence,
				Reasons:       []string{DecisionFalseStart},
			})
		}
	}

	return applyOverrides(removes, in.Overrides)
}

func applyOverrides(auto []Decision, overrides []Override) []Decision {
	if len(overrides) == 0 {
		return auto
	}

	ranges := make([]Range, 0, len(auto))
	for _, d := range auto {
		if d.Action != ActionRemove {
			continue
		}
		ranges = append(ranges, Range{
			StartMs: d.SourceStartMs,
			EndMs:   d.SourceEndMs,
			Reasons: append([]string(nil), d.Reasons...),
		})
	}
	ranges = NormalizeRanges(ranges)

	for _, o := range overrides {
		switch o.Action {
		case ActionKeep:
			ranges = subtractRange(ranges, Range{StartMs: o.SourceStartMs, EndMs: o.SourceEndMs})
		case ActionRemove:
			ranges = NormalizeRanges(append(ranges, Range{
				StartMs: o.SourceStartMs,
				EndMs:   o.SourceEndMs,
				Reasons: []string{DecisionManual},
			}))
		}
	}

	out := make([]Decision, 0, len(ranges))
	for _, r := range ranges {
		typ := DecisionManual
		if len(r.Reasons) > 0 {
			typ = r.Reasons[0]
		}
		mediaID := uuid.Nil
		if len(auto) > 0 {
			mediaID = auto[0].MediaFileID
		}
		out = append(out, Decision{
			MediaFileID:   mediaID,
			Type:          typ,
			SourceStartMs: r.StartMs,
			SourceEndMs:   r.EndMs,
			Action:        ActionRemove,
			Source:        SourceAutomatic,
			Reasons:       r.Reasons,
		})
	}
	return out
}

func subtractRange(ranges []Range, keep Range) []Range {
	if keep.EndMs <= keep.StartMs {
		return ranges
	}
	out := make([]Range, 0, len(ranges)+1)
	for _, r := range ranges {
		if keep.EndMs <= r.StartMs || keep.StartMs >= r.EndMs {
			out = append(out, r)
			continue
		}
		if keep.StartMs > r.StartMs {
			out = append(out, Range{StartMs: r.StartMs, EndMs: keep.StartMs, Reasons: r.Reasons})
		}
		if keep.EndMs < r.EndMs {
			out = append(out, Range{StartMs: keep.EndMs, EndMs: r.EndMs, Reasons: r.Reasons})
		}
	}
	return NormalizeRanges(out)
}
