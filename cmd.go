package main

import (
	"fmt"
	html "html/template"
	"os"

	enbypub "github.com/ironiridis/enbypub/enbypublib"
)

type Publish struct {
	Feed *enbypub.Feed
	Text *enbypub.Text
	Meta *enbypub.MetaT
}

type Index struct {
	Meta          *enbypub.MetaT
	Feed          *enbypub.Feed
	FeedStructure *enbypub.FeedStructure
	Index         []*enbypub.Text
	Seg           string
}

func main() {
	Generator := must1(enbypub.NewGenerator(args.PublicDir))
	Content := must1(WalkContent())
	Generator.Templates = must1(html.ParseFS(rootDir, args.TemplatesDir+"/*.html"))
	Feeds := must1(enbypub.LoadFeedsFromFile(args.FeedsYaml, Generator))
	must("populate feeds", Feeds.Scan(Content))

	for _, F := range Feeds {
		F.SortByCreatedDescending()
		CS := must1(F.CanonicalStructure())
		must("build directory structure", EnsurePath(CS))
		for fn, T := range CS.Files {
			must("generate output file", Generator.Template(
				must1(F.Get(enbypub.TextAttributeTemplate, T)),
				&Publish{Feed: F, Text: T, Meta: enbypub.Meta()},
				T.Modified,
				fn))
		}
		F.Close()
	}
	fmt.Fprintf(os.Stdout, "generated files:\n%+v\n", Generator.Manifest())
}
