package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/amaterasu/markdex-cli/internal/api"
	"github.com/amaterasu/markdex-cli/internal/cache"
	"github.com/amaterasu/markdex-cli/internal/config"
)

var (
	addFlagAI         bool
	addFlagTitle      string
	addFlagTags       []string
	addFlagDesc       string
	addFlagSourceFile string
	addFlagJSON       bool
)

var addCmd = &cobra.Command{
	Use:   "add <url>",
	Short: "Add a new bookmark (AI or manual)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := strings.TrimSpace(args[0])
		if url == "" {
			return errors.New("url required")
		}
		cfg, _ := config.Load()
		base := firstNonEmpty(apiBase, cfg.APIBase)
		if base == "" {
			return fmt.Errorf("API base not set (use markdex config set --api <url>)")
		}

		// If AI mode, ignore manual fields (except source-file) unless provided.
		req := api.CreateBookmarkRequest{
			URL:         url,
			AI:          addFlagAI,
			Title:       addFlagTitle,
			Tags:        addFlagTags,
			Description: addFlagDesc,
			SourceFile:  addFlagSourceFile,
		}
		if addFlagAI {
			// In AI mode, we allow explicit Title/Tags/Description overrides if user passed them; backend may fill missing.
		}
		bk, err := api.CreateBookmark(base, req)
		if err != nil {
			return err
		}
		if addFlagJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(bk)
		}
		// Invalidate local cache so next list/pick reflects new bookmark
		c := cache.New()
		_ = os.Remove(c.Path)
		fmt.Printf("Added %s (%s) tags=%v\n", bk.Title, bk.Hash, bk.Tags)
		return nil
	},
}

func init() {
	addCmd.Flags().BoolVar(&addFlagAI, "ai", false, "Use AI to enrich bookmark details")
	addCmd.Flags().StringVarP(&addFlagTitle, "title", "T", "", "Title (manual mode or override AI)")
	addCmd.Flags().StringSliceVarP(&addFlagTags, "tag", "t", nil, "Tag(s) (repeat or comma separated)")
	addCmd.Flags().StringVarP(&addFlagDesc, "description", "d", "", "Description (manual mode or override AI)")
	addCmd.Flags().StringVarP(&addFlagSourceFile, "source-file", "f", "", "Source file (e.g., inbox.md)")
	addCmd.Flags().BoolVar(&addFlagJSON, "json", false, "Output created bookmark as JSON")
}
