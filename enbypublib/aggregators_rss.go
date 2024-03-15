package enbypub

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/language"
)

// RSSDate formats a time.Time as an RFC 1123 time.
type RSSDate time.Time

func (d *RSSDate) MarshalText() ([]byte, error) {
	if d == nil {
		return nil, errors.New("nil")
	}
	buf := (*time.Time)(d).Format(time.RFC1123)
	return []byte(buf), nil
}

type RSSGUID struct {
	*uuid.UUID `xml:"guid"`
	PermaLink  bool `xml:"isPermaLink,attr"`
}

// RSSAggregator provides an HTML article index intended for consumption with a web browser.
type RSSAggregator struct {
	f *Feed
	g *Generator

	// Kind is always "rss"
	Kind     string
	MinPath  *int    `yaml:",omitempty"`
	MaxPath  *int    `yaml:",omitempty"`
	Filename *string `yaml:",omitempty"`

	// Title, if provided, is used as the RSS feed title value. Otherwise a default is made from the Feed slug.
	Title *string `yaml:",omitempty"`

	// Description, if provided, is used as the RSS description. Otherwise a default is made from the Feed slug.
	Description *string `yaml:",omitempty"`

	// PublicBaseURL is used to calculate the public URLs in this feed.
	PublicBaseURL string

	// TTL is the minimum recommended refresh speed for consumers.
	TTL *time.Duration `yaml:",omitempty"`

	newest time.Time
	index  []*Text

	// indexes is a map of path components to Text lists
	indexes map[string]*RSSFeedDocument
}

func (a *RSSAggregator) document() (*RSSFeedDocument, error) {
	fd := &RSSFeedDocument{
		NS:      "http://www.w3.org/2005/Atom",
		Version: "2.0",
		Docs:    "https://www.rssboard.org/rss-specification",
	}
	if a.Title != nil {
		fd.Title = *a.Title
	} else {
		fd.Title = fmt.Sprintf("Feed: %s", *a.f.Slug)
	}
	if a.Description != nil {
		fd.Description = *a.Description
	} else {
		var t string
		switch len(a.f.Tags) {
		case 0:
			t = fmt.Sprintf("in the %s feed", *a.f.Slug)
		case 1:
			t = fmt.Sprintf("tagged as %s", a.f.Tags[0])
		case 2:
			t = fmt.Sprintf("tagged as %s or %s", a.f.Tags[0], a.f.Tags[1])
		default:
			t = fmt.Sprintf("tagged as %s, or %s", strings.Join(a.f.Tags[:len(a.f.Tags)-1], ", "), a.f.Tags[len(a.f.Tags)-1])
		}
		fd.Description = fmt.Sprintf("The latest posts %s", t)
	}
	if a.TTL != nil {
		fd.TTL = int(a.TTL.Minutes())
	}
	if u, err := url.Parse(a.PublicBaseURL); err != nil {
		return nil, fmt.Errorf("cannot produce RSS document, failed to parse public base URL %q: %w", a.PublicBaseURL, err)
	} else {
		fd.baseURL = u
		fd.Link = u.String()
	}
	fd.LastBuildDate = RSSDate(time.Now())
	return fd, nil
}

type RSSFeedItem struct {
	Title       string     `xml:"title"`
	Link        string     `xml:"link"`
	Description string     `xml:"description,omitempty"`
	Author      string     `xml:"author,omitempty"`
	PubDate     *RSSDate   `xml:"pubDate,omitempty"`
	Id          *uuid.UUID `xml:"guid,omitempty"`
}

type RSSFeedDocument struct {
	XMLName xml.Name `xml:"rss"`
	NS      string   `xml:"xmlns:atom,attr"` // always "http://www.w3.org/2005/Atom"
	Version string   `xml:"version,attr"`    // always "2.0"

	LastBuildDate RSSDate `xml:"channel>lastBuildDate"`

	Title       string        `xml:"channel>title,omitempty"`
	Link        string        `xml:"channel>link,omitempty"`
	Description string        `xml:"channel>description,omitempty"`
	Language    *language.Tag `xml:"channel>language,omitempty"`
	PubDate     *RSSDate      `xml:"channel>pubDate,omitempty"`
	Docs        string        `xml:"channel>docs,omitempty"` // always "https://www.rssboard.org/rss-specification"
	TTL         int           `xml:"channel>ttl,omitempty"`  // minutes

	// advises clients when not to re-fetch
	SkipHours *[]int    `xml:"channel>skipHours>hour,omitempty"` // hour of day in GMT
	SkipDays  *[]string `xml:"channel>skipDays>day,omitempty"`   // day of week

	Items []*RSSFeedItem `xml:"channel>item,omitempty"`

	baseURL *url.URL
}

func NewRSSFeedDocument() *RSSFeedDocument {
	fd := &RSSFeedDocument{
		NS:      "http://www.w3.org/2005/Atom",
		Version: "2.0",
		Docs:    "https://www.rssboard.org/rss-specification",
	}
	return fd
}

type RSSAggregatorContent struct {
	Meta  *MetaT
	Feed  *Feed
	Index []*Text
}

func (a *RSSAggregator) Init(f *Feed, g *Generator) error {
	a.f = f
	a.g = g
	if a.Filename == nil {
		a.Filename = strptr("rss.xml")
	}
	a.index = make([]*Text, 0, len(f.Index))
	a.indexes = make(map[string]*RSSFeedDocument)
	return nil
}

func (a *RSSAggregator) AddText(t *Text) error {
	if t.Created != nil && t.Created.After(a.newest) {
		a.newest = *t.Created
	}
	if t.Modified != nil && t.Modified.After(a.newest) {
		a.newest = *t.Modified
	}
	cpaths, err := a.f.Path(t)
	if err != nil {
		return fmt.Errorf("cannot get canonical path for index: %w", err)
	}
	for depth := range len(cpaths) {
		if a.MinPath != nil && depth < *a.MinPath {
			continue
		}
		if a.MaxPath != nil && depth > *a.MaxPath {
			continue
		}
		idxpath := filepath.Join(cpaths[:depth]...)
		doc := a.indexes[idxpath]
		if doc == nil {
			doc, err = a.document()
			if err != nil {
				return fmt.Errorf("cannot create RSS feed document at %q: %w", idxpath, err)
			}
			a.indexes[idxpath] = doc
		}
		doc.Items = append(doc.Items, &RSSFeedItem{
			Title:   *t.Title,
			Link:    doc.baseURL.JoinPath(a.f.GetPath(t.Id.String())).String(), // TODO - gross
			PubDate: (*RSSDate)(t.Created),
			Id:      t.Id,
		})
	}
	return nil
}

func (a *RSSAggregator) Close() error {
	var err error
	for p, doc := range a.indexes {
		//slices.SortFunc(ts, func(a *Text, b *Text) int { return b.Created.Compare(*a.Created) })
		fp := a.g.Create(p, *a.Filename).As(strptr("application/rss+xml")).At(&a.newest)
		err = xml.NewEncoder(fp).Encode(doc)
		if err != nil {
			return fmt.Errorf("failed to write RSS feed document for %q: %w", p, err)
		}
		err = fp.Close()
		if err != nil {
			return fmt.Errorf("failed to close RSS feed document for %q: %w", p, err)
		}
	}
	return nil
}
