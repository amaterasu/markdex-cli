package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/amaterasu/markdex-cli/internal/api"
	"github.com/amaterasu/markdex-cli/internal/config"
	"github.com/amaterasu/markdex-cli/internal/util"
)

var openHashCmd = &cobra.Command{
	Use:   "open <hash-prefix>",
	Short: "Open a bookmark by its hash prefix (first 3+ chars)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prefix := strings.ToLower(args[0])
		if len(prefix) < 3 {
			return errors.New("hash prefix must be at least 3 characters")
		}
		cfg, _ := config.Load()
		base := firstNonEmpty(apiBase, cfg.APIBase)
		if base == "" {
			return fmt.Errorf("API base not set")
		}
		items, err := api.FetchBookmarks(base, url.Values{})
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return errors.New("no bookmarks")
		}
		// Stable order by title for deterministic ambiguity listing
		sort.Slice(items, func(i, j int) bool { return strings.ToLower(items[i].Title) < strings.ToLower(items[j].Title) })
		var matches []api.Bookmark
		for _, b := range items {
			if b.Hash == "" {
				continue
			}
			if strings.HasPrefix(strings.ToLower(b.Hash), prefix) {
				matches = append(matches, b)
			}
		}
		if len(matches) == 0 {
			return fmt.Errorf("no bookmark with hash prefix %s", prefix)
		}
		if len(matches) > 1 {
			// list ambiguous options (limit to 10 for brevity)
			var lines []string
			for i, m := range matches {
				if i >= 10 { // cap
					lines = append(lines, fmt.Sprintf("... and %d more", len(matches)-10))
					break
				}
				shortHash := m.Hash
				if len(shortHash) > 7 {
					shortHash = shortHash[:7]
				}
				lines = append(lines, fmt.Sprintf("%s  %s", shortHash, m.Title))
			}
			return fmt.Errorf("ambiguous hash prefix %s, matches:\n%s", prefix, strings.Join(lines, "\n"))
		}
		return util.OpenBrowser(matches[0].URL)
	},
}

func init() {
	rootCmd.AddCommand(openHashCmd)
}
