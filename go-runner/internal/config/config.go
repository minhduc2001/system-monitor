package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	HotReload HotReloadConfig `mapstructure:"hot_reload"`
}

type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	Mode         string `mapstructure:"mode"` // debug, release, test
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	Path     string `mapstructure:"path"` // For SQLite
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type HotReloadConfig struct {
	Enabled     bool     `mapstructure:"enabled"`
	WatchDirs   []string `mapstructure:"watch_dirs"`
	ExcludeDirs []string `mapstructure:"exclude_dirs"`
	IncludeExts []string `mapstructure:"include_exts"`
	ExcludeExts []string `mapstructure:"exclude_exts"`
	Delay       int      `mapstructure:"delay"` // milliseconds
	BuildCmd    string   `mapstructure:"build_cmd"`
	RunCmd      string   `mapstructure:"run_cmd"`
	LogLevel    string   `mapstructure:"log_level"`
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/go-runner")

	// Set default values
	setDefaults()

	// Enable reading from environment variables
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found, using defaults and environment variables")
		} else {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
		os.Exit(1)
	}

	// Ensure data directory exists for SQLite
	if config.Database.Driver == "sqlite" {
		dir := filepath.Dir(config.Database.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating data directory: %v\n", err)
		}
	}

	return &config
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	// Database defaults
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.path", "./data/project.db")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "go_runner")
	viper.SetDefault("database.sslmode", "disable")

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	// Hot reload defaults
	viper.SetDefault("hot_reload.enabled", true)
	viper.SetDefault("hot_reload.watch_dirs", []string{".", "cmd", "internal"})
	viper.SetDefault("hot_reload.exclude_dirs", []string{"tmp", "vendor", "node_modules", ".git"})
	viper.SetDefault("hot_reload.include_exts", []string{".go", ".yaml", ".yml", ".json"})
	viper.SetDefault("hot_reload.exclude_exts", []string{".log", ".tmp"})
	viper.SetDefault("hot_reload.delay", 1000)
	
	// Set platform-specific defaults
	setPlatformSpecificDefaults()
	
	viper.SetDefault("hot_reload.log_level", "info")
}

// setPlatformSpecificDefaults sets platform-specific default values
func setPlatformSpecificDefaults() {
	switch runtime.GOOS {
	case "windows":
		viper.SetDefault("hot_reload.build_cmd", "go build -o ./tmp/main.exe cmd/server/main.go")
		viper.SetDefault("hot_reload.run_cmd", "./tmp/main.exe")
	case "darwin", "linux":
	viper.SetDefault("hot_reload.build_cmd", "go build -o ./tmp/main.exe cmd/server/main.go")
	viper.SetDefault("hot_reload.run_cmd", "./tmp/main.exe")
	default:
		// Fallback for other platforms
	viper.SetDefault("hot_reload.build_cmd", "go build -o ./tmp/main.exe cmd/server/main.go")
	viper.SetDefault("hot_reload.run_cmd", "./tmp/main.exe")
	}
}
