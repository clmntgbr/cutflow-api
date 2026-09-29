package mediafile

import "testing"

func TestDecideInitialTimelineEnqueue(t *testing.T) {
	tests := []struct {
		name         string
		silenceReady bool
		textReady    bool
		want         bool
	}{
		{name: "neither", silenceReady: false, textReady: false, want: false},
		{name: "silence only", silenceReady: true, textReady: false, want: false},
		{name: "text only", silenceReady: false, textReady: true, want: false},
		{name: "both", silenceReady: true, textReady: true, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecideInitialTimelineEnqueue(tt.silenceReady, tt.textReady)
			if got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
