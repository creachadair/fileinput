package fileinput_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/creachadair/fileinput"
)

// createTestFiles populates a temp directory with the given files, and returns
// the directory path. The directory will be cleaned up when t ends.
func createTestFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "fileinput")
	if err != nil {
		t.Fatalf("Creating testdata directory: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0644); err != nil {
			t.Fatalf("Writing testdata: %v", err)
		}
	}
	return dir
}

func TestCat(t *testing.T) {
	dir := createTestFiles(t, map[string]string{
		"a": "alice",
		"b": "basil",
		"c": "clara",
		"d": "desmond",
	})

	const testWant = "alicebasilclaradesmond"
	c := fileinput.Cat([]string{
		filepath.Join(dir, "a"),
		filepath.Join(dir, "b"),
		filepath.Join(dir, "c"),
		filepath.Join(dir, "d"),
	})
	bits, err := io.ReadAll(c)
	cerr := c.Close()
	if err != nil {
		t.Errorf("ReadAll failed: %v", err)
	}
	if cerr != nil {
		t.Errorf("Close failed: %v", err)
	}
	if got := string(bits); got != testWant {
		t.Errorf("Cat: got %q, want %q", got, testWant)
	}
}
