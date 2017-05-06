package fileinput

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

type fileMap map[string]string

var testFiles = fileMap{
	"a": "alice",
	"b": "basil",
	"c": "clara",
	"d": "desmond",
}

func (f fileMap) open(_ context.Context, path string) (io.ReadCloser, error) {
	s, ok := f[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return ioutil.NopCloser(strings.NewReader(s)), nil
}

func init() { Open = testFiles.open }

func stubStdin(s string) func() {
	// Replace os.Stdin with a pipe into which s gets dumped.
	r, w, err := os.Pipe()
	if err != nil {
		log.Fatalf("Pipe: %v", err)
	}
	go func() {
		io.WriteString(w, s)
		w.Close()
	}()
	saved := os.Stdin
	os.Stdin = r
	return func() { os.Stdin = saved }
}

const testWant = "alicebasilclaradesmond"

func TestEach(t *testing.T) {
	ctx := context.Background()
	var got []string
	var input = []string{"a", "b", "c", "d"}

	rf := func(r io.Reader, err error) error {
		if err != nil {
			return err
		} else if s, err := ioutil.ReadAll(r); err != nil {
			return err
		} else {
			got = append(got, string(s))
		}
		return nil
	}

	got = nil
	if err := Each(ctx, input, rf); err != nil {
		t.Errorf("Each: unexpected error: %v", err)
	} else if got := strings.Join(got, ""); got != testWant {
		t.Errorf("Each: got %q, want %q", got, testWant)
	}

	defer stubStdin(testWant)()

	got = nil
	if err := EachOrStdin(ctx, nil, rf); err != nil {
		t.Errorf("EachOrStdin: unexpected error: %v", err)
	} else if got := strings.Join(got, ""); got != testWant {
		t.Errorf("EachOrStdin: got %q, want %q", got, testWant)
	}
}

func TestCat(t *testing.T) {
	rc := Cat(context.Background(), []string{"a", "b", "c", "d"})
	bits, err := ioutil.ReadAll(rc)
	rc.Close()
	if err != nil {
		t.Errorf("ReadAll failed: %v", err)
	}
	if got := string(bits); got != testWant {
		t.Errorf("Cat: got %q, want %q", got, testWant)
	}
}

func TestCatOrFile(t *testing.T) {
	defer stubStdin(testWant)()

	rc := CatOrFile(context.Background(), nil, os.Stdin)
	bits, err := ioutil.ReadAll(rc)
	rc.Close()
	if err != nil {
		t.Errorf("ReadAll failed: %v", err)
	}
	if got := string(bits); got != testWant {
		t.Errorf("CatOrFile: got %q, want %q", got, testWant)
	}
}
