// Package git provides utilities for interacting with Git repositories.
package git

import (
	"os"
	"path/filepath"

	"github.com/sabhiram/go-gitignore"
)

// getGitignore locates the first .gitignore file in specified repository path,
// returning an error if no such file is found.
func getGitignore(repoPath string) (string, error) {
	files, err := filepath.Glob(filepath.Join(repoPath, "*.gitignore"))
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", os.ErrNotExist
	}

	return files[0], nil
}

// LsFiles lists absolute file paths of all files in the specified repository
// path, excluding those ignored by .gitignore. Only .gitignore files in the
// root of the repository are considered. Symlinks are listed as files.
//
// Note that the target directory is scanned regardless of whether it contains a
// .git subdirectory. This means LsFiles returns files for any directory, not
// just Git repositories.
func LsFiles(rootPath string) ([]string, error) {
	// Check for .gitignore file
	gitignorePath, err := getGitignore(rootPath)
	hasGitignore := err == nil

	// Compile .gitignore if it exists
	var gitignore *ignore.GitIgnore
	if hasGitignore {
		gitignore, err = ignore.CompileIgnoreFile(gitignorePath)
		if err != nil {
			return nil, err
		}
	}

	files := []string{}
	walk := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		} else if !d.IsDir() {
			relPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return err
			}

			if !hasGitignore || !gitignore.MatchesPath(relPath) {
				// Append absolute path
				absPath, err := filepath.Abs(filepath.Join(rootPath, relPath))
				if err != nil {
					return err
				}
				files = append(files, absPath)
			}
		}

		return nil
	}

	// Walk the repository directory
	err = filepath.WalkDir(rootPath, walk)
	if err != nil {
		return nil, err
	}

	return files, nil
}
