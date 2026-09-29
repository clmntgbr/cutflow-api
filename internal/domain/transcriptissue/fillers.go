package transcriptissue

import "strings"

func fillerDict(language string) (single map[string]bool, phrases map[string]bool) {
	lang := strings.ToLower(strings.TrimSpace(language))
	if len(lang) > 2 {
		lang = lang[:2]
	}
	switch lang {
	case "en":
		return map[string]bool{
				"uh": true, "um": true, "uhm": true, "er": true, "ah": true, "eh": true, "hmm": true, "mm": true,
			}, map[string]bool{
				"you know": true, "i mean": true,
			}
	default: // fr and fallback
		return map[string]bool{
				"euh": true, "heu": true, "eh": true, "hum": true, "hmm": true, "mmh": true, "mm": true,
				"ben": true, "bah": true, "pff": true, "pfft": true,
			}, map[string]bool{
				"du coup": true, "en fait": true,
			}
	}
}
