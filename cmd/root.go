package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	apiBase string
)

// These are set via -ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "markdex",
	Short: "Markdex CLI - interact with your bookmarks",
	Long:  "Markdex CLI provides fast access to listing, searching, and opening bookmarks.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiBase, "api", "", "API base URL (overrides config)")
	rootCmd.AddCommand(listCmd)
	//rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(pickCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(&cobra.Command{Use: "version", Short: "Show version info", Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("markdex %s (commit %s, built %s)\n", version, commit, date)
	}})
}

func infof(format string, a ...any) { color.New(color.FgHiBlack).Printf(format+"\n", a...) }
