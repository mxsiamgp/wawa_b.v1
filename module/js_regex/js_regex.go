package js_regex

import "regexp"

var escapeTextPatternRegexp *regexp.Regexp = regexp.MustCompile("\\\\|\\^|\\$|\\.|\\*|\\+|\\?|\\(|\\)|\\[|]|\\{|}|\\|")

func EscapeTextPattern(text string) string {
	return escapeTextPatternRegexp.ReplaceAllStringFunc(text, func(frag string) string {
		return "\\" + frag;
	})
}
