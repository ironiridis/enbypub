package enbypub

import (
	"fmt"
	"path/filepath"
	"slices"
	"time"
)

// IndexAggregator provides an HTML article index intended for consumption with a web browser.
type IndexAggregator struct {
	f *Feed
	g *Generator

	// Kind is always "index"
	Kind     string
	MinPath  *int    `yaml:",omitempty"`
	MaxPath  *int    `yaml:",omitempty"`
	Paginate *int    `yaml:",omitempty"` // TODO
	Filename *string `yaml:",omitempty"`
	Template *string `yaml:",omitempty"`
	Sort     *string `yaml:",omitempty"`

	newest time.Time
	index  []*Text

	// indexes is a map of path components to Text lists
	indexes map[string][]*Text
}

type IndexAggregatorContent struct {
	Meta  *MetaT
	Feed  *Feed
	Index []*Text
}

func (ia *IndexAggregator) Init(f *Feed, g *Generator) error {
	ia.f = f
	ia.g = g
	if ia.Sort == nil {
		ia.Sort = strptr("newest-first")
	}
	if ia.Template == nil {
		ia.Template = strptr("index.html")
	}
	if ia.Filename == nil {
		ia.Filename = strptr("index.html")
	}
	ia.index = make([]*Text, 0, len(f.Index))
	ia.indexes = make(map[string][]*Text)
	return nil
}

func (ia *IndexAggregator) AddText(t *Text) error {
	if t.Created != nil && t.Created.After(ia.newest) {
		ia.newest = *t.Created
	}
	if t.Modified != nil && t.Modified.After(ia.newest) {
		ia.newest = *t.Modified
	}
	cpaths, err := ia.f.Path(t)
	if err != nil {
		return fmt.Errorf("cannot get canonical path for index: %w", err)
	}
	for depth := range len(cpaths) {
		if ia.MinPath != nil && depth < *ia.MinPath {
			continue
		}
		if ia.MaxPath != nil && depth > *ia.MaxPath {
			continue
		}
		idxpath := filepath.Join(cpaths[:depth]...)
		ia.indexes[idxpath] = append(ia.indexes[idxpath], t)
	}
	return nil
}

func (ia *IndexAggregator) Close() (err error) {
	for p, ts := range ia.indexes {
		switch *ia.Sort {
		case "newest-first":
			slices.SortFunc(ts, func(a *Text, b *Text) int { return b.Created.Compare(*a.Created) })
		case "oldest-first":
			slices.SortFunc(ts, func(a *Text, b *Text) int { return a.Created.Compare(*b.Created) })
		}
		err = ia.g.Template(*ia.Template, &IndexAggregatorContent{
			Meta:  Meta(),
			Feed:  ia.f,
			Index: ts,
		}, &ia.newest, p, *ia.Filename)
		if err != nil {
			break
		}
	}
	return
}
