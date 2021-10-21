// Package fileinput provides utility functions for handling files named on the
// command line.
package fileinput

import (
	"io"
	"os"
)

// A CatReader implements io.ReadCloser for the concatenation of a sequence of
// file paths. The files are opened and closed on-demand as the data are read.
// The caller must close the reader to release any open filehandle.
type CatReader struct {
	cur   io.ReadCloser
	paths []string
}

// Cat constructs a CatReader for the specified file paths. If no files are
// specified, the resulting reader is empty, returning io.EOF for all reads.
func Cat(paths []string) *CatReader { return &CatReader{paths: fixPaths(paths)} }

// Read implements the io.Reader interface. It reports io.EOF when all the
// files have been completely read and no further data are available.
func (c *CatReader) Read(data []byte) (int, error) {
	// If there is no reader active, try to open the next file.
	// When all files are exhausted, the reader is done.
	if c.cur == nil {
		if len(c.paths) == 0 {
			return 0, io.EOF
		}
		f, err := os.Open(c.paths[len(c.paths)-1])
		c.paths = c.paths[:len(c.paths)-1]
		if err != nil {
			return 0, err
		}
		c.cur = f
	}

	// Note that it is possible we may read 0 bytes without error.  This is
	// permitted by the definition of io.Reader, and will only happen if we
	// happen to already be at EOF from a previous read that did not report it.

	nr, err := c.cur.Read(data)
	if err == io.EOF {
		c.cur.Close()
		c.cur = nil
		return nr, nil
	}
	return nr, err
}

// Close implements the io.Closer interface. After closing c, any further reads
// will report io.EOF, even if there were unconsumed files prior to close.
func (c *CatReader) Close() error {
	var err error
	if c.cur != nil {
		err = c.cur.Close()
	}
	c.cur = nil
	c.paths = nil
	return err
}

func fixPaths(paths []string) []string {
	cp := make([]string, len(paths))
	copy(cp, paths)
	for i, j := 0, len(cp)-1; i < j; i++ {
		cp[i], cp[j] = cp[j], cp[i]
		j--
	}
	return cp
}
