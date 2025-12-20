package files

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTextFile(t *testing.T) {
	cases := []struct {
		relPath string
		want    bool
	}{
		{"testdata/is_text_file/binary.png", false},
		{"testdata/is_text_file/empty.md", false},
		{"testdata/is_text_file/no_extension", true},
		{"testdata/is_text_file/standard.txt", true},
		{"testdata/is_text_file/fake.txt", false},
	}
	for _, c := range cases {
		// Relative path check
		got, err := IsTextFile(c.relPath)
		assert.NoError(t, err)
		assert.Equal(t, c.want, got)

		// Absolute path check
		absPath, err := filepath.Abs(c.relPath)
		assert.NoError(t, err)
		got, err = IsTextFile(absPath)
		assert.NoError(t, err)
		assert.Equal(t, c.want, got)

		// Non-existent file check
		_, err = IsTextFile("nonexistent/file/path.txt")
		assert.Error(t, err)
	}
}

func TestDirExists(t *testing.T) {
	cases := []struct {
		relPath string
		want    bool
	}{
		{"testdata/dir_exists/parent", true},
		{"testdata/dir_exists/parent/child", true},
		{"testdata/dir_exists/nonexistent", false},
		{"/nonexistent/absolute/path", false},
	}

	for _, c := range cases {
		// Relative path check
		relExists, err := DirExists(c.relPath)
		assert.NoError(t, err)
		assert.Equal(t, c.want, relExists)

		// Absolute path check
		absPath, err := filepath.Abs(c.relPath)
		assert.NoError(t, err)
		absExists, err := DirExists(absPath)
		assert.NoError(t, err)
		assert.Equal(t, c.want, absExists)
	}
}

func TestRemoveChildren(t *testing.T) {
	dirs := []string{
		"testdata/remove_children/dir",
		"testdata/remove_children/nested",
		"testdata/remove_children/nested/subdir",
		"testdata/remove_children/empty",
	}
	files := []string{
		"testdata/remove_children/dir/file1.txt",
		"testdata/remove_children/dir/file2.log",
		"testdata/remove_children/nested/file3.md",
		"testdata/remove_children/nested/subdir/file4.csv",
	}

	// Setup: create directories and files
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0o755)
		assert.NoError(t, err)
	}
	for _, file := range files {
		f, err := os.Create(file)
		assert.NoError(t, err)
		f.Close()
	}

	// Test removal
	err := RemoveChildren("testdata/remove_children")
	assert.NoError(t, err)
	entries, err := os.ReadDir("testdata/remove_children")
	assert.NoError(t, err)
	assert.Empty(t, entries)
}
