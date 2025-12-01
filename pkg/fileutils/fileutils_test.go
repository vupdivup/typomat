package fileutils

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.NoError(t, err)

		got, err := IsTextFile(absPath)
		assert.NoError(t, err)

		assert.Equal(t, c.want, got)
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
		assert.NoError(t, err)

		got, err := GetFingerprint(absPath)
		assert.NoError(t, err)

		assert.Equal(t, c.want, got)
	}
}
