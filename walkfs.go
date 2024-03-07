package main

import (
	"fmt"
	"io/fs"

	enbypub "github.com/ironiridis/enbypub/enbypublib"
)

func WalkContent() (Texts, error) {
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
	T := make(Texts)
	for i := range files {
		t, err := enbypub.LoadTextFromFile(files[i])
		if err != nil {
			return nil, fmt.Errorf("encountered an error while scanning content directory %q: %w", args.ContentDir, err)
		}
		T[*t.Id] = t
	}
	return T, nil
}
