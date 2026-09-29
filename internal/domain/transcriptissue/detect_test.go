package transcriptissue

import (
	"testing"

	domaintranscript "go-api/internal/domain/transcript"
)

func TestDetect_Filler(t *testing.T) {
	words := []domaintranscript.Word{
		word(0, "Donc", 1000, 1200, 0.99),
		word(1, "euh", 12500, 12900, 0.98),
		word(2, "je", 13000, 13200, 0.97),
	}
	issues := Detect(words, DetectOptions{DetectFillers: true, Language: "fr"})
	if len(issues) != 1 {
		t.Fatalf("len=%d want 1: %+v", len(issues), issues)
	}
	if issues[0].Type != TypeFiller || issues[0].Text != "euh" {
		t.Fatalf("got %+v", issues[0])
	}
	if issues[0].SourceStartMs != 12500 || issues[0].SourceEndMs != 12900 {
		t.Fatalf("times %+v", issues[0])
	}
}

func TestDetect_RepetitionKeepsLast(t *testing.T) {
	words := []domaintranscript.Word{
		word(0, "Je", 0, 100, 0.9),
		word(1, "je", 120, 200, 0.9),
		word(2, "pense", 220, 400, 0.9),
	}
	issues := Detect(words, DetectOptions{DetectRepetitions: true, Language: "fr"})
	if len(issues) != 1 || issues[0].Type != TypeRepetition || issues[0].Text != "Je" {
		t.Fatalf("got %+v", issues)
	}
	if issues[0].WordStartIndex != 0 || issues[0].WordEndIndex != 0 {
		t.Fatalf("indexes %+v", issues[0])
	}
}

func TestDetect_FalseStartRemovesFirstPassage(t *testing.T) {
	words := []domaintranscript.Word{
		word(0, "Je", 0, 100, 0.9),
		word(1, "vais", 120, 200, 0.9),
		word(2, "je", 400, 500, 0.9),
		word(3, "vais", 520, 600, 0.9),
		word(4, "vous", 620, 700, 0.9),
		word(5, "montrer", 720, 900, 0.9),
	}
	issues := Detect(words, DetectOptions{DetectFalseStarts: true, Language: "fr"})
	if len(issues) != 1 || issues[0].Type != TypeFalseStart {
		t.Fatalf("got %+v", issues)
	}
	if issues[0].Text != "Je vais" {
		t.Fatalf("text=%q", issues[0].Text)
	}
	if issues[0].WordStartIndex != 0 || issues[0].WordEndIndex != 1 {
		t.Fatalf("indexes %+v", issues[0])
	}
}

func TestApplyKinds(t *testing.T) {
	words := []domaintranscript.Word{
		word(0, "euh", 0, 100, 0.9),
		word(1, "ok", 120, 200, 0.9),
	}
	issues := Detect(words, DetectOptions{DetectFillers: true, Language: "fr"})
	got := ApplyKinds(words, issues)
	if got[0].Kind != domaintranscript.KindFiller {
		t.Fatalf("kind=%s", got[0].Kind)
	}
	if got[1].Kind != domaintranscript.KindSpeech {
		t.Fatalf("kind=%s", got[1].Kind)
	}
}

func word(i int, text string, start, end int64, conf float64) domaintranscript.Word {
	c := conf
	return domaintranscript.Word{
		WordIndex:     i,
		Text:          text,
		SourceStartMs: start,
		SourceEndMs:   end,
		Confidence:    &c,
		Kind:          domaintranscript.KindSpeech,
	}
}
