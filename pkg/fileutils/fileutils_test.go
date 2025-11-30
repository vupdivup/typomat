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

func TestGetFingerprint(t *testing.T) {
	basePath := "./../../test/data/fileutils/get_fingerprint"
	cases := []struct {
		path string
		want string
	}{
		{"empty.txt", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"standard.txt", "55bf0e806704c09fa5fe8a97e8ce90729024454a922c4e67b55ab93060e00338"},
		{"binary.png", "3eb10792d1f0c7e07e7248273540f1952d9a5a2996f4b5df70ab026cd9f05517"},
	}

	for _, c := range cases {
		absPath, err := filepath.Abs(filepath.Join(basePath, c.path))
		if err != nil {
			t.Error(err.Error())
			continue
		}

		got, err := GetFingerprint(absPath)
		if err != nil {
			t.Error(err.Error())
			continue
		}

		if got != c.want {
			t.Errorf("GetFingerprint(%q) = %v; want %v", absPath, got, c.want)
		}
	}
}