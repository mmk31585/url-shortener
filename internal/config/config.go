package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	cfg     *Config
	loadErr error
	once    sync.Once
)

type Config struct {
	AppEnv          string
	ServerAddress   string
	DB              DBConfig
	LogLevel        string
	LogFormat       string
	ShortcodeLength int
	MaxURL          int
}
type DBConfig struct {
	DBAddr       string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

func Load() (*Config, error) {
	once.Do(func() {
		err := loadDotEnv()
		if err != nil {
			loadErr = err
			return
		}
		appEnv := GetString("APP_ENV", "development")
		serverAddress := GetString("SERVER_ADDRESS", ":8080")
		addr := GetString("DB_ADDR", "")
		maxOpenConns := GetInt("DB_MAX_OPEN_CONNS", 30)
		maxIdleConns := GetInt("DB_MAX_IDLE_CONNS", 30)
		maxIdleTime := GetString("DB_MAX_LIFE_TIME", "5m")

		logLevel := GetString("LOG_LEVEL", "info")
		logFormat := GetString("LOG_FORMAT", "")
		shortcodeLength := GetInt("SHORTCODE_LENGTH", 8)
		maxURL := GetInt("MAX_URL_LENGTH", 2048)

		if addr == "" {
			loadErr = fmt.Errorf("config: DB_ADDR is required")
			return
		}
		if shortcodeLength < 1 || shortcodeLength > 32 {
			loadErr = fmt.Errorf("config: SHORTCODE_LENGTH must be between 1 and 32, got %d", shortcodeLength)
			return
		}
		if maxURL < 1 {
			loadErr = fmt.Errorf("config: MAX_URL must be positive, got %d", maxURL)
			return
		}
		if logFormat == "" {
			if appEnv == "production" {
				logFormat = "json"
			} else {
				logFormat = "text"
			}
		}
		dbConfig := &DBConfig{
			DBAddr:       addr,
			MaxOpenConns: maxOpenConns,
			MaxIdleConns: maxIdleConns,
			MaxIdleTime:  maxIdleTime,
		}
		cfg = &Config{
			AppEnv:          appEnv,
			ServerAddress:   serverAddress,
			DB:              *dbConfig,
			LogLevel:        logLevel,
			LogFormat:       logFormat,
			ShortcodeLength: shortcodeLength,
			MaxURL:          maxURL,
		}
	})

	return cfg, loadErr
}
func GetConfig() *Config {
	return cfg
}

func ResetForTest() {
	once = sync.Once{}
	cfg = nil
	loadErr = nil
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
