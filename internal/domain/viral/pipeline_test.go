package viral

import (
	"testing"

	domaintranscript "go-api/internal/domain/transcript"
)

func TestFormatTimestamped(t *testing.T) {
	words := []domaintranscript.Word{
		{Text: "Hello", SourceStartMs: 0, SourceEndMs: 500},
		{Text: "world", SourceStartMs: 520, SourceEndMs: 900},
		{Text: "again", SourceStartMs: 6000, SourceEndMs: 6500},
	}
	got := FormatTimestamped(words, 5000)
	if got == "" || !contains(got, "[00:00.000") {
		t.Fatalf("unexpected format: %q", got)
	}
}

func TestResolveBoundaries(t *testing.T) {
	words := []domaintranscript.Word{
		{Text: "A", SourceStartMs: 1000, SourceEndMs: 1200},
		{Text: "B", SourceStartMs: 1300, SourceEndMs: 1500},
		{Text: "C", SourceStartMs: 2000, SourceEndMs: 2300},
	}
	start, end := ResolveBoundaries(words, 1250, 2100)
	if start != 1300 || end != 2300 {
		t.Fatalf("got %d-%d", start, end)
	}
}

func TestDeduplicateAndRank(t *testing.T) {
	a := &Candidate{SourceStartMs: 10000, SourceEndMs: 40000, Score: 0.9}
	b := &Candidate{SourceStartMs: 12000, SourceEndMs: 42000, Score: 0.8}
	c := &Candidate{SourceStartMs: 100000, SourceEndMs: 130000, Score: 0.85}
	got := Rank([]*Candidate{a, b, c}, 10)
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].Score != 0.9 {
		t.Fatalf("first score=%v", got[0].Score)
	}
}

func TestValidateCandidate(t *testing.T) {
	c := &Candidate{SourceStartMs: 0, SourceEndMs: 25000, Score: 0.8}
	err := ValidateCandidate(c, ValidateConfig{
		MediaDurationMs: 60000,
		MinDurationMs:   20000,
		MaxDurationMs:   90000,
		MinScore:        0.7,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || stringIndex(s, sub) >= 0)
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
