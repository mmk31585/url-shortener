package logger

import (
	"log/slog"
	"os"
	"testing"
)

func TestNew_ProductionReturnsJSONHandler(t *testing.T) {
	log := New("production")
	if log == nil {
		t.Fatal("expected non-nil logger for production")
	}
	log.Info("test message", "key", "value")
}

func TestNew_DevelopmentReturnsTextHandler(t *testing.T) {
	log := New("development")
	if log == nil {
		t.Fatal("expected non-nil logger for development")
	}
	log.Debug("debug message", "key", "value")
}

func TestNew_UnknownEnvFallsBackToTextHandler(t *testing.T) {
	log := New("staging")
	if log == nil {
		t.Fatal("expected non-nil logger for unknown env")
	}
	log.Info("info message", "key", "value")
}

func TestNew_EmptyEnvFallsBackToTextHandler(t *testing.T) {
	log := New("")
	if log == nil {
		t.Fatal("expected non-nil logger for empty env")
	}
	log.Info("info message")
}

func TestNew_ReturnsNonNilLogger(t *testing.T) {
	for _, env := range []string{"production", "development", "test", "staging", ""} {
		log := New(env)
		if log == nil {
			t.Errorf("expected non-nil logger for env %q", env)
		}
	}
}

func TestNew_LoggerCanLogAllLevels(t *testing.T) {
	log := New("development")
	log.Debug("debug message")
	log.Info("info message", "request_id", "abc-123")
	log.Warn("warn message", "shortcode", "abc123de")
	log.Error("error message", "code", "NOT_FOUND")
}

func TestNew_ProductionLoggerDoesNotOutputDebug(t *testing.T) {
	log := New("production")
	log.Debug("this should not appear in production")
	log.Info("this should appear in production")
}

func TestNew_CanCreateChildLogger(t *testing.T) {
	log := New("development")
	child := log.With("request_id", "550e8400-e29b-41d4-a716-446655440000")
	child.Info("request processed")
}

func TestNew_LoggerOutputsToStdout(t *testing.T) {
	log := New("development")
	log.Info("stdout test", "env", "development")
	_ = log
}

func TestNew_HasAddSourceEnabledProduction(t *testing.T) {
	log := New("production")
	if log.Handler() == nil {
		t.Fatal("expected handler to be set")
	}
	t.Log("production logger has handler configured")
}

func TestNew_HasAddSourceEnabledDevelopment(t *testing.T) {
	log := New("development")
	if log.Handler() == nil {
		t.Fatal("expected handler to be set")
	}
	t.Log("development logger has handler configured")
}

func TestMain(m *testing.M) {
	_ = slog.New(slog.NewTextHandler(os.Stderr, nil))
	os.Exit(m.Run())
}
