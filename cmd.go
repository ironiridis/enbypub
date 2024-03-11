package main

import (
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
}

func main() {
	Content := must1(WalkContent())
	Templates := must1(html.ParseFS(rootDir, args.TemplatesDir+"/*.html"))
	Feeds := must1(enbypub.LoadFeedsFromFile(args.FeedsYaml))
	must("populate feeds", Feeds.Scan(Content))

	Meta := enbypub.Meta()
	for _, F := range Feeds {
		F.SortByCreatedDescending()
		CS := must1(F.CanonicalStructure())
		must("build directory structure", EnsurePath(CS))
		for fn, T := range CS.Files {
			fp := must1(os.Create(args.Root + "/" + args.PublicDir + "/" + fn))
			render := must1(F.Get(enbypub.TextAttributeTemplate, T))
			must("render text", Templates.ExecuteTemplate(fp, render, &Publish{
				Feed: F,
				Text: T,
				Meta: Meta,
			}))
			must("close output file", fp.Close())
		}
		for seg := range CS.Segments {
			fp := must1(os.Create(args.Root + "/" + args.PublicDir + "/" + seg + "/index.html"))
			must("render index", Templates.ExecuteTemplate(fp, "index.html", &Index{
				Meta:          Meta,
				Feed:          F,
				FeedStructure: CS,
				Index:         CS.Segments[seg],
			}))
		}
	}
}
