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

func main() {
	Content := must1(WalkContent())
	Templates := must1(html.ParseFS(rootDir, args.TemplatesDir+"/*.html"))
	Feeds := must1(enbypub.LoadFeedsFromFile(args.FeedsYaml))
	must("populate feeds", Feeds.Scan(Content))

	Meta := enbypub.Meta()
	for _, F := range Feeds {
		F.SortByCreated()
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
	}
}
