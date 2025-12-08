package domain

import (
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/vupdivup/recital/internal/data"
	"github.com/vupdivup/recital/pkg/fileutils"
	"github.com/vupdivup/recital/pkg/git"
	"github.com/vupdivup/recital/pkg/random"
	"github.com/vupdivup/recital/pkg/random/lazy"
	"github.com/vupdivup/recital/pkg/tokenizer"
	"go.uber.org/zap"
)

var (
	// dbId is the identifier of the current token database, static throughout
	// the app lifecycle.
	dbId         string
	hasTokenized bool = false
)

const (
	maxFileSize = 1_000_000 // 1 MB

	minTokenLen = 2
	maxTokenLen = 12

	maxWordLen = 24

	tokenBufferSize = 10000
)

// Prompt generates a prompt of the maximum specified character length
// from tokens of the specified directory.
func Prompt(dirPath string, maxLen int) (string, error) {
	zap.S().Infow("Generating prompt from directory text content",
		"dir_path", dirPath,
		"max_len", maxLen)

	// Normalize path
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		zap.S().Errorw("Failed to get absolute path",
			"dir_path", dirPath,
			"error", err)
		return "", err
	}
	dbId = absPath

	if !hasTokenized {
		// Tokenize directory if not done already (1st call)
		if err := tokenizeDirectory(absPath); err != nil {
			return "", err
		}
		hasTokenized = true
	}

	return generatePrompt(maxLen)
}

// tokenizeDirectory tokenizes all eligible files in the specified directory
// and stores the tokens in the database.
func tokenizeDirectory(dirPath string) error {
	tokens := []data.Token{}
	changedFiles := []data.File{}
	newFiles := []data.File{}

	dbFilesTmp, err := data.GetFiles(dbId)
	if err != nil {
		return err
	}

	dbFiles := map[string]data.File{}
	removedFiles := map[string]data.File{}
	for _, dbFile := range dbFilesTmp {
		dbFiles[dbFile.Path] = dbFile
		removedFiles[dbFile.Path] = dbFile
	}

	flushTokens := func() error {
		// Delete tokens and file entries for changed or new files to purge
		// token sets
		for _, file := range changedFiles {
			// TODO: delete tokens only?
			if err := data.DeleteFile(dbId, file, true); err != nil {
				return err
			}
		}

		// (Re-)upload files previously deleted
		if err := data.UpsertFiles(dbId, changedFiles); err != nil {
			return err
		}

		// Upload new files
		if err := data.UpsertFiles(dbId, newFiles); err != nil {
			return err
		}

		if len(tokens) == 0 {
			return nil
		}

		if err := data.UpsertTokens(dbId, tokens); err != nil {
			return err
		}

		changedFiles = changedFiles[:0]
		newFiles = newFiles[:0]
		tokens = tokens[:0]
		return nil
	}

	// Get files in directory, recursively
	paths, err := git.LsFiles(dirPath)
	if err != nil {
		return err
	}

	for _, filepath := range paths {
		// Exclude unwanted files
		if isDesired, err := isFileEligible(filepath); err != nil {
			return err
		} else if !isDesired {
			zap.S().Debugw("Skipping ineligible file",
				"file_path", filepath)
			continue
		}

		// Calculate size and mtime
		size, err := fileutils.Size(filepath)
		if err != nil {
			return err
		}
		mtime, err := fileutils.Mtime(filepath)
		if err != nil {
			return err
		}

		file := data.File{Path: filepath, Size: size, Mtime: mtime}

		if dbFile, ok := dbFiles[filepath]; ok {
			// File from db still exists, so remove from deletion list
			delete(removedFiles, filepath)

			hasFileChanged, err := hasFileChanged(dbFile)
			if err != nil {
				return err
			}

			if !hasFileChanged {
				// File unchanged, skip tokenization
				zap.S().Debugw("Skipping unchanged file",
					"file_path", filepath)
				continue
			} else {
				// File changed since last tokenization
				zap.S().Debugw("Re-tokenizing changed file",
					"file_path", filepath)
				changedFiles = append(changedFiles, file)
			}
		} else {
			// New file
			zap.S().Debugw("Tokenizing new file",
				"file_path", filepath)
			newFiles = append(newFiles, file)
		}

		// Tokenize file
		fileTokens, err := tokenizer.TokenizeFile(filepath, isWordEligible)
		if err != nil {
			return err
		}

		// Retrieve first occurence of each token
		lookup := map[string]bool{}
		for _, fileToken := range fileTokens {
			if _, ok := lookup[fileToken]; isTokenEligible(fileToken) && !ok {
				tokens = append(
					tokens, data.Token{Path: filepath, Value: fileToken})
				lookup[fileToken] = true
			}
		}

		if len(tokens) > tokenBufferSize {
			if err := flushTokens(); err != nil {
				return err
			}
		}
	}

	// Flush remaining tokens
	if err := flushTokens(); err != nil {
		return err
	}

	// Delete tokens and entries of files that don't exist anymore
	for _, file := range removedFiles {
		zap.S().Infow("Deleting removed file from database",
			"file_path", file.Path)
		if err := data.DeleteFile(dbId, file, true); err != nil {
			return err
		}
	}

	return nil
}

// generatePrompt creates a prompt of up to maxLen characters by randomly
// sampling tokens from the database.
func generatePrompt(maxLen int) (string, error) {
	// Estimate max number of words needed to reach maxLen
	maxWordsNeeded := int(
		math.Round(float64(maxLen+1) / float64((minTokenLen + 1))))

	// Get random tokens, sample more than needed to account for length cutoff
	tokenResults := lazy.Sample(data.IterUniqueTokens(dbId), maxWordsNeeded)
	tokens := []string{}
	for _, tr := range tokenResults {
		if tr.Err != nil {
			return "", tr.Err
		}
		tokens = append(tokens, tr.Token.Value)
	}

	// Shuffle tokens to ensure randomness
	shuffled := random.Shuffle(tokens)

	// Select tokens in shuffle order until reaching maxLen
	promptLen := 0
	promptTokens := []string{}
	for i, token := range shuffled {
		// Account for space before token if not the first one
		if i > 0 {
			promptLen++
		}

		promptLen += len([]rune(token))
		if promptLen > maxLen {
			break
		}
		promptTokens = append(promptTokens, token)
	}

	return strings.Join(promptTokens, " "), nil
}

// isFileEligible returns true if the file should be included for tokenization.
func isFileEligible(fpath string) (bool, error) {
	stat, err := os.Stat(fpath)
	if err != nil {
		return false, err
	}

	isTextFile, err := fileutils.IsTextFile(fpath)
	if err != nil {
		return false, err
	}

	return isTextFile && stat.Size() < maxFileSize, nil
}

// isTokenEligible returns true if the token should be included for tokenization.
func isTokenEligible(token string) bool {
	runes := []rune(token)
	return minTokenLen < len(runes) && len(runes) < maxTokenLen
}

// isWordEligible returns true if the word should be included for tokenization.
func isWordEligible(word string) bool {
	runes := []rune(word)
	return len(runes) < maxWordLen
}

// hasFileChanged checks if the file has changed compared to the stored
// metadata.
// It compares size and modification time.
func hasFileChanged(file data.File) (bool, error) {
	size, err := fileutils.Size(file.Path)
	if err != nil {
		return false, err
	}

	mtime, err := fileutils.Mtime(file.Path)
	if err != nil {
		return false, err
	}
	return size != file.Size || mtime.Unix() != file.Mtime.Unix(), nil
}
