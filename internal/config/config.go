package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv          string
	ServerAddress   string
	DatabaseURL     string
	LogLevel        string
	LogFormat       string
	ShortcodeLength int
	MaxURL          int
}

func Load() (*Config, error) {
	err := loadDotEnv()
	if err != nil {
		return nil, err
	}
	appEnv := GetString("APP_ENV", "development")
	serverAddress := GetString("SERVER_ADDRESS", ":8080")
	databaseURL := GetString("DATABASE_URL", "")
	logLevel := GetString("LOG_LEVEL", "info")
	logFormat := GetString("LOG_FORMAT", "")
	shortcodeLength := GetInt("SHORTCODE_LENGTH", 8)
	maxURL := GetInt("MAX_URL", 2048)

	if databaseURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL is required")
	}
	if shortcodeLength < 1 || shortcodeLength > 32 {
		return nil, fmt.Errorf("config: SHORTCODE_LENGTH must be between 1 and 32, got %d", shortcodeLength)
	}
	if maxURL < 1 {
		return nil, fmt.Errorf("config: MAX_URL must be positive, got %d", maxURL)
	}
	if logFormat == "" {
		if appEnv == "production" {
			logFormat = "json"
		} else {
			logFormat = "text"
		}
	}

	return &Config{
		AppEnv:          appEnv,
		ServerAddress:   serverAddress,
		DatabaseURL:     databaseURL,
		LogLevel:        logLevel,
		LogFormat:       logFormat,
		ShortcodeLength: shortcodeLength,
		MaxURL:          maxURL,
	}, nil
}

func loadDotEnv() error {
	envPath := GetString("DOTENV_PATH", ".env")
	if envPath == "" {
		return nil
	}

	absPath, err := filepath.Abs(envPath)
	if err != nil {
		return err
	}

	file, err := os.Open(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		} else if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
			value = value[1 : len(value)-1]
		}

		_ = os.Setenv(key, value)
	}

	return scanner.Err()
}

func GetString(key string, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return val
}

func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return valAsInt
}
