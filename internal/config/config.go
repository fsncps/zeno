package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DBHost string
	DBPort string
	DBName string
	DBUser string
	DBPass string
	Theme  string

	OpenAIKey string

	ConfigFile string
	EnvFile    string
}

type yamlConfig struct {
	Database struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
		Name string `yaml:"name"`
	} `yaml:"database"`
	UI struct {
		Theme string `yaml:"theme"`
	} `yaml:"ui"`
}

func Load() (Config, error) {
	cfg := defaults()
	cfg.ConfigFile = configPath()
	cfg.EnvFile = envPath()

	if data, err := os.ReadFile(cfg.ConfigFile); err == nil {
		var yc yamlConfig
		if err := yaml.Unmarshal(data, &yc); err != nil {
			fmt.Fprintf(os.Stderr, "WARN: malformed config.yaml, using defaults: %v\n", err)
		} else {
			if yc.Database.Host != "" {
				cfg.DBHost = yc.Database.Host
			}
			if yc.Database.Port != "" {
				cfg.DBPort = yc.Database.Port
			}
			if yc.Database.Name != "" {
				cfg.DBName = yc.Database.Name
			}
			if yc.UI.Theme != "" {
				cfg.Theme = yc.UI.Theme
			}
		}
	}

	if data, err := os.ReadFile(cfg.EnvFile); err == nil {
		parseEnvFile(data, &cfg)
	}

	if v := os.Getenv("ZENODB_USER"); v != "" {
		cfg.DBUser = v
	}
	if v := os.Getenv("ZENODB_PASS"); v != "" {
		cfg.DBPass = v
	}
	if v := os.Getenv("ZENODB_HOST"); v != "" {
		cfg.DBHost = v
	}
	if v := os.Getenv("ZENODB_PORT"); v != "" {
		cfg.DBPort = v
	}
	if v := os.Getenv("ZENODB_NAME"); v != "" {
		cfg.DBName = v
	}
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		cfg.OpenAIKey = v
	}

	if cfg.DBUser == "" || cfg.DBPass == "" {
		return cfg, fmt.Errorf("database credentials not found.\nSet ZENODB_USER and ZENODB_PASS in ~/.config/zeno/.env\nor environment variables")
	}

	return cfg, nil
}

func defaults() Config {
	return Config{
		DBHost: "127.0.0.1",
		DBPort: "3306",
		DBName: "zeno",
		DBUser: "",
		DBPass: "",
		Theme:  "catppuccin-macchiato",
	}
}

func configPath() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "zeno", "config.yaml")
	}
	return ""
}

func envPath() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "zeno", ".env")
	}
	return ""
}

func parseEnvFile(data []byte, cfg *Config) {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "ZENODB_USER":
			cfg.DBUser = value
		case "ZENODB_PASS":
			cfg.DBPass = value
		case "ZENODB_HOST":
			cfg.DBHost = value
		case "ZENODB_PORT":
			cfg.DBPort = value
		case "ZENODB_NAME":
			cfg.DBName = value
		case "OPENAI_API_KEY":
			cfg.OpenAIKey = value
		}
	}
}
