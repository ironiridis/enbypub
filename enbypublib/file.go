package enbypub

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// File holds an open os.File pointer representing a pending piece of written content.
type File struct {
	g           *Generator
	err         error
	path        string
	ospath      string
	fp          *os.File
	opens       uint
	contentType *string
	modtime     *time.Time
}

// As sets the intended content type of the output File
func (f *File) As(contentType *string) *File {
	if f.err != nil {
		return f
	}
	f.contentType = contentType
	return f
}

// At sets the intended modification time of the output File.
func (f *File) At(modtime *time.Time) *File {
	if f.err != nil {
		return f
	}
	f.modtime = modtime
	return f
}

// From reads the File contents from path. If the File's content type has not been
// set yet, it's guessed from the extension of path (if any). The File is Close()d.
func (f *File) From(path string) error {
	if f.err != nil {
		return f.err
	}
	if f.fp == nil {
		f.err = fmt.Errorf("cannot create %q from %q: destination file not open", f.path, path)
		return f.err
	}
	// No known content type yet?
	if f.contentType == nil {
		// Does path have an extension?
		if i := strings.LastIndexByte(path, '.'); i >= 0 {
			// Does that extension have a known content type?
			if ct := ContentTypeFromExtension(path[i:]); ct != "" {
				f.contentType = &ct
			}
		}
	}
	fp, err := os.Open(path)
	if err != nil {
		f.err = fmt.Errorf("cannot create %q: cannot open %q to read from: %w", f.path, path, err)
		return f.err
	}
	defer fp.Close()
	io.Copy(f, fp)
	return f.Close()
}

// Write wraps the underlying os.*File.Write method
func (f *File) Write(b []byte) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	if f.fp == nil {
		return 0, fmt.Errorf("cannot Write() to %q: file not open", f.path)
	}
	return f.fp.Write(b)
}

// WriteAt wraps the underlying os.*File.WriteAt method
func (f *File) WriteAt(b []byte, off int64) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	if f.fp == nil {
		return 0, fmt.Errorf("cannot WriteAt() to %q: file not open", f.path)
	}
	return f.fp.WriteAt(b, off)
}

// WriteString wraps the underlying os.*File.WriteString method
func (f *File) WriteString(s string) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	if f.fp == nil {
		return 0, fmt.Errorf("cannot WriteString() to %q: file not open", f.path)
	}
	return f.fp.WriteString(s)
}

// ReadFrom wraps the underlying os.*File.ReadFrom method
func (f *File) ReadFrom(r io.Reader) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	if f.fp == nil {
		return 0, fmt.Errorf("cannot ReadFrom() to %q: file not open", f.path)
	}
	return f.fp.ReadFrom(r)
}

// Seek wraps the underlying os.*File.Seek method
func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	if f.fp == nil {
		return 0, fmt.Errorf("cannot Seek() on %q: file not open", f.path)
	}
	return f.fp.Seek(offset, whence)
}

// Sync wraps the underlying os.*File.Sync method
func (f *File) Sync() error {
	if f.err != nil {
		return f.err
	}
	if f.fp == nil {
		return fmt.Errorf("cannot Sync() on %q: file not open", f.path)
	}
	return f.fp.Sync()
}

// Truncate wraps the underlying os.*File.Truncate method
func (f *File) Truncate(size int64) error {
	if f.err != nil {
		return f.err
	}
	if f.fp == nil {
		return fmt.Errorf("cannot Truncate() on %q: file not open", f.path)
	}
	return f.fp.Truncate(size)
}

// Close calls the underlying os.*File.Close() method, then sets the modification time
// specified with At(), if any. Close is always safe to call, even if the original file
// open failed.
func (f *File) Close() error {
	if f.err != nil {
		return f.err
	}
	if f.fp == nil { // f is already closed; return the last error set, if any
		return f.err
	}
	f.opens--
	if f.opens > 0 {
		return nil
	}
	f.err, f.fp = f.fp.Close(), nil // fp must always be set to nil
	if f.err != nil {
		f.err = fmt.Errorf("failed to Close() %q: %w", f.path, f.err)
		return f.err
	}
	if f.modtime != nil {
		if f.err = os.Chtimes(f.ospath, time.Time{}, *f.modtime); f.err != nil {
			f.err = fmt.Errorf("failed to set modification time on %q to %v: %w", f.path, *f.modtime, f.err)
			return f.err
		}
	}
	return nil
}

// Err returns the most recent error this File encountered.
func (f *File) Err() error {
	return f.err
}
