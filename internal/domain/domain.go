package domain

import (
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"

	"github.com/vupdivup/recital/internal/data"
	"github.com/vupdivup/recital/pkg/fileutils"
	"github.com/vupdivup/recital/pkg/git"
	"github.com/vupdivup/recital/pkg/tokenizer"
	"go.uber.org/zap"
)

var (
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
	filesToUpsert := []data.File{}

	dbFilesTmp, err := data.GetFiles(dbId)
	if err != nil {
		return err
	}

	dbFiles := map[string]data.File{}
	filesToRemove := map[string]data.File{}
	for _, dbFile := range dbFilesTmp {
		dbFiles[dbFile.Path] = dbFile
		filesToRemove[dbFile.Path] = dbFile
	}

	flushTokens := func() error {
		// Delete tokens and file entries for changed or new files to purge
		// token sets
		for _, file := range filesToUpsert {
			// TODO: don't delete if new
			zap.S().Infow("Updating tokens for file",
				"file_path", file.Path)
			if err := data.DeleteFile(dbId, file, true); err != nil {
				return err
			}
		}

		// (Re-)upload files previously deleted
		if err := data.UpsertFiles(dbId, filesToUpsert); err != nil {
			return err
		}

		if len(tokens) == 0 {
			return nil
		}

		if err := data.UpsertTokens(dbId, tokens); err != nil {
			return err
		}

		filesToUpsert = filesToUpsert[:0]
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
			zap.S().Infow("Skipping ineligible file",
				"file_path", filepath)
			continue
		}

		// Calculate file fingerprint
		fingerprint, err := fileutils.GetFingerprint(filepath)
		if err != nil {
			return err
		}
		if dbFile, ok := dbFiles[filepath]; ok {
			// File still exists, so remove from deletion list
			delete(filesToRemove, filepath)

			if dbFile.Fingerprint == fingerprint {
				// File unchanged, skip tokenization
				zap.S().Infow("Skipping unchanged file",
					"file_path", filepath)
				continue
			}
		}

		// Mark file for upsert
		filesToUpsert = append(
			filesToUpsert, data.File{Path: filepath, Fingerprint: fingerprint})

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
	for _, file := range filesToRemove {
		zap.S().Infow("Deleting removed file from database",
			"file_path", file.Path)
		if err := data.DeleteFile(dbId, file, true); err != nil {
			return err
		}
	}

	return nil
}

// generatePrompt generates a prompt of the maximum specified character length
// from tokens of the database.
func generatePrompt(maxChars int) (string, error) {
	tokens := []string{}
	promptLen := 0
	for {
		// Get a random token
		token, err := getRandomToken()
		if err != nil {
			return "", err
		}

		// Account for space before token
		if promptLen != 0 {
			promptLen += 1
		}

		// Check if adding the token would exceed maxChars
		promptLen += len([]rune(token))
		if promptLen > maxChars {
			break
		}

		tokens = append(tokens, token)
	}

	prompt := strings.Join(tokens, " ")

	return prompt, nil
}

// getRandomToken returns a random token from the database using reservoir
// sampling.
func getRandomToken() (string, error) {
	iter := data.IterUniqueTokens(dbId)
	token := ""
	i := 0

	for item, err := range iter {
		if err != nil {
			return "", err
		} else {
			j := rand.IntN(i + 1)
			if j == i {
				token = item.Value
			}
		}

		i++
	}

	return token, nil
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
