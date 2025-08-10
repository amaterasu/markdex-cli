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
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(pickCmd)
}

func infof(format string, a ...any) { color.New(color.FgHiBlack).Printf(format+"\n", a...) }
