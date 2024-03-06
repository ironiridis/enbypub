package enbypub

import (
	"time"
)

// A PathComponent describes either a static string or an attribute of a published Text
type PathComponent struct {
	String    *string
	Attribute *TextAttribute
}

type Feed struct {
	// CanonicalPath specifies the location where a Text will be publicly reachable relative to a public root
	CanonicalPath []PathComponent

	// MaximumCount, if specified, limits the feed to the MaximumCount most recent Texts according to their Created time
	MaximumCount *uint

	// MaximumAge, if specified, limits the feed to the Texts with a Created time no older than MaximumAge
	MaximumAge *time.Duration
}

// type PublishedFeed struct {
// 	feed    *Feed
// 	Updated time.Time
// 	Texts   []uuid.UUID
// }
