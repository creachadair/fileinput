// Package fileinput provides utility functions for handling files named on the
// command line.
package fileinput

import (
	"context"
	"io"
	"os"
)

// An Opener function opens a file path for reading.
type Opener func(context.Context, string) (io.ReadCloser, error)

// Open is used to open files for reading. The default implementation delegates
// to the os.Open function.
var Open Opener = osOpener

func osOpener(_ context.Context, path string) (io.ReadCloser, error) { return os.Open(path) }

// A ReadFunc receives an open io.Reader or an error for processing.
// If err != nil, the value of r is unspecified.
type ReadFunc func(r io.Reader, err error) error

// Each opens each specified path for reading in the order given, passes its
// reader to rf, and closes the file when f returns.
//
// If there is an error opening a file it is passed to rf. If rf reports an
// error, Each stops processing and returns that error.
func Each(ctx context.Context, paths []string, rf ReadFunc) error {
	for _, path := range paths {
		rc, err := Open(ctx, path)
		ferr := rf(rc, err)
		if err == nil {
			rc.Close()
		}
		if ferr != nil {
			return ferr
		}
	}
	return nil
}

// EachOrStdin acts as Each, but if no paths are specified rf is called on
// os.Stdin instead.
func EachOrStdin(ctx context.Context, paths []string, rf ReadFunc) error {
	if len(paths) == 0 {
		return rf(os.Stdin, nil)
	}
	return Each(ctx, paths, rf)
}
