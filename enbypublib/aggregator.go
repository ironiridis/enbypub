package enbypub

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v2"
)

type Aggregator interface {
	Init(*Feed, *Generator) error
	AddText(*Text) error
	Close() error
}

type Aggregators []Aggregator

func (a *Aggregators) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var tmp []*genericAggregator
	err = unmarshal(&tmp)
	if err != nil {
		return
	}
	*a = make([]Aggregator, len(tmp))
	for i := range tmp {
		(*a)[i], err = tmp[i].specialize()
		if err != nil {
			return
		}
	}
	return nil
}

type genericAggregator struct {
	Kind       string
	Attributes map[string]any `yaml:",inline"`
}

func (ga *genericAggregator) specialize() (a Aggregator, err error) {
	switch ga.Kind {
	case "index":
		a = &IndexAggregator{Kind: ga.Kind}
	case "rss":
		a = &RSSAggregator{Kind: ga.Kind}
	default:
		err = fmt.Errorf("cannot specialize into an unknown aggregator kind %q", ga.Kind)
		return
	}

	var b bytes.Buffer
	if err = yaml.NewEncoder(&b).Encode(ga); err != nil {
		err = fmt.Errorf("cannot re-encode generic aggregator: %w", err)
		return
	}

	if err = yaml.NewDecoder(&b).Decode(a); err != nil {
		err = fmt.Errorf("cannot decode into %T: %w", a, err)
		return
	}
	return
}

// AtomAggregator provides an Atom feed XML document intended for consumption with an Atom feed reader.
type AtomAggregator struct {
	MinPath *uint `yaml:",omitempty"`
	MaxPath *uint `yaml:",omitempty"`
}

// SitemapAggregator generates a sitemap XML document intended for consumption by search engines. It also
// adds a pointer to this generated sitemap file in robots.txt.
type SitemapAggregator struct {
}

// RobotsExcludeAggregator adds the path component to an exclusion directive in robots.txt.
type RobotsExcludeAggregator struct {
	MinPath *uint `yaml:",omitempty"`
}

// SearchAggregator produces a basic search index of keywords to enable a client-side search.
type SearchAggregator struct {
	IndexJSONPath     *string `yaml:",omitempty"`
	IndexHTMLPath     *string `yaml:",omitempty"`
	IndexHTMLTemplate *string `yaml:",omitempty"`
}
