package logger

import (
	"log/slog"
	"os"
)

func New(env string) *slog.Logger {
	var Handler slog.Handler

	switch env {
	case "production":
		Handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		})
	case "development":
		Handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})
	default:
		Handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})
	}
	return slog.New(Handler)
}
