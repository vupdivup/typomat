package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vupdivup/typomat/internal/config"
	"github.com/vupdivup/typomat/internal/ui"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("%s [DIRECTORY]", config.AppName),
	Short: "Turn your code into muscle memory",
	Long: `
typomat is a command-line typing practice tool that creates exercises
from the contents of your repository.

It runs through a directory's source code, extracting words from variable
declarations, string literals and function signatures. These words are then used
to build short, randomized typing prompts relevant to your codebase.

Run typomat without any arguments to practice on the current directory.
To use a different source, provide a local path for the program.`,
	Args: cobra.MaximumNArgs(1),
	RunE: run,
}

func run(cmd *cobra.Command, args []string) error {
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

	dirPath := "."
	if len(args) > 0 {
		dirPath = args[0]
	}

	return ui.Launch(dirPath)
}

func init() {
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	rootCmd.Flags().BoolP(
		"purge", "p", false, "purge application cache")
}

func main() {
	// Defer zap logger sync
	defer zap.S().Sync() //nolint:errcheck

	// Configure application
	if err := config.Init(); err != nil {
		zap.S().Error("Failed to initialize configuration", "error", err)
		exitWithErr(err)
	}

	// Clean old files
	// NOTE: cleanup errors don't block application startup
	if err := config.RemoveOldFiles(); err != nil {
		zap.S().Error("Failed to clean old files", "error", err)
	}

	// Run main command
	if err := rootCmd.Execute(); err != nil {
		exitWithErr(err)
	}
}

// exitWithErr prints an error message to stderr along with the usage
// instructions.
// It then exits the application with a non-zero status code.
func exitWithErr(err error) {
	zap.S().Error("Application error", "error", err)
	fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
	fmt.Fprint(os.Stderr, rootCmd.UsageString()+"\n")
	zap.S().Sync() //nolint:errcheck
	os.Exit(1)
}
