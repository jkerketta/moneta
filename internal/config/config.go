package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jkerketta/stocktui/internal/models"
	"github.com/spf13/viper"
)

func Load(customPath string) (*models.AppConfig, error) {
	if customPath != "" {
		viper.SetConfigFile(customPath)
	} else if envPath := os.Getenv("STOCK_TUI_CONFIG"); envPath != "" {
		viper.SetConfigFile(envPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		viper.AddConfigPath(".")

		// XDG / Standard paths
		configDir, err := os.UserConfigDir()
		if err == nil {
			viper.AddConfigPath(filepath.Join(configDir, "stock-tui"))
		}
	}

	// Defaults
	viper.SetDefault("symbols", []string{"BTC-USD", "ETH-USD", "AAPL", "GOOGL"})
	viper.SetDefault("refresh_interval", "5s")
	viper.SetDefault("provider", "simulator")
	viper.SetDefault("default_range", "24H")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
		// Config not found is fine, we use defaults
	}

	var cfg models.AppConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Minimal validation
	if cfg.RefreshInterval < time.Second {
		cfg.RefreshInterval = time.Second
	}

	return &cfg, nil
}
