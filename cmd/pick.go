package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/amaterasu/markdex-cli/internal/api"
	"github.com/amaterasu/markdex-cli/internal/cache"
	"github.com/amaterasu/markdex-cli/internal/config"
	"github.com/amaterasu/markdex-cli/internal/util"
)

var (
	pickFlagTag     string
	pickFlagSearch  string
	pickFlagMulti   bool
	pickFlagCopy    bool
	pickFlagNoCache bool
	pickFlagFzfPath string
)

var pickCmd = &cobra.Command{
	Use:   "pick [initial-query]",
	Short: "Fuzzy-pick bookmarks via fzf",
	Long:  "Loads bookmarks (optionally filtered) and invokes external 'fzf' for fuzzy selection. Enter opens, --multi allows multiple selection. Ctrl-Y copies the hash of the highlighted entry. An optional initial-query argument seeds the fuzzy filter (client-side only).",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		base := firstNonEmpty(apiBase, cfg.APIBase)
		if base == "" {
			return fmt.Errorf("API base not set")
		}
		fzfPath := pickFlagFzfPath
		if fzfPath == "" {
			fzfPath = "fzf"
		}
		if _, err := exec.LookPath(fzfPath); err != nil {
			return errors.New("fzf not found in PATH (install: https://github.com/junegunn/fzf)")
		}

		c := cache.New()
		var items []api.Bookmark
		var ok bool
		useCache := pickFlagSearch == "" && pickFlagTag == "" && !pickFlagNoCache
		if useCache {
			items, ok = c.Read()
		}
		if !ok {
			q := url.Values{}
			if pickFlagSearch != "" {
				q.Set("q", pickFlagSearch)
			}
			if pickFlagTag != "" {
				q.Set("tags", pickFlagTag)
			}
			fetched, err := api.FetchBookmarks(base, q)
			if err != nil {
				return err
			}
			items = fetched
			if useCache {
				c.Write(items)
			}
		}
		if len(items) == 0 {
			return errors.New("no bookmarks")
		}

		// stable sort by title
		sort.Slice(items, func(i, j int) bool { return strings.ToLower(items[i].Title) < strings.ToLower(items[j].Title) })

		lines := make([]string, len(items))
		for i, b := range items {
			shortHash := b.Hash
			if len(shortHash) > 7 {
				shortHash = shortHash[:7]
			}
			tags := strings.Join(b.Tags, ",")
			desc := b.Description
			if len(desc) > 300 { // keep preview concise
				desc = desc[:297] + "..."
			}
			// Columns: index, shortHash, title, tags, description
			lines[i] = fmt.Sprintf("%d\t%s\t%-40s\t%s\t%s", i, shortHash, sanitizeTabs(truncate(b.Title, 40)), tags, sanitizeTabs(desc))
		}

		// Show only short hash and title (cols 2,3)
		fzfArgs := []string{"--with-nth", "2,3", "--delimiter", "\t", "--ansi", "--prompt", "markdex> "}
		if len(args) == 1 && args[0] != "" {
			fzfArgs = append(fzfArgs, "--query", args[0])
		}
		if pickFlagMulti {
			fzfArgs = append(fzfArgs, "--multi")
		}
		// Preview includes tags (column 4) while list shows only hash+title
		preview := "echo TITLE: {3}; echo TAGS: {4}; echo HASH: {2}; echo; echo DESCRIPTION:; echo {5}"
		fzfArgs = append(fzfArgs, "--preview", preview)
		// Key binding: Ctrl-Y copies hash (column 2) to clipboard (macOS pbcopy). Abort to exit without opening.
		fzfArgs = append(fzfArgs, "--bind", "ctrl-y:execute-silent(echo -n {2} | pbcopy)+abort")
		cmdFzf := exec.Command(fzfPath, fzfArgs...)
		stdin, err := cmdFzf.StdinPipe()
		if err != nil {
			return err
		}
		cmdFzf.Stderr = os.Stderr
		out, err := cmdFzf.StdoutPipe()
		if err != nil {
			return err
		}

		if err := cmdFzf.Start(); err != nil {
			return err
		}
		go func() {
			w := bufio.NewWriter(stdin)
			for _, l := range lines {
				fmt.Fprintln(w, l)
			}
			w.Flush()
			stdin.Close()
		}()

		scanner := bufio.NewScanner(out)
		var selected []int
		for scanner.Scan() {
			parts := strings.Split(scanner.Text(), "\t")
			if len(parts) == 0 {
				continue
			}
			idxStr := parts[0]
			var idx int
			fmt.Sscanf(idxStr, "%d", &idx)
			if idx >= 0 && idx < len(items) {
				selected = append(selected, idx)
			}
		}
		err = cmdFzf.Wait()
		// If nothing selected, treat as normal cancel regardless of exit code.
		if len(selected) == 0 {
			return nil
		}
		// If there was an error, ignore common fzf cancel/no-match codes (1, 130), otherwise return it.
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				code := ee.ExitCode()
				if code == 1 || code == 130 { // 1=no match, 130=interrupted/ESC
					return nil
				}
			}
			return err
		}

		if pickFlagCopy {
			// Track usage for the first selected entry, then copy its URL
			if h := strings.TrimSpace(items[selected[0]].Hash); h != "" {
				vals := url.Values{}
				vals.Set("hash", h)
				vals.Set("user_id", cfg.UserID)
				_, _ = api.UseBookmark(base, vals)
			}
			return copyToClipboard(items[selected[0]].URL)
		}

		// track usage and open each
		for _, si := range selected {
			if h := strings.TrimSpace(items[si].Hash); h != "" {
				vals := url.Values{}
				vals.Set("hash", h)
				vals.Set("user_id", cfg.UserID)
				_, _ = api.UseBookmark(base, vals)
			}
			_ = util.OpenBrowser(items[si].URL)
		}
		return nil
	},
}

func init() {
	pickCmd.Flags().StringVarP(&pickFlagTag, "tag", "t", "", "Filter by tag")
	pickCmd.Flags().StringVarP(&pickFlagSearch, "search", "s", "", "Server-side search query before fuzzy picking")
	pickCmd.Flags().BoolVar(&pickFlagMulti, "multi", false, "Allow selecting multiple bookmarks")
	pickCmd.Flags().BoolVar(&pickFlagCopy, "copy", false, "Copy first selected URL to clipboard instead of opening")
	pickCmd.Flags().BoolVar(&pickFlagNoCache, "no-cache", false, "Bypass local cache")
	pickCmd.Flags().StringVar(&pickFlagFzfPath, "fzf", "", "Path to fzf binary (defaults to looking in PATH)")
}

func sanitizeTabs(s string) string { return strings.ReplaceAll(s, "\t", " ") }

// hostFromURL removed (unused)

func copyToClipboard(text string) error {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("pbcopy")
		stdin, _ := cmd.StdinPipe()
		_ = cmd.Start()
		stdin.Write([]byte(text))
		stdin.Close()
		return cmd.Wait()
	case "linux":
		// try xclip then xsel
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd := exec.Command("xclip", "-selection", "clipboard")
			stdin, _ := cmd.StdinPipe()
			_ = cmd.Start()
			stdin.Write([]byte(text))
			stdin.Close()
			return cmd.Wait()
		}
		if _, err := exec.LookPath("xsel"); err == nil {
			cmd := exec.Command("xsel", "--clipboard", "--input")
			stdin, _ := cmd.StdinPipe()
			_ = cmd.Start()
			stdin.Write([]byte(text))
			stdin.Close()
			return cmd.Wait()
		}
	case "windows":
		cmd := exec.Command("powershell", "-Command", "Set-Clipboard -Value @'"+text+"'@")
		return cmd.Run()
	}
	return errors.New("clipboard utility not found")
}
