package util

import "regexp"

var re = regexp.MustCompile(`(\-|\+|\(|\)|\*|\_|\[|\])`)

// TODO: escaping should instead escape only markdown characters that doesn't make sense, like:
// "[Hello!]" should escape '[', ']' and '!', since it is not a link
func EscapeMarkdown(s string) string {
	return re.ReplaceAllString(s, `\$1`)
}
