package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "habibi-go",
	Short: "Agentic Coding Management Platform",
	Long: `A unified platform for managing AI coding agents across projects and sessions.
	
Complete documentation is available at https://github.com/your-org/habibi-go`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	
	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.habibi-go/config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	
	// Bind flags to viper
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	
	// Add subcommands
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(projectCmd)
	rootCmd.AddCommand(sessionCmd)
	rootCmd.AddCommand(agentCmd)
	rootCmd.AddCommand(configCmd)
}

func initConfig() {
	if cfgFile := viper.GetString("config"); cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		
		// Search config in home directory with name ".habibi-go" (without extension)
		viper.AddConfigPath(home + "/.habibi-go")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}
	
	viper.AutomaticEnv()
	
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}