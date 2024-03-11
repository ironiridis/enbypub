package enbypub

import (
	"regexp"
	"strings"
)

var SlugMakeDash = regexp.MustCompile(`[^[:alnum:]]+`)

func Sluggify(in string) string {
	out_dashes := SlugMakeDash.ReplaceAllString(in, "-")
	out_downcased := strings.ToLower(out_dashes)
	return out_downcased
}
