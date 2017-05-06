package fileinput

import (
	"context"
	"io"
	"io/ioutil"
	"strings"
)

// Cat returns an io.ReadCloser that delivers the logical concatenation of the
// specified input files. The files are read sequentially, and any non-nil,
// non-EOF error from one of the underlying files is returned.  If no files are
// specified, the resulting reader is empty.
func Cat(ctx context.Context, paths []string) io.ReadCloser {
	if len(paths) == 0 {
		return ioutil.NopCloser(strings.NewReader(""))
	}
	return &catReader{ctx: ctx, paths: paths}
}

// CatOrFile acts as Cat, but if no paths are specified it returns f.
func CatOrFile(ctx context.Context, paths []string, f io.ReadCloser) io.ReadCloser {
	if len(paths) == 0 {
		return f
	}
	return &catReader{ctx: ctx, paths: paths}
}

type catReader struct {
	ctx   context.Context
	cur   io.ReadCloser // the file currently being read, or nil
	paths []string      // the paths remaining to be read
}

func (c *catReader) Read(data []byte) (int, error) {
	// If there is no reader active, try to open the next file.
	// When all files are exhausted, the reader is done.
	if c.cur == nil {
		if len(c.paths) == 0 {
			return 0, io.EOF
		}
		rc, err := Open(c.ctx, c.paths[0])
		c.paths = c.paths[1:]
		if err != nil {
			return 0, err
		}
		c.cur = rc
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

func (c *catReader) Close() error {
	var err error
	if c.cur != nil {
		err = c.cur.Close()
	}
	c.cur = nil
	c.paths = nil
	return err
}
