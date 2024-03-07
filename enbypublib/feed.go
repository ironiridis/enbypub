package enbypub

import (
	"time"

	"github.com/google/uuid"
)

// A PathComponent describes either a static string or an attribute of a published Text
type PathComponent struct {
	String   *string
	TextAttr *TextAttribute
	FeedAttr *FeedAttribute
}

type FeedAttribute string

const (
	FeedAttributeSlug = FeedAttribute("slug")
	FeedAttributeId   = FeedAttribute("id")
)

type Feed struct {
	// CanonicalPath specifies the location where a Text will be publicly reachable relative to a public root
	// If empty/nil, this feed does not produce output files
	CanonicalPath []PathComponent `yaml:",omitempty"`

	// MaximumCount, if specified, limits the feed to the MaximumCount most recent Texts according to their Created time
	MaximumCount *uint `yaml:",omitempty"`

	// MaximumAge, if specified, limits the feed to the Texts with a Created time no older than MaximumAge
	MaximumAge *time.Duration `yaml:",omitempty"`

	// Id is a randomly generated UUID for this Feed
	Id *uuid.UUID `yaml:",omitempty"`

	// Slug is a URL-friendly reference to this Feed
	Slug *string `yaml:",omitempty"`
}

// type PublishedFeed struct {
// 	feed    *Feed
// 	Updated time.Time
// 	Texts   []uuid.UUID
// }
