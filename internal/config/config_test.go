package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mmk31585/url-shortener/internal/config"
)

func TestLoad_AllDefaults(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	defer os.Clearenv()

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
	if cfg.DatabaseURL != "postgres://localhost:5432/test?sslmode=disable" {
		t.Errorf("expected DATABASE_URL from env, got %q", cfg.DatabaseURL)
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
}

func TestLoad_ProductionSetsJsonLogFormat(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_ENV", "production")
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	defer os.Clearenv()

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
	os.Setenv("DATABASE_URL", "postgres://user:pass@db:5432/mydb?sslmode=disable")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("SHORTCODE_LENGTH", "12")
	os.Setenv("MAX_URL", "4096")
	defer os.Clearenv()

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
	if cfg.DatabaseURL != "postgres://user:pass@db:5432/mydb?sslmode=disable" {
		t.Errorf("expected custom DatabaseURL, got %q", cfg.DatabaseURL)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel=debug, got %q", cfg.LogLevel)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("expected LogFormat=json, got %q", cfg.LogFormat)
	}
	if cfg.ShortcodeLength != 12 {
		t.Errorf("expected ShortcodeLength=12, got %d", cfg.ShortcodeLength)
	}
	if cfg.MaxURL != 4096 {
		t.Errorf("expected MaxURL=4096, got %d", cfg.MaxURL)
	}
}

func TestLoad_MissingDatabaseURLReturnsError(t *testing.T) {
	os.Clearenv()

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing, got nil")
	}
	if err.Error() != "config: DATABASE_URL is required" {
		t.Errorf("expected specific error message, got: %v", err)
	}
}

func TestLoad_EmptyDatabaseURLReturnsError(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_URL", "")
	defer os.Unsetenv("DATABASE_URL")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is empty, got nil")
	}
}

func TestLoad_InvalidShortcodeLengthTooSmall(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("SHORTCODE_LENGTH", "0")
	defer os.Clearenv()

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when SHORTCODE_LENGTH=0, got nil")
	}
}

func TestLoad_InvalidShortcodeLengthTooLarge(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("SHORTCODE_LENGTH", "33")
	defer os.Clearenv()

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when SHORTCODE_LENGTH=33, got nil")
	}
}

func TestLoad_InvalidMaxURLZero(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("MAX_URL", "0")
	defer os.Clearenv()

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when MAX_URL=0, got nil")
	}
}

func TestLoad_InvalidMaxURLNegative(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("MAX_URL", "-1")
	defer os.Clearenv()

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when MAX_URL=-1, got nil")
	}
}

func TestLoad_InvalidShortcodeLengthNonNumericFallsBackToDefault(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("SHORTCODE_LENGTH", "not-a-number")
	defer os.Clearenv()

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
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("MAX_URL", "not-a-number")
	defer os.Clearenv()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected non-numeric MAX_URL to fall back to default, got: %v", err)
	}
	if cfg.MaxURL != 2048 {
		t.Errorf("expected MaxURL=2048 (default) for non-numeric input, got %d", cfg.MaxURL)
	}
}

func TestLoad_LogFormatOverridesAutoDefault(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("LOG_FORMAT", "text")
	os.Setenv("APP_ENV", "production")
	defer os.Clearenv()

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
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("APP_ENV", "production")
	defer os.Clearenv()

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
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("APP_ENV", "development")
	defer os.Clearenv()

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
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/test?sslmode=disable")
	os.Setenv("APP_ENV", "production")
	os.Setenv("LOG_FORMAT", "text")
	defer os.Clearenv()

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
DATABASE_URL=postgres://user:pass@localhost:5432/mydb?sslmode=disable
APP_ENV=staging
LOG_LEVEL=debug
SHORTCODE_LENGTH=12
`), 0644)
	if err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	os.Setenv("DOTENV_PATH", envFile)
	defer os.Unsetenv("DOTENV_PATH")
	defer os.Clearenv()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error loading .env via Load, got: %v", err)
	}

	if cfg.DatabaseURL != "postgres://user:pass@localhost:5432/mydb?sslmode=disable" {
		t.Errorf("expected DATABASE_URL from .env, got %q", cfg.DatabaseURL)
	}
	if cfg.AppEnv != "staging" {
		t.Errorf("expected AppEnv=staging from .env, got %q", cfg.AppEnv)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel=debug from .env, got %q", cfg.LogLevel)
	}
	if cfg.ShortcodeLength != 12 {
		t.Errorf("expected ShortcodeLength=12 from .env, got %d", cfg.ShortcodeLength)
	}
}

func TestLoad_DotEnv_IgnoresCommentsAndBlankLines(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envFile, []byte(`
# Comment line

DATABASE_URL=postgres://localhost:5432/test?sslmode=disable

# Another comment
APP_ENV=development
`), 0644)
	if err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	os.Setenv("DOTENV_PATH", envFile)
	defer os.Unsetenv("DOTENV_PATH")
	defer os.Clearenv()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.DatabaseURL != "postgres://localhost:5432/test?sslmode=disable" {
		t.Errorf("expected DATABASE_URL to be set, got %q", cfg.DatabaseURL)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("expected AppEnv=development, got %q", cfg.AppEnv)
	}
}

func TestLoad_DotEnv_QuotedValues(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envFile, []byte(`
DATABASE_URL="postgres://user:pass@host/db"
APP_ENV='development'
`), 0644)
	if err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	os.Setenv("DOTENV_PATH", envFile)
	defer os.Unsetenv("DOTENV_PATH")
	defer os.Clearenv()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.DatabaseURL != "postgres://user:pass@host/db" {
		t.Errorf("expected unquoted DATABASE_URL, got %q", cfg.DatabaseURL)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("expected APP_ENV=development, got %q", cfg.AppEnv)
	}
}

func TestLoad_DotEnv_MissingFileIsIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "nonexistent.env")

	os.Setenv("DOTENV_PATH", envFile)
	defer os.Unsetenv("DOTENV_PATH")
	defer os.Clearenv()

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error because DATABASE_URL is required and missing")
	}
}

func TestLoad_DotEnv_InvalidLinesAreSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envFile, []byte(`
DATABASE_URL=postgres://localhost:5432/test?sslmode=disable
this-is-not-a-valid-env-line
KEY=
`), 0644)
	if err != nil {
		t.Fatalf("failed to create .env file: %v", err)
	}

	os.Setenv("DOTENV_PATH", envFile)
	defer os.Unsetenv("DOTENV_PATH")
	defer os.Clearenv()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.DatabaseURL != "postgres://localhost:5432/test?sslmode=disable" {
		t.Errorf("expected DATABASE_URL to be set, got %q", cfg.DatabaseURL)
	}
}
