package handler

import (
	"database/sql"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/mail"
)

type Handler struct {
	db         *sql.DB
	timeSource domain.TimeSource
	mailer     *mail.Mailer
	config     domain.Config
}

func NewHandler(
	db *sql.DB,
	timeSource domain.TimeSource,
	mailer *mail.Mailer,
	config domain.Config,
) *Handler {
	return &Handler{
		db:         db,
		timeSource: timeSource,
		mailer:     mailer,
		config:     config,
	}
}
