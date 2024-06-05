package enbypub

import (
	"fmt"
	html "html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Generator struct {
	Root      string
	Files     map[string]*File
	Templates *html.Template
}

func NewGenerator(root string) (*Generator, error) {
	fi, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("cannot use %q as a Generator root: %w", root, err)
	}
	if !fi.Mode().IsDir() {
		return nil, fmt.Errorf("cannot use %q as a Generator root: not a directory", root)
	}
	return &Generator{Root: root, Files: make(map[string]*File)}, nil
}

func (g *Generator) OSPath(path string) string {
	return filepath.Join(g.Root, path)
}

// Create returns a *File for writing that has many of the semantics of an os.*File.
// The method is intended to be chained together with other File methods. path may be a
// single single composed path (eg `Create("directory/file.txt")`) or a series of path
// components (eg `Create("directory", "file.txt")`) or a slice of path components (eg
// `Create(pathSlice...)`).
// The returned *File may not be ready to use. Any error on opening the file to create
// it will be stored in the *File and returned on subsequent writes.
func (g *Generator) Create(path ...string) (f *File) {
	p := filepath.Join(path...)

	// Does this Generator already have a *File at that path?
	if f = g.Files[p]; f != nil {
		f.opens++
		return
	}

	f = &File{g: g, path: p, ospath: g.OSPath(p), opens: 1}

	// Does path have an extension?
	if i := strings.LastIndexByte(p, '.'); i >= 0 {
		// Does that extension have a known content type?
		if ct := ContentTypeFromExtension(p[i:]); ct != "" {
			f.contentType = &ct
		}
	}

	if fp, err := os.Create(f.ospath); err != nil {
		f.err = fmt.Errorf("cannot create %q: %w", f.ospath, err)
	} else {
		f.fp = fp
	}
	g.Files[p] = f
	return
}

// Template is complicated
func (g *Generator) Template(template string, data any, mtime *time.Time, path ...string) error {
	fp := g.Create(path...).At(mtime)
	defer fp.Close()
	return g.Templates.ExecuteTemplate(fp, template, data)
}

// Manifest returns a list of each created file.
func (g *Generator) Manifest() (m []string) {
	m = make([]string, 0, len(g.Files))
	for k := range g.Files {
		m = append(m, filepath.Join(g.Root, k))
	}
	return
}
