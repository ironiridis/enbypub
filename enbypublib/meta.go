package enbypub

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"time"
)

// MetaT is a placeholder for metadata about the enbypub library generating the content
type MetaT struct {
	// Ver is the version of the enbypub package if available, or the main package
	Ver *string

	// Commit is either the abbreviated git sha-1 commit id, or the whole revision if it doesn't look like a git commit
	Commit *string

	// BuildTime is the reported build time of the main package
	BuildTime *time.Time

	// Package is the enbypub library package path
	Package string

	// MainPackage is the main package path (ie the repo importing this package)
	MainPackage string
}

var metaCached *MetaT

func Meta() (M *MetaT) {
	if metaCached != nil {
		M = metaCached
		return
	}
	M = &MetaT{
		Package:     "github.com/ironiridis/enbypub/enbypublib",
		MainPackage: "github.com/ironiridis/enbypub",
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		M.Package = bi.Path
		M.MainPackage = bi.Main.Path

		// take the main binary version by default, but...
		M.Ver = &bi.Main.Version
		// ... scan the deps of main for specifically enbypub and report that
		// version if its available in the dep list
		for dep := range bi.Deps {
			if bi.Deps[dep].Path == M.Package {
				M.Ver = &bi.Deps[dep].Version
			}
		}
		for i := range bi.Settings {
			switch bi.Settings[i].Key {
			case "vcs.revision":
				// if this looks like a git sha-1, extract a git commit abbreviation
				shamatch := regexp.MustCompile(`^([0-9a-fA-F]{7})[0-9a-fA-F]{33}$`)
				if m := shamatch.FindStringSubmatch(bi.Settings[i].Value); len(m) > 1 {
					M.Commit = &m[1]
				} else {
					// if it doesn't look like a git sha-1, just report the whole thing
					M.Commit = &bi.Settings[i].Value
				}
			case "vcs.time":
				if t, err := time.Parse(time.RFC3339, bi.Settings[i].Value); err == nil {
					M.BuildTime = &t
				}
			}
		}
	}
	metaCached = M
	return
}

// Generator returns a string describing the specific build of either enbypub or
// whatever is importing it.
func (M *MetaT) Generator() string {
	if M.Ver != nil {
		return fmt.Sprintf("enbypub/%v %v", *M.Ver, M.MainPackage)
	}
	if M.Commit != nil {
		if M.BuildTime != nil {
			return fmt.Sprintf("enbypub@%v built %v %v", *M.Commit, M.BuildTime.UTC(), M.MainPackage)
		}
		return fmt.Sprintf("enbypub@%v %v", *M.Commit, M.MainPackage)
	}
	return fmt.Sprintf("enbypub (no build info) %v", M.MainPackage)
}
