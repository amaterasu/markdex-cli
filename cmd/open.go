package cmd

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/amaterasu/markdex/cli/internal/api"
	"github.com/amaterasu/markdex/cli/internal/config"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <index>",
	Short: "Open a bookmark by index (from list)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		base := firstNonEmpty(apiBase, cfg.APIBase)
		if base == "" {
			return fmt.Errorf("API base not set")
		}
		items, err := api.FetchBookmarks(base, url.Values{})
		if err != nil {
			return err
		}
		idx, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		if idx < 1 || idx > len(items) {
			return fmt.Errorf("index out of range")
		}
		return openBrowser(items[idx-1].URL)
	},
}

func openBrowser(u string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}
	args = append(args, u)
	return exec.Command(cmd, args...).Start()
}
