package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vupdivup/typomat/internal/config"
	"github.com/vupdivup/typomat/internal/ui"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("%s <directory>", config.AppName),
	Short: "Turn your code into muscle memory",
	Long: `typomat is a command-line typing practice tool that creates exercises
from the contents of your repository.

It runs through a directory's source code, extracting words from variable
declarations, string literals and function signatures. These words are then used
to build short, randomized typing prompts relevant to your codebase.

Start a typing session by passing the path to the directory you'd like to
practice on.`,
	Args: cobra.ExactArgs(1),
	RunE: run,
}

func run(cmd *cobra.Command, args []string) error {
	// Configure application
	if err := config.Init(); err != nil {
		zap.S().Error("Failed to initialize configuration", "error", err)
		return err
	}

	// Handle purge flag
	purge, err := cmd.Flags().GetBool("purge")
	if err != nil {
		return err
	}
	if purge {
		if err := config.PurgeCache(); err != nil {
			return err
		}
	}

	// Parse args
	dirPath := args[0]

	// Launch UI
	return ui.Launch(dirPath)
}

func init() {
	rootCmd.Flags().BoolP("purge", "p", false, "purge application cache")
}

func main() {
	defer zap.S().Sync() //nolint:errcheck

	// Run main command
	rootCmd.Execute() // nolint:errcheck
}
