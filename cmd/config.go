package cmd

import (
	"fmt"

	"github.com/amaterasu/markdex-cli/internal/config"
	"github.com/spf13/cobra"
)

var cfgAPI string

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration values",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _ := config.Load()
		if cfgAPI != "" {
			c.APIBase = cfgAPI
		}
		if err := config.Save(c); err != nil {
			return err
		}
		fmt.Println("Saved config")
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _ := config.Load()
		fmt.Printf("apiBase: %s\n", c.APIBase)
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print config file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(config.Path())
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)
	configSetCmd.Flags().StringVar(&cfgAPI, "api", "", "API base URL")
}
