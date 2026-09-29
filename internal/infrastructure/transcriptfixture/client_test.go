package fixture

import (
	"testing"
)

func TestWordsFromSRT(t *testing.T) {
	srt := "1\n00:00:00,260 --> 00:00:01,140\nLes amis, j'espère\n\n2\n00:00:01,640 --> 00:00:02,200\nallez bien.\n"
	words, text := wordsFromSRT(srt)
	if text == "" {
		t.Fatal("expected text")
	}
	if len(words) < 4 {
		t.Fatalf("expected words, got %d (%v)", len(words), words)
	}
	if words[0].StartMs != 260 {
		t.Fatalf("first word start=%d", words[0].StartMs)
	}
}
