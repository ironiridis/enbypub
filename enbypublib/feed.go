package enbypub

import (
	"time"

	"github.com/google/uuid"
)

// A PathComponent describes a static string, an attribute of a published Text, or an attribute of the Feed
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
	// Tags specifies the list of tags that will be scanned to add content to this Feed
	Tags []string `yaml:",omitempty"`

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

	// Aggregators is a list of zero or more aggregation methods, like the Feed Index or RSS
	Aggregators []map[string]any `yaml:",omitempty"`

	// plz ignore
	LinkTags []string
	MetaTags []string
}

// type PublishedFeed struct {
// 	feed    *Feed
// 	Updated time.Time
// 	Texts   []uuid.UUID
// }
