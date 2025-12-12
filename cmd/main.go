package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vupdivup/recital/internal/config"
	"github.com/vupdivup/recital/internal/ui"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:  fmt.Sprintf("%s [DIRECTORY]", config.AppName),
	Args: cobra.MaximumNArgs(1),
	RunE: run,
}

func run(cmd *cobra.Command, args []string) error {
	dirPath := "."
	if len(args) > 0 {
		dirPath = args[0]
	}

	return ui.Launch(dirPath)
}

func main() {
	defer zap.S().Sync()
	if err := rootCmd.Execute(); err != nil {
		// Exit silently on error
		// Error message is diplayed through the runE handler anyway
		os.Exit(0)
	}
}
