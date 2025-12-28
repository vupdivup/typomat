// Package domain implements the core business logic of the application.
package domain

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/vupdivup/typomat/internal/data"
	"github.com/vupdivup/typomat/pkg/files"
	"github.com/vupdivup/typomat/pkg/git"
	"github.com/vupdivup/typomat/pkg/random"
	"github.com/vupdivup/typomat/pkg/random/lazy"
	"github.com/vupdivup/typomat/pkg/tokenizer"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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
	maxFileSize = 24_000_000 // 24 MB

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

// FileStatus represents the status of a file with respect to the database.
type FileStatus int

const (
	// FileStatusUnchanged indicates the file is unchanged since last
	// tokenization.
	FileStatusUnchanged FileStatus = iota
	// FileStatusChanged indicates the file has changed since last tokenization.
	FileStatusChanged
	// FileStatusNew indicates the file is new and not present in the database.
	FileStatusNew
	// FileStatusIneligible indicates the file is ineligible for tokenization.
	FileStatusIneligible
)

// fileProcessingResult encapsulates the result of processing a file.
type fileProcessingResult struct {
	// file is the processed file metadata.
	file data.File
	// status is the status of the file with respect to the database.
	status FileStatus
	// tokens are the unique tokens extracted from the file. Empty if file
	// is unchanged or ineligible.
	tokens []data.Token
	// err is any error encountered during processing.
	err error
}

// Prompt generates a prompt of the maximum specified character length
// from tokens of the specified directory.
// This is the main entry point of the domain package.
func Prompt(dirPath string, maxLen int) (string, error) {
	zap.S().Infow("Generating prompt from directory text content",
		"dir_path", dirPath,
		"max_len", maxLen)

	// Check if directory exists
	dirExists, err := files.DirExists(dirPath)
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
		if err := ProcessDirectory(absPath); err != nil {
			return "", err
		}
		hasTokenized = true
	}

	return generatePrompt(maxLen)
}

// ProcessDirectory tokenizes all eligible files in the specified directory
// and stores the tokens in the database.
func ProcessDirectory(dirPath string) error {
	var tokens []data.Token
	var changedFiles []data.File
	var newFiles []data.File

	dbFiles := make(map[string]data.File)
	removedFiles := make(map[string]data.File)

	// Populate DB file lookup and removed files lookup
	dbFilesTmp, err := data.GetFiles(dbId)
	if err != nil {
		return err
	}
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
	if len(paths) == 0 {
		zap.S().Errorw("No files found in directory",
			"dir_path", dirPath)
		return ErrEmptyDir
	}

	// Prepare for concurrent processing
	maxWorkers := runtime.NumCPU()
	// Use less workers if there are fewer files than CPUs
	// Also ensure at least one batch
	numBatches := max(min(len(paths)/maxWorkers, len(paths)), 1)
	batches := slices.Chunk(paths, numBatches)
	zap.S().Infow("Starting file processing",
		"dir_path", dirPath,
		"file_count", len(paths),
		"workers", maxWorkers)

	// Process files concurrently
	group, ctx := errgroup.WithContext(context.Background())
	results := make(chan fileProcessingResult)
	for batch := range batches {
		group.Go(func() error {
			return processFileBatch(batch, dbFiles, results, ctx)
		})
	}

	// Close results channel when all processing is done
	groupErr := make(chan error, 1)
	go func() {
		err := group.Wait()
		close(results)
		groupErr <- err
	}()

	// Receive file processed signals and flush tokens as needed
	for result := range results {
		if result.err != nil {
			return result.err
		}

		// File still exists, remove from removed files lookup
		// Ineligible files are kept for deletion in case they were previously
		// tokenized
		if result.status != FileStatusIneligible {
			delete(removedFiles, result.file.Path)
		}

		switch result.status {
		case FileStatusIneligible:
			zap.S().Debugw("Skipping ineligible file",
				"file_path", result.file.Path)
			continue
		case FileStatusUnchanged:
			zap.S().Debugw("Skipping unchanged file",
				"file_path", result.file.Path)
			continue
		case FileStatusChanged:
			zap.S().Debugw("Processing changed file",
				"file_path", result.file.Path,
				"token_count", len(result.tokens))
			changedFiles = append(changedFiles, result.file)
		case FileStatusNew:
			zap.S().Debugw("Processing new file",
				"file_path", result.file.Path,
				"token_count", len(result.tokens))
			newFiles = append(newFiles, result.file)
		}

		// Add received tokens to token buffer and flush if needed
		tokens = append(tokens, result.tokens...)
		if len(tokens) >= tokenBufferSize {
			if err := flushTokens(); err != nil {
				return err
			}
		}
	}

	// Check for errors from goroutines
	if err := <-groupErr; err != nil {
		return err
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

// processFileBatch processes a batch of files and sends the results to the
// provided channel.
func processFileBatch(
	paths []string, dbFiles map[string]data.File,
	results chan fileProcessingResult, ctx context.Context,
) error {
	for _, path := range paths {
		// Process file
		// TODO: is this necessary? Can we just skip a file?
		result := processFile(path, dbFiles)
		if result.err != nil {
			return result.err
		}

		// Check for context cancellation
		select {
		case results <- result:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// processFile processes a single file and returns the processing results.
func processFile(
	path string, dbFiles map[string]data.File,
) fileProcessingResult {
	// Exclude unwanted files
	if isDesired, err := isFileEligible(path); err != nil {
		zap.S().Errorw("Failed to check if file is eligible",
			"file_path", path,
			"error", err)
		return fileProcessingResult{err: ErrFileOperation}
	} else if !isDesired {
		return fileProcessingResult{
			file: data.File{Path: path}, status: FileStatusIneligible,
		}
	}

	// Calculate size and mtime
	stat, err := os.Stat(path)
	if err != nil {
		zap.S().Errorw("Failed to stat file",
			"file_path", path,
			"error", err)
		return fileProcessingResult{err: ErrFileOperation}
	}
	size := int(stat.Size())
	mtime := stat.ModTime()

	file := data.File{Path: path, Size: size, Mtime: mtime}

	// Determine file status
	var fileStatus FileStatus
	if dbFile, ok := dbFiles[path]; ok {
		if file.VersionEquals(dbFile) {
			fileStatus = FileStatusUnchanged
		} else {
			fileStatus = FileStatusChanged
		}
	} else {
		fileStatus = FileStatusNew
	}

	if fileStatus == FileStatusUnchanged {
		// File unchanged, skip tokenization
		return fileProcessingResult{file: file, status: fileStatus}
	}

	// Tokenize file and collect unique tokens
	uniqueFileTokens, err := getUniqueTokensOfFile(path)
	if err != nil {
		return fileProcessingResult{err: err}
	}

	return fileProcessingResult{
		file: file, tokens: uniqueFileTokens, status: fileStatus,
	}
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

	isTextFile, err := files.IsTextFile(fpath)
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
