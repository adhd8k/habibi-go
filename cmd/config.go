package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"habibi-go/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  `View and manage configuration settings.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run:   runConfigShow,
}

func init() {
	configCmd.AddCommand(configShowCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	fmt.Println("Current Configuration:")
	fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Database: %s\n", cfg.Database.Path)
	fmt.Printf("Projects Directory: %s\n", cfg.Projects.DefaultDirectory)
	fmt.Printf("Log Level: %s\n", cfg.Logging.Level)
	fmt.Printf("Slack Enabled: %t\n", cfg.Slack.Enabled)
}