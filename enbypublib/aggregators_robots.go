package enbypub

import (
	"fmt"
	"strings"
	"text/template"
)

var robotsExclusions = must1(template.New("exclusions").Parse(
	`{{ range $path, $exclude := . }}User-Agent: *
{{ if $exclude }}Disallow{{ else }}Allow{{ end }}: /{{ $path }}
{{ end }}
`))

// RobotsExcludeAggregator adds the path component to an exclusion directive in robots.txt.
type RobotsExcludeAggregator struct {
	f *Feed

	// Kind is always "robotsexclude"
	Kind    string
	MinPath int `yaml:",omitempty"`

	robots *File
	paths  map[string]bool
}

func (rea *RobotsExcludeAggregator) Init(f *Feed, g *Generator) error {
	rea.paths = make(map[string]bool)
	rea.f = f
	rea.robots = g.Create("robots.txt")
	return nil
}

func (rea *RobotsExcludeAggregator) AddText(t *Text) error {
	p, err := rea.f.Path(t)
	if err != nil {
		return fmt.Errorf("cannot determine path for robots exclusion: %w", err)
	}
	rea.paths[strings.Join(p[:rea.MinPath], "/")] = true
	return nil
}

func (rea *RobotsExcludeAggregator) Close() error {
	if err := robotsExclusions.Execute(rea.robots, rea.paths); err != nil {
		return fmt.Errorf("failed to write robots.txt lines: %w", err)
	}
	return rea.robots.Close()
}
