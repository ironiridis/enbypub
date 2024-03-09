package main

import (
	"fmt"
	"io"
	"os"
	"text/template"

	enbypub "github.com/ironiridis/enbypub/enbypublib"
)

func main() {
	T, err := WalkContent()
	fmt.Fprintf(os.Stderr, "err=%v\n", err)

	art, err := template.ParseFS(rootDir, "templates/*.html")
	if err != nil {
		panic(err)
	}

	Meta := enbypub.Meta()
	type Z struct {
		Feed *enbypub.Feed // stubbed for now
		Text *enbypub.Text
		Meta *enbypub.MetaT
	}
	for k := range T {
		z := Z{
			Feed: &enbypub.Feed{},
			Text: T[k],
			Meta: Meta,
		}
		var dest io.Writer
		if T[k].Slug != nil {
			fn := fmt.Sprintf("%s/%s/%s/%s.html", args.Root, args.PublicDir, *T[k].Style, *T[k].Slug)
			fp, err := os.Create(fn)
			if err != nil {
				panic(err)
			}
			defer fp.Close()
			dest = fp
		} else {
			dest = os.Stdout
		}
		err := art.Execute(dest, &z)
		if err != nil {
			panic(err)
		}
	}
}
