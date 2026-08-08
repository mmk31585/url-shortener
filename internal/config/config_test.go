package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mmk31585/url-shortener/internal/config"
)

func TestLoad_AllDefaults(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_ADDR", "postgres://localhost:5432/test?sslmode=disable")
	defer os.Clearenv()

	config.ResetForTest()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error with all defaults, got: %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Errorf("expected AppEnv=development, got %q", cfg.AppEnv)
	}
	if cfg.ServerAddress != ":8080" {
		t.Errorf("expected ServerAddress=:8080, got %q", cfg.ServerAddress)
	}
	if cfg.DB.DBAddr != "postgres://localhost:5432/test?sslmode=disable" {
		t.Errorf("expected DB_ADDR from env, got %q", cfg.DB.DBAddr)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected LogLevel=info, got %q", cfg.LogLevel)
	}
	if cfg.LogFormat != "text" {
		t.Errorf("expected LogFormat=text in non-production, got %q", cfg.LogFormat)
	}
	if cfg.ShortcodeLength != 8 {
		t.Errorf("expected ShortcodeLength=8, got %d", cfg.ShortcodeLength)
	}
	if cfg.MaxURL != 2048 {
		t.Errorf("expected MaxURL=2048, got %d", cfg.MaxURL)
	}
	if cfg.DB.MaxOpenConns != 30 {
		t.Errorf("expected DB.MaxOpenConns=30, got %d", cfg.DB.MaxOpenConns)
	}
	if cfg.DB.MaxIdleConns != 30 {
		t.Errorf("expected DB.MaxIdleConns=30, got %d", cfg.DB.MaxIdleConns)
	}
	if cfg.DB.MaxIdleTime != "5m" {
		t.Errorf("expected DB.MaxIdleTime=5m, got %q", cfg.DB.MaxIdleTime)
	}
}

func TestLoad_ProductionSetsJsonLogFormat(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_ENV", "production")
	os.Setenv("DB_ADDR", "postgres://localhost:5432/test?sslmode=disable")
	defer os.Clearenv()

	config.ResetForTest()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.LogFormat != "json" {
		t.Errorf("expected LogFormat=json in production, got %q", cfg.LogFormat)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_ENV", "staging")
	os.Setenv("SERVER_ADDRESS", ":9090")
	os.Setenv("DB_ADDR", "postgres://user:pass@db:5432/mydb?sslmode=disable")
	os.Setenv("DB_MAX_OPEN_CONNS", "50")
	os.Setenv("DB_MAX_IDLE_CONNS", "10")
	os.Setenv("DB_MAX_LIFE_TIME", "10m")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("SHORTCODE_LENGTH", "8")
	os.Setenv("MAX_URL_LENGTH", "4096")
	defer os.Clearenv()

	config.ResetForTest()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.AppEnv != "staging" {
		t.Errorf("expected AppEnv=staging, got %q", cfg.AppEnv)
	}
	if cfg.ServerAddress != ":9090" {
		t.Errorf("expected ServerAddress=:9090, got %q", cfg.ServerAddress)
	}
	if cfg.DB.DBAddr != "postgres://user:pass@db:5432/mydb?sslmode=disable" {
		t.Errorf("expected custom DatabaseURL, got %q", cfg.DB.DBAddr)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel=debug, got %q", cfg.LogLevel)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("expected LogFormat=json, got %q", cfg.LogFormat)
	}
	if cfg.ShortcodeLength != 8 {
		t.Errorf("expected ShortcodeLength=8, got %d", cfg.ShortcodeLength)
	}
	if cfg.MaxURL != 4096 {
		t.Errorf("expected MaxURL=4096, got %d", cfg.MaxURL)
	}
	if cfg.DB.MaxOpenConns != 50 {
		t.Errorf("expected DB.MaxOpenConns=50, got %d", cfg.DB.MaxOpenConns)
	}
	if cfg.DB.MaxIdleConns != 10 {
		t.Errorf("expected DB.MaxIdleConns=10, got %d", cfg.DB.MaxIdleConns)
	}
	if cfg.DB.MaxIdleTime != "10m" {
		t.Errorf("expected DB.MaxIdleTime=10m, got %q", cfg.DB.MaxIdleTime)
	}
}

func TestLoad_MissingDBAddrReturnsError(t *testing.T) {
	os.Clearenv()

	config.ResetForTest()
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when DB_ADDR is not set, got nil")
	}
	if err.Error() != "config: DB_ADDR is required" {
		t.Errorf("expected specific error message, got: %v", err)
	}
}

func TestLoad_EmptyDBAddrReturnsError(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_ADDR", "")
	defer os.Unsetenv("DB_ADDR")

	config.ResetForTest()
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when DB_ADDR is empty, got nil")
	}
}

func TestLoad_InvalidShortcodeLengthTooSmall(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_ADDR", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("SHORTCODE_LENGTH", "0")
	defer os.Clearenv()

	config.ResetForTest()
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when SHORTCODE_LENGTH=0, got nil")
	}
}

func TestLoad_InvalidShortcodeLengthTooLarge(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_ADDR", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("SHORTCODE_LENGTH", "9")
	defer os.Clearenv()

	config.ResetForTest()
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when SHORTCODE_LENGTH=9, got nil")
	}
}

func TestLoad_InvalidMaxURLZero(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_ADDR", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("MAX_URL_LENGTH", "0")
	defer os.Clearenv()

	config.ResetForTest()
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when MAX_URL_LENGTH=0, got nil")
	}
}

func TestLoad_InvalidMaxURLNegative(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_ADDR", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("MAX_URL_LENGTH", "-1")
	defer os.Clearenv()

	config.ResetForTest()
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when MAX_URL_LENGTH=-1, got nil")
	}
}

func TestLoad_InvalidShortcodeLengthNonNumericFallsBackToDefault(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_ADDR", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("SHORTCODE_LENGTH", "not-a-number")
	defer os.Clearenv()

	config.ResetForTest()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected non-numeric SHORTCODE_LENGTH to fall back to default, got: %v", err)
	}
	if cfg.ShortcodeLength != 8 {
		t.Errorf("expected ShortcodeLength=8 (default) for non-numeric input, got %d", cfg.ShortcodeLength)
	}
}

func TestLoad_InvalidMaxURLNonNumericFallsBackToDefault(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_ADDR", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("MAX_URL_LENGTH", "not-a-number")
	defer os.Clearenv()

	config.ResetForTest()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected non-numeric MAX_URL_LENGTH to fall back to default, got: %v", err)
	}
	if cfg.MaxURL != 2048 {
		t.Errorf("expected MaxURL=2048 (default) for non-numeric input, got %d", cfg.MaxURL)
	}
}

func TestLoad_LogFormatOverridesAutoDefault(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_ADDR", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("LOG_FORMAT", "text")
	os.Setenv("APP_ENV", "production")
	defer os.Clearenv()

	config.ResetForTest()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.LogFormat != "text" {
		t.Errorf("expected explicit LOG_FORMAT=text to override auto-default, got %q", cfg.LogFormat)
	}
}

func TestLoad_LogFormatDefaultsToJsonInProduction(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_ADDR", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("APP_ENV", "production")
	defer os.Clearenv()

	config.ResetForTest()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("expected LogFormat=json in production with no explicit LOG_FORMAT, got %q", cfg.LogFormat)
	}
}

func TestLoad_LogFormatDefaultsToTextInDevelopment(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_ADDR", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("APP_ENV", "development")
	defer os.Clearenv()

	config.ResetForTest()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.LogFormat != "text" {
		t.Errorf("expected LogFormat=text in development with no explicit LOG_FORMAT, got %q", cfg.LogFormat)
	}
}

func TestLoad_ProductionLogFormatNotOverriddenByExplicitText(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_ADDR", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("APP_ENV", "production")
	os.Setenv("LOG_FORMAT", "text")
	defer os.Clearenv()

	config.ResetForTest()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.LogFormat != "text" {
		t.Errorf("expected explicit LOG_FORMAT=text to take precedence in production, got %q", cfg.LogFormat)
	}
}

func TestGetString_UnsetReturnsFallback(t *testing.T) {
	os.Clearenv()

	val := config.GetString("NONEXISTENT_KEY", "fallback_value")
	if val != "fallback_value" {
		t.Errorf("expected fallback_value for unset key, got %q", val)
	}
}

func TestGetString_SetReturnsActualValue(t *testing.T) {
	os.Setenv("TEST_STRING_KEY", "actual_value")
	defer os.Unsetenv("TEST_STRING_KEY")

	val := config.GetString("TEST_STRING_KEY", "fallback")
	if val != "actual_value" {
		t.Errorf("expected actual_value for set key, got %q", val)
	}
}

func TestGetInt_UnsetReturnsFallback(t *testing.T) {
	os.Clearenv()

	val := config.GetInt("NONEXISTENT_INT_KEY", 42)
	if val != 42 {
		t.Errorf("expected 42 for unset int key, got %d", val)
	}
}

func TestGetInt_SetReturnsParsedValue(t *testing.T) {
	os.Setenv("TEST_INT_KEY", "123")
	defer os.Unsetenv("TEST_INT_KEY")

	val := config.GetInt("TEST_INT_KEY", 0)
	if val != 123 {
		t.Errorf("expected 123 for set int key, got %d", val)
	}
}

func TestGetInt_NonNumericReturnsFallback(t *testing.T) {
	os.Setenv("TEST_INT_KEY", "not-a-number")
	defer os.Unsetenv("TEST_INT_KEY")

	val := config.GetInt("TEST_INT_KEY", 7)
	if val != 7 {
		t.Errorf("expected 7 (fallback) for non-numeric int key, got %d", val)
	}
}

func TestLoad_DotEnv_ValidFileLoadsValues(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envFile, []byte(`
DB_ADDR=postgres://user:pass@localhost:5432/mydb?sslmode=disable
APP_ENV=staging
LOG_LEVEL=debug
SHORTCODE_LENGTH=8
`), 0644)
	if err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	os.Setenv("DOTENV_PATH", envFile)
	defer os.Unsetenv("DOTENV_PATH")
	defer os.Clearenv()

	config.ResetForTest()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error loading .env via Load, got: %v", err)
	}

	if cfg.DB.DBAddr != "postgres://user:pass@localhost:5432/mydb?sslmode=disable" {
		t.Errorf("expected DB_ADDR from .env, got %q", cfg.DB.DBAddr)
	}
	if cfg.AppEnv != "staging" {
		t.Errorf("expected AppEnv=staging from .env, got %q", cfg.AppEnv)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel=debug from .env, got %q", cfg.LogLevel)
	}
	if cfg.ShortcodeLength != 8 {
		t.Errorf("expected ShortcodeLength=8 from .env, got %d", cfg.ShortcodeLength)
	}
}

func TestLoad_DotEnv_IgnoresCommentsAndBlankLines(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envFile, []byte(`
# Comment line

DB_ADDR=postgres://localhost:5432/test?sslmode=disable

# Another comment
APP_ENV=development
`), 0644)
	if err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	os.Setenv("DOTENV_PATH", envFile)
	defer os.Unsetenv("DOTENV_PATH")
	defer os.Clearenv()

	config.ResetForTest()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.DB.DBAddr != "postgres://localhost:5432/test?sslmode=disable" {
		t.Errorf("expected DB_ADDR to be set, got %q", cfg.DB.DBAddr)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("expected AppEnv=development, got %q", cfg.AppEnv)
	}
}

func TestLoad_DotEnv_QuotedValues(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envFile, []byte(`
DB_ADDR="postgres://user:pass@host/db"
APP_ENV='development'
`), 0644)
	if err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	os.Setenv("DOTENV_PATH", envFile)
	defer os.Unsetenv("DOTENV_PATH")
	defer os.Clearenv()

	config.ResetForTest()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.DB.DBAddr != "postgres://user:pass@host/db" {
		t.Errorf("expected unquoted DB_ADDR, got %q", cfg.DB.DBAddr)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("expected APP_ENV=development, got %q", cfg.AppEnv)
	}
}

func TestLoad_DotEnv_MissingFileIsIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "nonexistent.env")

	os.Clearenv()
	os.Setenv("DOTENV_PATH", envFile)
	defer os.Unsetenv("DOTENV_PATH")
	defer os.Clearenv()

	config.ResetForTest()
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error because DB_ADDR is required and missing")
	}
}

func TestLoad_DotEnv_InvalidLinesAreSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envFile, []byte(`
DB_ADDR=postgres://localhost:5432/test?sslmode=disable
this-is-not-a-valid-env-line
KEY=
`), 0644)
	if err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	os.Setenv("DOTENV_PATH", envFile)
	defer os.Unsetenv("DOTENV_PATH")
	defer os.Clearenv()

	config.ResetForTest()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.DB.DBAddr != "postgres://localhost:5432/test?sslmode=disable" {
		t.Errorf("expected DB_ADDR to be set, got %q", cfg.DB.DBAddr)
	}
}
