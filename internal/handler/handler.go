package handler

import (
	"context"
	"log/slog"

	"github.com/mmk31585/url-shortener/internal/service"
)

type Pinger interface {
	PingContext(ctx context.Context) error
}

type BaseHandler struct {
	service service.URLService
	logger  *slog.Logger
	db      Pinger
}

func NewBaseHandler(srv service.URLService, logger *slog.Logger, db Pinger) *BaseHandler {
	return &BaseHandler{
		service: srv,
		logger:  logger,
		db:      db,
	}
}
