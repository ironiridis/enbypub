package enbypub

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"html/template"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	md "github.com/yuin/goldmark"
	"gopkg.in/yaml.v2"
)

// TextUnlikelyCreationDate represents an arbitrary threshold where filesystem times before this
// point are not trusted. If enbypub examines a file with a modification time before this time,
// that modification time will be ignored and will instead be substituted with the current time.
var TextUnlikelyCreationDate = must1(time.Parse(time.RFC3339, "1993-08-31T23:59:59Z"))

// TextMetadataDelimiter indicates the text boundary between metadata and body in a Text. This is
// set to the traditional three (or more) dashes typically used for Markdown front matter.
var TextMetadataDelimiter = regexp.MustCompile(`(?m:^---+[\r\n]+)`)

type Text struct {
	// file is the os.File this Text was parsed from
	file *os.File

	// originalFilename is the original filename this Text was read from
	originalFilename string

	// raw is the unparsed Body of the Text
	raw []byte

	// Title specifies the title of the Text
	Title *string `yaml:",omitempty"`

	// Slug is a generated URL-friendly string usually derived from the Title
	Slug *string `yaml:",omitempty"`

	// Created indicates the time this Text was first processed by enbypub
	Created *time.Time `yaml:",omitempty"`

	// Modified indicates the time this Text was most recently processed by enbypub
	Modified *time.Time `yaml:",omitempty"`

	// Id is the unique identifier assigned to this Text on first processing
	Id *uuid.UUID `yaml:",omitempty"`

	// Template specifies what template to use for rendering this Text
	Template *string `yaml:",omitempty"`

	// Tags specifies the distrubition of this Text
	Tags []string `yaml:",omitempty"`

	// Body is the parsed Markdown Document of the Text
	// Body *mda.Document `yaml:"-"`

	// Checksum determines whether the body of the Text has been changed since last processed
	Checksum *string `yaml:",omitempty"`
}

func (T *Text) Get(a Attribute) (string, error) {
	switch a {
	case TextAttributeSlug:
		if T.Slug != nil {
			return *T.Slug, nil
		}
	case TextAttributeId:
		if T.Id != nil {
			return T.Id.String(), nil
		}
	case TextAttributeTemplate:
		if T.Template != nil {
			return *T.Template, nil
		}
	case TextAttributeYear:
		if T.Created != nil {
			return (*T.Created).Format("2006"), nil
		}
	case TextAttributeMonth:
		if T.Created != nil {
			return (*T.Created).Format("01"), nil
		}
	case TextAttributeDay:
		if T.Created != nil {
			return (*T.Created).Format("02"), nil
		}
	case TextAttributeDayOfWeek:
		if T.Created != nil {
			return (*T.Created).Format("Mon"), nil
		}
	case TextAttributeYMD:
		if T.Created != nil {
			return (*T.Created).Format("20060102"), nil
		}
	}
	return "", fmt.Errorf("text does not have an attribute %q", string(a))
}

type Texts map[uuid.UUID]*Text

func (t Texts) Get(id string) *Text {
	if u, err := uuid.Parse(id); err == nil {
		return t[u]
	}
	return nil
}

func LoadTextFromFile(fn string) (*Text, error) {
	fstat, err := os.Stat(fn)
	if err != nil {
		return nil, fmt.Errorf("cannot stat file %q: %w", fn, err)
	}

	mt := fstat.ModTime()
	if mt.Before(TextUnlikelyCreationDate) {
		mt = time.Now()
	}

	fp, err := os.OpenFile(fn, os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("cannot open file %q: %w", fn, err)
	}
	defer fp.Close()
	T := Text{file: fp, originalFilename: fn}
	fbuf, err := io.ReadAll(fp)
	if err != nil {
		return nil, fmt.Errorf("cannot read contents of file %q: %w", fn, err)
	}

	delimpos := TextMetadataDelimiter.FindAllIndex(fbuf, 2)
	if len(delimpos) == 2 {
		err = yaml.Unmarshal(fbuf[delimpos[0][1]:delimpos[1][1]], &T)
		if err != nil {
			return nil, fmt.Errorf("cannot read metadata from %q: %w", fn, err)
		}
		T.raw = fbuf[delimpos[1][1]:]
	} else {
		T.raw = fbuf
	}

	if T.Created == nil {
		T.Created = &mt
	}

	// T.Body = mda.NewDocument()
	// T.Body.AppendChild(T.Body, md.DefaultParser().Parse(mdt.NewReader(T.raw)))

	// If the Text has been modified (or this is the first time we've seen it)
	if !T.ChecksumMatch() || (T.Modified != nil && T.Modified.Sub(mt).Abs() > time.Second*10) {
		// Set the Modified timestamp to the filesystem modification time
		T.Modified = &mt

		// If the current Created timestamp is later than the Modified timestamp, set that too
		if T.Created.After(*T.Modified) {
			T.Created = &mt
		}

		if err := T.Process(); err != nil {
			return nil, fmt.Errorf("cannot process %q: %w", fn, err)
		}

		if err := T.UpdateFile(); err != nil {
			return nil, fmt.Errorf("cannot update %q: %w", fn, err)
		}
	}

	return &T, nil
}

// ChecksumMatch calculates the hash of the raw Text body. If the hash doesn't exist in the
// Text value, or if the calculated hash doesn't match the Text value, ChecksumMatch updates
// the Text value and returns false. Otherwise, it returns true.
//
// The Checkum value of the Text should be opaque, but may match the regex:
//
//	(sha1|sha256|md5):[0-9a-f]{32,}
func (T *Text) ChecksumMatch() bool {
	var algo string
	var hash hash.Hash
	switch {
	case T.Checksum != nil && strings.HasPrefix(*T.Checksum, "sha256:"):
		algo = "sha256"
		hash = crypto.SHA256.New()
	case T.Checksum != nil && strings.HasPrefix(*T.Checksum, "md5:"):
		algo = "md5"
		hash = crypto.MD5.New()
	default:
		algo = "sha1"
		hash = crypto.SHA1.New()
	}
	hash.Write(T.raw)
	calc := algo + ":" + hex.EncodeToString(hash.Sum([]byte{}))
	if T.Checksum != nil && strings.EqualFold(calc, *T.Checksum) {
		return true
	}
	T.Checksum = &calc
	return false
}

func (T *Text) Process() error {
	if T.Id == nil {
		u := uuid.New()
		T.Id = &u
	}
	if T.Title == nil {
		T.Title = Titlenate(&T.originalFilename)
	}
	if T.Slug == nil {
		T.Slug = Sluggify(T.Title)
	}
	return nil
}

func (T *Text) UpdateFile() error {
	if T.file == nil {
		return errors.New("file handle is unexpectedly nil")
	}
	if _, err := T.file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("cannot reset to start of file: %w", err)
	}
	if err := T.file.Truncate(0); err != nil {
		return fmt.Errorf("cannot truncate file: %w", err)
	}
	if _, err := io.WriteString(T.file, "---\n"); err != nil {
		return fmt.Errorf("cannot write first metadata delimiter: %w", err)
	}
	if err := yaml.NewEncoder(T.file).Encode(T); err != nil {
		return fmt.Errorf("cannot re-write metadata: %w", err)
	}
	if _, err := io.WriteString(T.file, "---\n"); err != nil {
		return fmt.Errorf("cannot write second metadata delimiter: %w", err)
	}
	if _, err := T.file.Write(T.raw); err != nil {
		return fmt.Errorf("cannot re-write body: %w", err)
	}

	return nil
}

// IsModified returns true if the Modified attribute is at least 5 minutes after the Created attribute
func (T *Text) IsModified() bool {
	if T.Modified == nil {
		return false
	}
	return T.Modified.Sub(*T.Created) > time.Minute*5
}

// HTML returns an HTML fragment for the document body
func (T *Text) HTML() template.HTML {
	var B bytes.Buffer
	err := md.Convert(T.raw, &B)
	if err != nil {
		panic(err)
	}
	return template.HTML(B.String())
}

func (T Text) IsTagged(tag ...string) bool {
	for a := range tag {
		if slices.Contains(T.Tags, tag[a]) {
			return true
		}
	}
	return false
}
