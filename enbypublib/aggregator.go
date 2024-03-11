package enbypub

// IndexAggregator provides an HTML article index intended for consumption with a web browser.
type IndexAggregator struct {
	MinPath  *uint `yaml:",omitempty"`
	MaxPath  *uint `yaml:",omitempty"`
	Paginate *uint `yaml:",omitempty"`
	Template string
}

// RSSAggregator provides an RSS feed XML document intended for consumption with an RSS reader.
type RSSAggregator struct {
	MinPath *uint `yaml:",omitempty"`
	MaxPath *uint `yaml:",omitempty"`
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
