package fileutils

import (
	"path/filepath"
	"testing"
)

func TestIsTextFile(t *testing.T) {
	basePath := "./../../test/data/fileutils/is_text_file"
	cases := []struct {
		path string
		want bool
	}{
		{"binary.png", false},
		{"empty.md", false},
		{"no_extension", true},
		{"standard.txt", true},
	}
	for _, c := range cases {
		absPath, err := filepath.Abs(filepath.Join(basePath, c.path))
		if err != nil {
			t.Error(err.Error())
			continue
		}
		got, err := IsTextFile(absPath)
		if err != nil {
			t.Error(err.Error())
			continue
		}

		if got != c.want {
			t.Errorf("IsTextFile(%q) = %v; want %v", absPath, got, c.want)
		}
	}
}

func TestGetFilesInTree(t *testing.T) {
	basePath := "./../../test/data/fileutils/get_files_in_tree"
	cases := []struct {
		rootDir string
		want    []string
	}{
		{"empty", []string{}},
		{"flat", []string{"0.txt", "1.txt", "2.txt"}},
		{"nested", []string{"0.txt", "subdir/1.txt", "subdir/subsubdir/2.txt"}},
	}

	for _, c := range cases {
		absRootDir, err := filepath.Abs(filepath.Join(basePath, c.rootDir))
		if err != nil {
			t.Error(err.Error())
			continue
		}

		got, err := GetFilesInTree(absRootDir)

		if err != nil {
			t.Error(err.Error())
			continue
		}

		for i := range got {
			// Construct the expected absolute path for comparison
			wantAbsPath := filepath.Join(absRootDir, c.want[i])

			// Compare the obtained path with the expected absolute path
			if got[i] != wantAbsPath {
				t.Errorf(
					"GetFilesInTree(%q) = %v; want %v",
					absRootDir, got, wantAbsPath,
				)
			}
		}
	}
}
