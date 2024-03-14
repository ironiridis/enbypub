package enbypub

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	Aggregators Aggregators `yaml:",omitempty"`

	// Index contains a list of every Text that is published to this Feed
	// Note that this list is not neccessarily orderd in any particular way; the Index must be sorted if that's desired
	Index []*Text `yaml:"-"`

	// DefaultTemplate set a default "Style" value for a Text if one is not set
	DefaultTemplate *string `yaml:",omitempty"`

	fs  *FeedStructure
	gen *Generator
}

// Sorts the Feed Index by the Created date of each Text, with the oldest first.
func (F *Feed) SortByCreatedAscending() {
	slices.SortFunc(F.Index, func(a, b *Text) int {
		return a.Created.Compare(*b.Created)
	})
}

// Sorts the Feed Index by the Created date of each Text, with the newest first.
func (F *Feed) SortByCreatedDescending() {
	slices.SortFunc(F.Index, func(a, b *Text) int {
		return b.Created.Compare(*a.Created)
	})
}

// GetPath gets the whole canonical path plus filename for a Text id
func (f *Feed) GetPath(id string) string {
	tid, err := uuid.Parse(id)
	if err != nil {
		return ""
	}
	for _, t := range f.Index {
		if *t.Id == tid {
			ps, err := f.Path(t)
			if err != nil {
				continue
			}
			fn, err := f.Filename(t)
			if err != nil {
				continue
			}
			ps = append(ps, fn)
			return filepath.Join(ps...)
		}
	}
	return ""
}

func (F *Feed) Path(T *Text) ([]string, error) {
	var err error
	if len(F.CanonicalPath) == 0 {
		return nil, fmt.Errorf("cannot get Path for %v because CanonicalPath is empty", F)
	}
	C := make([]string, len(F.CanonicalPath)-1)
	for i := range F.CanonicalPath[:len(F.CanonicalPath)-1] {
		C[i], err = F.CanonicalPath[i].Get(F, T)
		if err != nil {
			return nil, fmt.Errorf("cannot build canonical path for text %q from %+v: %w", T.Id, F.CanonicalPath[i], err)
		}
	}
	return C, nil
}

func (F *Feed) Filename(T *Text) (string, error) {
	if len(F.CanonicalPath) == 0 {
		return "", fmt.Errorf("cannot get Filename for %v because CanonicalPath is empty", F)
	}
	fn, err := F.CanonicalPath[len(F.CanonicalPath)-1].Get(F, T)
	if err != nil {
		return "", fmt.Errorf("failed to get Filename for %v in %v: %w", T, F, err)
	}

	return fn + ".html", nil
}

func (F Feed) CanonicalStructure() (*FeedStructure, error) {
	fs := &FeedStructure{
		Segments: make(map[string][]*Text),
		Files:    make(map[string]*Text, len(F.Index)),
	}
	var P strings.Builder
	for _, T := range F.Index {
		paths, err := F.Path(T)
		if err != nil {
			return nil, fmt.Errorf("cannot build structure for feed %v (%v): %w", F.Slug, F.Id, err)
		}
		P.Reset()
		for _, p := range paths {
			P.WriteString(p)
			fs.Segments[P.String()] = append(fs.Segments[P.String()], T)
			P.WriteByte('/')
		}
		fn, err := F.Filename(T)
		if err != nil {
			return nil, fmt.Errorf("cannot build structure for feed %v (%v): %w", F.Slug, F.Id, err)
		}
		P.WriteString(fn)
		fs.Files[P.String()] = T
	}
	F.fs = fs
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

func (F *Feed) Add(T *Text) error {
	F.Index = append(F.Index, T)
	for a := range F.Aggregators {
		if err := F.Aggregators[a].AddText(T); err != nil {
			return fmt.Errorf("failed to add text %v to feed: %w", T, err)
		}
	}
	return nil
}

func (F *Feed) Close() error {
	for a := range F.Aggregators {
		if err := F.Aggregators[a].Close(); err != nil {
			return fmt.Errorf("failed to close aggregator %d: %w", a, err)
		}
	}
	return nil
}

type Feeds map[uuid.UUID]*Feed

func (f Feeds) Scan(t Texts) error {
	for _, F := range f {
		for _, T := range t {
			if T.IsTagged(F.Tags...) {
				if err := F.Add(T); err != nil {
					return fmt.Errorf("failed to scan texts for feed %v: %w", F, err)
				}
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

// GetPath scans the file map for the specified uuid and returns the published Feed
// location for that Text if found, or an empty string
func (fs *FeedStructure) GetPath(id string) string {
	tid, err := uuid.Parse(id)
	if err != nil {
		return ""
	}
	for p, t := range fs.Files {
		if *t.Id == tid {
			return p
		}
	}
	return ""
}

type rawFeed map[string]*Feed

func loadRawFeeds(r io.Reader) (rawFeed, error) {
	rf := rawFeed{}
	err := yaml.NewDecoder(r).Decode(&rf)
	return rf, err
}

func LoadFeedsFromFile(fn string, g *Generator) (Feeds, error) {
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
			feeds[k].Slug = Sluggify(&k)
			mustRewrite = true
		}
		feeds[k].gen = g
		for a := range feeds[k].Aggregators {
			if err := feeds[k].Aggregators[a].Init(feeds[k], g); err != nil {
				return nil, fmt.Errorf("failed to initialize aggregator %d of feed %+v: %w", a, feeds[k], err)
			}
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
