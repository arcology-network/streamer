package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Service string // OrderService
	Module  string // order.processor

	EnableStdout bool

	LogFile    string // ./logs/app.log
	MaxSize    int    // MB
	MaxAge     int    // days
	MaxBackups int

	// ✅✅✅ Minimum log level
	Level   Level
	Version string
}

// LoadConfigFromFile Load log configuration from file
func LoadConfigFromFile(filePath string) (*Config, error) {
	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	// Parse TOML into a struct
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse TOML config: %w", err)
	}

	// Set default values (if not specified in the config file)
	if config.MaxSize == 0 {
		config.MaxSize = 100 // 默认100MB
	}
	if config.LogFile == "" {
		// Use the default log file path
		config.LogFile = "./logs/app.log"
	}

	// Ensure the log directory exists
	if err := ensureLogDir(config.LogFile); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	return &config, nil
}

// ensureLogDir ensures that the directory for the log file exists
func ensureLogDir(logFile string) error {
	dir := filepath.Dir(logFile)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// LoadDefaultConfig Load default configuration
func LoadDefaultConfig() *Config {
	return &Config{
		Service:      "unknown",
		Module:       "unknown",
		EnableStdout: true,
		LogFile:      "./logs/app.log",
		MaxSize:      100, // Default 100MB
		MaxAge:       0,   // Do not delete by time by default
		MaxBackups:   0,   // Retain all backups by default
		Level:        InfoLevel,
		Version:      "1.0.0",
	}
}

// LoadConfigWithFallback Load configuration, return default if failed
func LoadConfigWithFallback(filePath string) *Config {
	config, err := LoadConfigFromFile(filePath)
	if err != nil {
		fmt.Printf("Warning: Failed to load config from %s, using default config: %v\n", filePath, err)
		return LoadDefaultConfig()
	}
	return config
}

// String method, for easy printing of configuration information
func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{Service:%s, Module:%s, EnableStdout:%v, LogFile:%s, MaxSize:%dMB, MaxAge:%ddays, MaxBackups:%d, Level:%v, Version:%s}",
		c.Service, c.Module, c.EnableStdout, c.LogFile, c.MaxSize, c.MaxAge, c.MaxBackups, c.Level, c.Version,
	)
}
