package enbypub

import (
	"errors"
)

// A PathComponent describes a static string, an attribute of a published Text, or an attribute of the Feed
type PathComponent struct {
	String *string    `yaml:",omitempty"`
	Attr   *Attribute `yaml:",omitempty"`
}

type Attribute string

const (
	TextAttributeSlug     = Attribute("slug")
	TextAttributeId       = Attribute("id")
	TextAttributeTemplate = Attribute("template")

	TextAttributeYear      = Attribute("year")
	TextAttributeMonth     = Attribute("month")
	TextAttributeDay       = Attribute("day")
	TextAttributeDayOfWeek = Attribute("dow")
	TextAttributeYMD       = Attribute("date")

	FeedAttributeSlug = Attribute("feedslug")
	FeedAttributeId   = Attribute("feedid")
)

func (pc PathComponent) Get(F *Feed, T *Text) (string, error) {
	if pc.String != nil && pc.Attr != nil {
		return "", errors.New("path component defines both an attribute and a string")
	}

	switch {
	case pc.String != nil:
		return *pc.String, nil
	case pc.Attr != nil:
		return F.Get(*pc.Attr, T)
	}

	return "", errors.New("path component does not define any values")
}
