package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vupdivup/typelines/internal/config"
	"github.com/vupdivup/typelines/internal/ui"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:  fmt.Sprintf("%s [DIRECTORY]", config.AppName),
	Args: cobra.MaximumNArgs(1),
	RunE: run,
}

func run(cmd *cobra.Command, args []string) error {
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
	rootCmd.Flags().BoolP(
		"purge", "p", false, "purge application cache on startup")
}

func main() {
	// Defer zap logger sync
	defer func() {
		if err := zap.S().Sync(); err != nil {
			println("Error:", err.Error())
		}
	}()

	// Configure application
	if err := config.Init(); err != nil {
		zap.S().Error("Failed to initialize configuration", "error", err)
		println("Error:", err.Error())
	}

	// Run main command
	if err := rootCmd.Execute(); err != nil {
		println("Error:", err.Error())
	}
}
