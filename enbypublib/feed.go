package enbypub

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

type Feed struct {
	// Tags specifies the list of tags that will be scanned to add content to this Feed
	Tags []string `yaml:",omitempty"`

	// CanonicalPath specifies the location where a Text will be publicly reachable relative to a public root
	// If empty/nil, this feed does not produce output files
	CanonicalPath []PathComponent `yaml:",omitempty"`

	// MaximumCount, if specified, limits the feed to the MaximumCount most recent Texts according to their Created time
	// (not yet implemented)
	MaximumCount *uint `yaml:",omitempty"`

	// MaximumAge, if specified, limits the feed to the Texts with a Created time no older than MaximumAge
	// (not yet implemented)
	MaximumAge *time.Duration `yaml:",omitempty"`

	// Id is a randomly generated UUID for this Feed
	Id *uuid.UUID `yaml:",omitempty"`

	// Slug is a URL-friendly reference to this Feed
	Slug *string `yaml:",omitempty"`

	// Aggregators is a list of zero or more aggregation methods, like the Feed Index or RSS
	// (not yet implemented)
	Aggregators []map[string]any `yaml:",omitempty"`

	// Index contains a list of every Text that is published to this Feed
	// Note that this list is not neccessarily orderd in any particular way; the Index must be sorted if that's desired
	Index []*Text `yaml:"-"`

	// DefaultTemplate set a default "Style" value for a Text if one is not set
	DefaultTemplate *string `yaml:",omitempty"`

	// (not yet implemented)
	LinkTags []string `yaml:"-"`
	MetaTags []string `yaml:"-"`

	fs *FeedStructure
}

func (F *Feed) SortByCreated() {
	slices.SortFunc(F.Index, func(a, b *Text) int {
		return a.Created.Compare(*b.Created)
	})
}

func (F *Feed) Canonical(T *Text) ([]string, error) {
	var err error
	C := make([]string, len(F.CanonicalPath))
	for i := range F.CanonicalPath {
		C[i], err = F.CanonicalPath[i].Get(F, T)
		if err != nil {
			return nil, fmt.Errorf("cannot build canonical path for text %q from %+v: %w", T.Id, F.CanonicalPath[i], err)
		}
	}
	return C, nil
}

func (F Feed) CanonicalStructure() (*FeedStructure, error) {
	fs := &FeedStructure{
		Segments: make(map[string][]*Text),
		Files:    make(map[string]*Text, len(F.Index)),
	}
	var P strings.Builder
	for _, T := range F.Index {
		paths, err := F.Canonical(T)
		if err != nil {
			return nil, fmt.Errorf("cannot build structure for feed %v (%v): %w", F.Slug, F.Id, err)
		}
		P.Reset()
		for _, p := range paths[:len(paths)-1] {
			P.WriteString(p)
			fs.Segments[P.String()] = append(fs.Segments[P.String()], T)
			P.WriteByte('/')
		}
		P.WriteString(paths[len(paths)-1])
		P.WriteString(".html")
		fs.Files[P.String()] = T
	}
	return fs, nil
}

func (F Feed) Get(a Attribute, T *Text) (string, error) {
	switch a {
	case FeedAttributeSlug:
		if F.Slug != nil {
			return *F.Slug, nil
		}
	case FeedAttributeId:
		if F.Id != nil {
			return F.Id.String(), nil
		}
	case TextAttributeTemplate:
		if T.Template != nil {
			return T.Get(a)
		}
		if F.DefaultTemplate != nil {
			return *F.DefaultTemplate, nil
		}
	}

	return T.Get(a)
}

type Feeds map[uuid.UUID]*Feed

func (f Feeds) Scan(t Texts) error {
	for _, F := range f {
		for _, T := range t {
			if T.IsTagged(F.Tags...) {
				F.Index = append(F.Index, T)
			}
		}
	}
	return nil
}

// FeedStructure describes the content structure published to the feed.
type FeedStructure struct {
	// Segments is a map of each distinct path segment and the Texts under that segment
	Segments map[string][]*Text

	// Files is the generated final content path for each Text
	Files map[string]*Text
}

// type PublishedFeed struct {
// 	feed    *Feed
// 	Updated time.Time
// 	Texts   []uuid.UUID
// }

type rawFeed map[string]*Feed

func loadRawFeeds(r io.Reader) (rawFeed, error) {
	rf := rawFeed{}
	err := yaml.NewDecoder(r).Decode(&rf)
	return rf, err
}

func LoadFeedsFromFile(fn string) (Feeds, error) {
	var mustRewrite bool
	fp, err := os.Open(fn)
	if err != nil {
		return nil, fmt.Errorf("unable to open feeds file %q: %w", fn, err)
	}
	feeds, err := loadRawFeeds(fp)
	fp.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to load feeds from %q: %w", fn, err)
	}

	F := make(Feeds, len(feeds))
	for k := range feeds {
		if feeds[k].Id == nil {
			id := uuid.New()
			feeds[k].Id = &id
			mustRewrite = true
		}
		if feeds[k].Slug == nil {
			slug := Sluggify(k)
			feeds[k].Slug = &slug
			mustRewrite = true
		}
		F[*feeds[k].Id] = feeds[k]
	}

	if mustRewrite {
		fp, err := os.Create(fn)
		if err != nil {
			return nil, fmt.Errorf("unable to open feeds file %q to update: %w", fn, err)
		}
		err = yaml.NewEncoder(fp).Encode(feeds)
		fp.Close()
		if err != nil {
			return nil, fmt.Errorf("unable to update %q: %w", fn, err)
		}
	}
	return F, nil
}
