// Package domain implements the core business logic of the application.
package domain

import (
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/vupdivup/typelines/internal/data"
	"github.com/vupdivup/typelines/pkg/fileutils"
	"github.com/vupdivup/typelines/pkg/git"
	"github.com/vupdivup/typelines/pkg/random"
	"github.com/vupdivup/typelines/pkg/random/lazy"
	"github.com/vupdivup/typelines/pkg/tokenizer"
	"go.uber.org/zap"
)

var (
	// dbId is the identifier of the current token database, static throughout
	// the app lifecycle.
	dbId string
	// hasTokenized indicates whether the directory has already been tokenized.
	// Set to true only once per app lifecycle.
	hasTokenized bool = false
)

const (
	// maxFileSize is the maximum file size (in bytes) eligible for
	// tokenization.
	maxFileSize = 1_000_000 // 1 MB

	// minTokenLen is the minimum length of a token to be included.
	minTokenLen = 2
	// maxTokenLen is the maximum length of a token to be included.
	maxTokenLen = 12

	// maxWordLen is the maximum length of a word to be considered for
	// tokenization.
	maxWordLen = 24

	// tokenBufferSize is the number of tokens to buffer before flushing to
	// the database.
	tokenBufferSize = 10000
)

// Prompt generates a prompt of the maximum specified character length
// from tokens of the specified directory.
func Prompt(dirPath string, maxLen int) (string, error) {
	zap.S().Infow("Generating prompt from directory text content",
		"dir_path", dirPath,
		"max_len", maxLen)

	// Check if directory exists
	dirExists, err := fileutils.DirExists(dirPath)
	if err != nil {
		zap.S().Errorw("Failed to check if directory exists",
			"dir_path", dirPath,
			"error", err)
		return "", ErrFileOperation
	}
	if !dirExists {
		zap.S().Errorw("Directory does not exist",
			"dir_path", dirPath)
		return "", ErrInvalidDirPath
	}

	// Normalize path
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		zap.S().Errorw("Failed to get absolute path",
			"dir_path", dirPath,
			"error", err)
		return "", ErrInvalidDirPath
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
		// Delete tokens of changed files for subsequent re-insertion
		for _, file := range changedFiles {
			if err := data.DeleteTokensOfFile(dbId, file.Path); err != nil {
				return err
			}
		}

		// Update changed files
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

		// Flush tokens to database
		if err := data.UpsertTokens(dbId, tokens); err != nil {
			return err
		}

		changedFiles = nil
		newFiles = nil
		tokens = nil
		return nil
	}

	// Get files in directory, recursively
	paths, err := git.LsFiles(dirPath)
	if err != nil {
		zap.S().Errorw("Failed to list files in directory",
			"dir_path", dirPath,
			"error", err)
		return ErrFileOperation
	}

	for _, filepath := range paths {
		// Exclude unwanted files
		if isDesired, err := isFileEligible(filepath); err != nil {
			zap.S().Errorw("Failed to check if file is eligible",
				"file_path", filepath,
				"error", err)
			return ErrFileOperation
		} else if !isDesired {
			zap.S().Debugw("Skipping ineligible file",
				"file_path", filepath)
			continue
		}

		// Calculate size and mtime
		size, err := fileutils.Size(filepath)
		if err != nil {
			zap.S().Errorw("Failed to get file size",
				"file_path", filepath,
				"error", err)
			return ErrFileOperation
		}
		mtime, err := fileutils.Mtime(filepath)
		if err != nil {
			zap.S().Errorw("Failed to get file mtime",
				"file_path", filepath,
				"error", err)
			return ErrFileOperation
		}

		file := data.File{Path: filepath, Size: size, Mtime: mtime}

		if dbFile, ok := dbFiles[filepath]; ok {
			// File from db still exists, so remove from deletion list
			delete(removedFiles, filepath)

			if file.VersionEquals(dbFile) {
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

		// Tokenize file and collect unique tokens
		uniqueFileTokens, err := getUniqueTokensOfFile(filepath)
		if err != nil {
			return err
		}
		tokens = append(tokens, uniqueFileTokens...)

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

// getUniqueTokensOfFile tokenizes the specified file and returns the unique
// eligible tokens.
func getUniqueTokensOfFile(path string) ([]data.Token, error) {
	// Tokenize file
	allTokens, err := tokenizer.TokenizeFile(path, isWordEligible)
	if err != nil {
		zap.S().Errorw("Failed to tokenize file",
			"file_path", path,
			"error", err)
		return nil, ErrTextProcessing
	}

	// Retrieve first occurrence of each token
	uniqueTokens := []data.Token{}
	lookup := map[string]bool{}
	for _, fileToken := range allTokens {
		if _, ok := lookup[fileToken]; isTokenEligible(fileToken) && !ok {
			uniqueTokens = append(
				uniqueTokens, data.Token{Path: path, Value: fileToken})
			lookup[fileToken] = true
		}
	}

	return uniqueTokens, nil
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
		zap.S().Errorw("Failed to stat file",
			"file_path", fpath,
			"error", err)
		return false, ErrFileOperation
	}

	isTextFile, err := fileutils.IsTextFile(fpath)
	if err != nil {
		zap.S().Errorw("Failed to determine if file is text",
			"file_path", fpath,
			"error", err)
		return false, ErrFileOperation
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
