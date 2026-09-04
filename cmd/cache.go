package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var cacheCleanCmd = &cobra.Command{
	Use:   "clean [directory]",
	Short: "Remove Dix scan cache",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) == 1 {
			dir = args[0]
		}
		cacheDir := filepath.Join(dir, ".dix", "cache")
		if err := os.RemoveAll(cacheDir); err != nil {
			fatalDixError(fmt.Errorf("failed to clean cache: %w", err))
		}
		fmt.Printf("\033[32m[Cache]\033[0m Removed %s\n", cacheDir)
	},
}

var cacheCmd = &cobra.Command{Use: "cache", Short: "Manage Dix scan cache"}

func init() {
	cacheCmd.AddCommand(cacheCleanCmd)
	rootCmd.AddCommand(cacheCmd)
}
