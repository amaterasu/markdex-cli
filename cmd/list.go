package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/amaterasu/markdex-cli/internal/api"
	"github.com/amaterasu/markdex-cli/internal/cache"
	"github.com/amaterasu/markdex-cli/internal/config"
)

var (
	flagTag     string
	flagSearch  string
	flagJSON    bool
	flagNoCache bool
)

var listCmd = &cobra.Command{
	Use:   "list [search]",
	Short: "List bookmarks (optionally filter by search query)",
	Long:  "List bookmarks. Optionally provide a search query as a positional argument, e.g. 'markdex list rust'. You can also use the -s flag for search, or -t for tag filtering.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// If a positional argument is provided, treat it as a search query
		if len(args) > 0 && flagSearch == "" {
			flagSearch = strings.Join(args, " ")
		}
		cfg, _ := config.Load()
		base := firstNonEmpty(apiBase, cfg.APIBase)
		if base == "" {
			return fmt.Errorf("API base not set (use markdex config set --api <url>)")
		}

		c := cache.New()
		if !flagNoCache {
			if items, ok := c.Read(); ok && (flagSearch == "" && flagTag == "") {
				return output(items, flagJSON)
			}
		}

		query := url.Values{}
		if flagSearch != "" {
			query.Set("q", flagSearch)
		}
		if flagTag != "" {
			query.Set("tags", flagTag)
		}
		items, err := api.FetchBookmarks(base, query)
		if err != nil {
			return err
		}
		// simple sort by title for deterministic output
		sort.Slice(items, func(i, j int) bool { return strings.ToLower(items[i].Title) < strings.ToLower(items[j].Title) })
		if flagSearch == "" && flagTag == "" {
			c.Write(items)
		}
		return output(items, flagJSON)
	},
}

func init() {
	listCmd.Flags().StringVarP(&flagTag, "tag", "t", "", "Filter by tag")
	listCmd.Flags().StringVarP(&flagSearch, "search", "s", "", "Search query")
	listCmd.Flags().BoolVar(&flagJSON, "json", false, "Output JSON")
	listCmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "Bypass local cache")
}

func output(items []api.Bookmark, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(items)
	}
	if len(items) == 0 {
		fmt.Println("No matching entries found.")
		return nil
	}
	for _, b := range items {
		fmt.Printf("%s  %-47s %s\n", b.Hash[:7], truncate(b.Title, 40), b.Tags)
	}

	return nil
}

func truncate(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	r := []rune(s)
	return string(r[:n-1]) + "…"
}

func firstNonEmpty(xs ...string) string {
	for _, x := range xs {
		if x != "" {
			return x
		}
	}
	return ""
}
