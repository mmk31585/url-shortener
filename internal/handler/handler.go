package handler

import (
	"database/sql"
	"log/slog"

	"github.com/mmk31585/url-shortener/internal/service"
)

type BaseHandler struct {
	service service.URLService
	logger  *slog.Logger
	db      *sql.DB
}

func NewBaseHandler(srv service.URLService, logger *slog.Logger, db *sql.DB) *BaseHandler {
	return &BaseHandler{
		service: srv,
		logger:  logger,
		db:      db,
	}
}
