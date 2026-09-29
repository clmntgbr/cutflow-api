package subtitle

// Style configures animated word-highlight ASS output.
type Style struct {
	FontFamily   string
	FontSize     int
	TextColor    string // #RRGGBB
	ActiveColor  string // #RRGGBB
	OutlineColor string // #RRGGBB
	OutlineSize  float64
	Shadow       float64
	ActiveScale  float64 // 1.0 = no scale, 1.1 = +10%
	Bold         bool
	MaxWords     int
	MaxGapMs     int64 // split group when silence between words exceeds this
	MaxChars     int   // soft wrap on visual length (0 = disabled)
	PlayResX     int
	PlayResY     int
	// Alignment: 2 bottom-center, 5 middle-center, 8 top-center (numpad).
	Alignment int
	MarginL   int
	MarginR   int
	MarginV   int
}

// TikTokClassic is the default vertical short-form preset.
func TikTokClassic() Style {
	return Style{
		FontFamily:   "Montserrat",
		FontSize:     64,
		TextColor:    "#FFFFFF",
		ActiveColor:  "#FFD600",
		OutlineColor: "#000000",
		OutlineSize:  4,
		Shadow:       2,
		ActiveScale:  1.1,
		Bold:         true,
		MaxWords:     3,
		MaxGapMs:     500,
		MaxChars:     28,
		PlayResX:     1080,
		PlayResY:     1920,
		Alignment:    2,
		MarginL:      40,
		MarginR:      40,
		MarginV:      180,
	}
}

func (s Style) normalized() Style {
	out := s
	if out.FontFamily == "" {
		out.FontFamily = "Arial"
	}
	if out.FontSize <= 0 {
		out.FontSize = 48
	}
	if out.TextColor == "" {
		out.TextColor = "#FFFFFF"
	}
	if out.ActiveColor == "" {
		out.ActiveColor = "#FFD600"
	}
	if out.OutlineColor == "" {
		out.OutlineColor = "#000000"
	}
	if out.OutlineSize < 0 {
		out.OutlineSize = 0
	}
	if out.Shadow < 0 {
		out.Shadow = 0
	}
	if out.ActiveScale <= 0 {
		out.ActiveScale = 1
	}
	if out.MaxWords <= 0 {
		out.MaxWords = 3
	}
	if out.MaxGapMs <= 0 {
		out.MaxGapMs = 500
	}
	if out.PlayResX <= 0 {
		out.PlayResX = 1080
	}
	if out.PlayResY <= 0 {
		out.PlayResY = 1920
	}
	if out.Alignment <= 0 {
		out.Alignment = 2
	}
	if out.MarginL < 0 {
		out.MarginL = 40
	}
	if out.MarginR < 0 {
		out.MarginR = 40
	}
	if out.MarginV < 0 {
		out.MarginV = 120
	}
	return out
}
