package main

import (
	"fmt"
	"io/fs"
	"os"

	enbypub "github.com/ironiridis/enbypub/enbypublib"
)

func WalkContent() (enbypub.Texts, error) {
	files := []string{}
	err := fs.WalkDir(rootDir, args.ContentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if args.TextFilePattern.MatchString(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("encountered an error while walking content directory %q: %w", args.ContentDir, err)
	}
	T := make(enbypub.Texts)
	for i := range files {
		t, err := enbypub.LoadTextFromFile(files[i])
		if err != nil {
			return nil, fmt.Errorf("encountered an error while scanning content directory %q: %w", args.ContentDir, err)
		}
		T[*t.Id] = t
	}
	return T, nil
}

func EnsurePath(fs *enbypub.FeedStructure) error {
	for seg := range fs.Segments {
		err := os.MkdirAll(args.Root+"/"+args.PublicDir+"/"+seg, 0777)
		if err != nil {
			return fmt.Errorf("cannot create path for %q: %w", seg, err)
		}
	}
	return nil
}
