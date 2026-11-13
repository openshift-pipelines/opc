package formatting

import "regexp"

// ansiEscapeRegexp matches terminal escape sequences: CSI sequences (colors,
// cursor movements), OSC sequences (e.g. hyperlinks) terminated by BEL or ST,
// and the remaining two-character escapes.
var ansiEscapeRegexp = regexp.MustCompile(`\x1b(?:\[[0-?]*[ -/]*[@-~]|\][^\x07\x1b]*(?:\x07|\x1b\\)|[@-Z\\-_])`)

// StripANSI removes terminal escape sequences, such as color codes, from s.
func StripANSI(s string) string {
	return ansiEscapeRegexp.ReplaceAllString(s, "")
}
