package git

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLsFiles(t *testing.T) {
	expectedFiles := []string{
		"testdata/.gitignore",
		"testdata/README.md",
		"testdata/config.json",
		"testdata/main.go",
		"testdata/src/app.go",
		"testdata/src/utils.go",
	}

	absPaths := []string{}
	for _, f := range expectedFiles {
		absPath, err := filepath.Abs(f)
		assert.NoError(t, err)
		absPaths = append(absPaths, absPath)
	}

	files, err := LsFiles("testdata")
	assert.NoError(t, err)
	assert.ElementsMatch(t, absPaths, files)
}
