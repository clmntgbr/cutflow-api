package subtitle

import (
	"strings"
	"testing"
)

func TestGenerateASS_WordHighlightEvents(t *testing.T) {
	words := []Word{
		{Text: "Je", StartMs: 1000, EndMs: 1200},
		{Text: "vais", StartMs: 1200, EndMs: 1500},
		{Text: "faire", StartMs: 1500, EndMs: 1800},
		{Text: "une", StartMs: 1800, EndMs: 1950},
		{Text: "vidéo", StartMs: 1950, EndMs: 2400},
	}
	style := TikTokClassic()
	style.MaxWords = 3
	style.MaxChars = 0

	ass := GenerateASS(words, style)
	if !strings.Contains(ass, "Dialogue: 0,0:00:01.00,0:00:01.20,Default,,0,0,0,,") {
		t.Fatalf("missing first dialogue window:\n%s", ass)
	}
	if !strings.Contains(ass, `{\c&H0000D6FF&\fscx110\fscy110}Je{\c&H00FFFFFF&\fscx100\fscy100} vais faire`) {
		t.Fatalf("missing active highlight on Je:\n%s", ass)
	}
	if !strings.Contains(ass, `Je {\c&H0000D6FF&\fscx110\fscy110}vais{\c&H00FFFFFF&\fscx100\fscy100} faire`) {
		t.Fatalf("missing active highlight on vais:\n%s", ass)
	}
	if !strings.Contains(ass, "une") && !strings.Contains(ass, "vidéo") {
		t.Fatalf("expected second group with une/vidéo:\n%s", ass)
	}
}

func TestGroupWords_SplitsOnMaxWordsAndPunctuation(t *testing.T) {
	words := []Word{
		{Text: "Hello.", StartMs: 0, EndMs: 200},
		{Text: "World", StartMs: 250, EndMs: 400},
		{Text: "foo", StartMs: 400, EndMs: 500},
		{Text: "bar", StartMs: 500, EndMs: 600},
		{Text: "baz", StartMs: 600, EndMs: 700},
	}
	groups := groupWords(words, Style{MaxWords: 3, MaxGapMs: 500, MaxChars: 0}.normalized())
	if len(groups) < 2 {
		t.Fatalf("expected split after punctuation, got %d groups", len(groups))
	}
	if groups[0].Words[0].Text != "Hello." {
		t.Fatalf("unexpected first group: %+v", groups[0].Words)
	}
}

func TestCssColorToASS(t *testing.T) {
	if got := cssColorToASS("#FFD600"); got != "&H0000D6FF" {
		t.Fatalf("got %s", got)
	}
	if got := cssColorToASS("#FFFFFF"); got != "&H00FFFFFF" {
		t.Fatalf("got %s", got)
	}
}
