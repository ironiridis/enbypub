package main

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"

	"github.com/alexflint/go-arg"
)

var args struct {
	Root            string         `arg:"--root,-d" default:"." placeholder:"DIR" help:"Base folder to work in"`
	PublicDir       string         `arg:"--pub,-p" default:"public" placeholder:"DIR" help:"Generated content is created in this folder relatve to root"`
	ContentDir      string         `arg:"--content,-c" default:"content" placeholder:"DIR" help:"Content is generated from files in this folder relative to root"`
	TextFilePattern *regexp.Regexp `arg:"--textfilepattern" default:"\\.md$" placeholder:"REGEX" help:"A regular expression for matching Text files relative to the content dir"`
	FeedsYaml       string         `arg:"--feeds,-f" default:"_feeds.yaml" placeholder:"FILE" help:"File relative to root where feeds are defined"`
}

var rootDir fs.FS

func init() {
	arg.MustParse(&args)
	if stat, err := os.Lstat(args.Root); err != nil {
		panic(fmt.Errorf("cannot use %q as a root: %w", args.Root, err))
	} else if !stat.Mode().IsDir() {
		panic(fmt.Errorf("cannot use %q as a root: not a directory", args.Root))
	}
	rootDir = os.DirFS(args.Root)
}
