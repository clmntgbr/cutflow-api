package timeline

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	domainsilence "go-api/internal/domain/silence"
	domaintranscriptissue "go-api/internal/domain/transcriptissue"

	"github.com/google/uuid"
)

// FingerprintInput is the stable material used to skip identical rebuilds.
type FingerprintInput struct {
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
	EngineVersion        string
}

func ComputeFingerprint(in FingerprintInput) string {
	var b strings.Builder
	fmt.Fprintf(&b, "media=%s\ndur=%d\nengine=%s\n", in.MediaFileID.String(), in.MediaDurationMs, in.EngineVersion)
	fmt.Fprintf(&b, "silence=%t filler=%t rep=%t\n", in.SilenceRemoval, in.FillerRemoval, in.RepetitionRemoval)
	fmt.Fprintf(&b, "minSil=%d padB=%d padA=%d speechMin=%d\n",
		in.SilenceMinDurationMs, in.SilencePadBeforeMs, in.SilencePadAfterMs, in.SpeechMinDurationMs)

	sil := append([]domainsilence.Interval(nil), in.Silences...)
	sort.Slice(sil, func(i, j int) bool {
		if sil[i].StartMs == sil[j].StartMs {
			return sil[i].EndMs < sil[j].EndMs
		}
		return sil[i].StartMs < sil[j].StartMs
	})
	for _, s := range sil {
		fmt.Fprintf(&b, "s:%d-%d\n", s.StartMs, s.EndMs)
	}

	issues := append([]*domaintranscriptissue.Issue(nil), in.Issues...)
	sort.Slice(issues, func(i, j int) bool {
		if issues[i] == nil {
			return true
		}
		if issues[j] == nil {
			return false
		}
		if issues[i].SourceStartMs == issues[j].SourceStartMs {
			return issues[i].SourceEndMs < issues[j].SourceEndMs
		}
		return issues[i].SourceStartMs < issues[j].SourceStartMs
	})
	for _, issue := range issues {
		if issue == nil {
			continue
		}
		fmt.Fprintf(&b, "i:%s:%d-%d\n", issue.Type, issue.SourceStartMs, issue.SourceEndMs)
	}

	overrides := append([]Override(nil), in.Overrides...)
	sort.Slice(overrides, func(i, j int) bool {
		if overrides[i].SourceStartMs == overrides[j].SourceStartMs {
			return overrides[i].SourceEndMs < overrides[j].SourceEndMs
		}
		return overrides[i].SourceStartMs < overrides[j].SourceStartMs
	})
	for _, o := range overrides {
		fmt.Fprintf(&b, "o:%s:%s:%d-%d\n", o.Type, o.Action, o.SourceStartMs, o.SourceEndMs)
	}

	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}
