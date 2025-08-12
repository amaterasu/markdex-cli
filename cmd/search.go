package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/amaterasu/markdex-cli/internal/api"
	"github.com/amaterasu/markdex-cli/internal/config"
)

var (
	searchJSON bool
)

var searchCmd = &cobra.Command{
	Use:   "search <natural-language-query>",
	Short: "AI-powered natural language bookmark search",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		q := strings.Join(args, " ")
		if strings.TrimSpace(q) == "" {
			return errors.New("empty query")
		}
		cfg, _ := config.Load()
		base := firstNonEmpty(apiBase, cfg.APIBase)
		if base == "" {
			return fmt.Errorf("API base not set (use markdex config set --api <url>)")
		}
		items, err := api.SearchAI(base, q)
		if err != nil {
			return err
		}
		// sort deterministically by title
		sort.Slice(items, func(i, j int) bool { return strings.ToLower(items[i].Title) < strings.ToLower(items[j].Title) })
		if searchJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(items)
		}
		if len(items) == 0 {
			fmt.Println("No matching entries found.")
			return nil
		}
		for i, b := range items {
			fmt.Printf("%3d  %-40s  %s\n", i+1, truncate(b.Title, 40), b.URL)
		}
		return nil
	},
}

func init() {
	searchCmd.Flags().BoolVar(&searchJSON, "json", false, "Output JSON")
	rootCmd.AddCommand(searchCmd)
}
