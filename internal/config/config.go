package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Projects ProjectsConfig `mapstructure:"projects"`
	Agents   AgentsConfig   `mapstructure:"agents"`
	Slack    SlackConfig    `mapstructure:"slack"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	Auth            AuthConfig    `mapstructure:"auth"`
}

type AuthConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type DatabaseConfig struct {
	Path           string        `mapstructure:"path"`
	BackupEnabled  bool          `mapstructure:"backup_enabled"`
	BackupInterval time.Duration `mapstructure:"backup_interval"`
	MaxConnections int           `mapstructure:"max_connections"`
}

type ProjectsConfig struct {
	DefaultDirectory  string `mapstructure:"default_directory"`
	AutoDiscover      bool   `mapstructure:"auto_discover"`
	WorktreeBasePath  string `mapstructure:"worktree_base_path"`
}

type AgentsConfig struct {
	DefaultTimeout       time.Duration  `mapstructure:"default_timeout"`
	MaxConcurrent        int            `mapstructure:"max_concurrent"`
	HealthCheckInterval  time.Duration  `mapstructure:"health_check_interval"`
	LogRetentionDays     int            `mapstructure:"log_retention_days"`
	ResourceLimits       ResourceLimits `mapstructure:"resource_limits"`
	ClaudeBinaryPath     string         `mapstructure:"claude_binary_path"`
}

type ResourceLimits struct {
	MemoryMB   int `mapstructure:"memory_mb"`
	CPUPercent int `mapstructure:"cpu_percent"`
}

type SlackConfig struct {
	Enabled             bool   `mapstructure:"enabled"`
	BotToken            string `mapstructure:"bot_token"`
	AppToken            string `mapstructure:"app_token"`
	SigningSecret       string `mapstructure:"signing_secret"`
	NotificationChannel string `mapstructure:"notification_channel"`
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	FilePath   string `mapstructure:"file_path"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	
	// Add config paths
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("$HOME/.habibi-go")
	
	// Environment variable support
	viper.SetEnvPrefix("HABIBI")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	
	// Set defaults
	setDefaults()
	
	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	// Expand home directory paths
	if err := expandPaths(&config); err != nil {
		return nil, fmt.Errorf("error expanding paths: %w", err)
	}
	
	return &config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.shutdown_timeout", "10s")
	
	// Auth defaults
	viper.SetDefault("server.auth.enabled", false)
	viper.SetDefault("server.auth.username", "")
	viper.SetDefault("server.auth.password", "")
	
	// Database defaults
	viper.SetDefault("database.path", "~/.habibi-go/data.db")
	viper.SetDefault("database.backup_enabled", true)
	viper.SetDefault("database.backup_interval", "24h")
	viper.SetDefault("database.max_connections", 10)
	
	// Projects defaults
	viper.SetDefault("projects.default_directory", "~/projects")
	viper.SetDefault("projects.auto_discover", true)
	viper.SetDefault("projects.worktree_base_path", ".worktrees")
	
	// Agents defaults
	viper.SetDefault("agents.default_timeout", "30m")
	viper.SetDefault("agents.max_concurrent", 10)
	viper.SetDefault("agents.health_check_interval", "30s")
	viper.SetDefault("agents.log_retention_days", 7)
	viper.SetDefault("agents.resource_limits.memory_mb", 1024)
	viper.SetDefault("agents.resource_limits.cpu_percent", 50)
	viper.SetDefault("agents.claude_binary_path", "claude")
	
	// Slack defaults
	viper.SetDefault("slack.enabled", false)
	viper.SetDefault("slack.notification_channel", "#dev")
	
	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.file_path", "~/.habibi-go/logs/app.log")
	viper.SetDefault("logging.max_size_mb", 100)
	viper.SetDefault("logging.max_backups", 5)
}

func expandPaths(config *Config) error {
	var err error
	
	// Expand database path
	if config.Database.Path, err = expandPath(config.Database.Path); err != nil {
		return err
	}
	
	// Expand projects default directory
	if config.Projects.DefaultDirectory, err = expandPath(config.Projects.DefaultDirectory); err != nil {
		return err
	}
	
	// Expand logging file path
	if config.Logging.FilePath, err = expandPath(config.Logging.FilePath); err != nil {
		return err
	}
	
	return nil
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

func (c *Config) CreateDirectories() error {
	// Create database directory
	if err := os.MkdirAll(filepath.Dir(c.Database.Path), 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}
	
	// Create logging directory
	if err := os.MkdirAll(filepath.Dir(c.Logging.FilePath), 0755); err != nil {
		return fmt.Errorf("failed to create logging directory: %w", err)
	}
	
	// Create projects directory
	if err := os.MkdirAll(c.Projects.DefaultDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create projects directory: %w", err)
	}
	
	return nil
}