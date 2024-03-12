package enbypub

import (
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var SlugMakeDash = regexp.MustCompile(`[^[:alnum:]]+`)

func Sluggify(in *string) *string {
	out := strings.ToLower(SlugMakeDash.ReplaceAllString(*in, "-"))
	return &out
}

func Titlenate(fn *string) *string {
	base := filepath.Base(*fn)
	if idx := strings.LastIndexByte(base, '.'); idx >= 1 {
		// remove outermost file extension
		base = base[:idx]
	}
	base = regexp.MustCompile(`[-_\.]+`).ReplaceAllString(base, " ")
	out := cases.Title(language.AmericanEnglish).String(base)
	return &out
}
