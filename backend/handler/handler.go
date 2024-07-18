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

type Deps struct {
	TimeSource domain.TimeSource
	Mailer     *mail.Mailer
}

func NewHandler(db *sql.DB, config domain.Config, deps Deps) *Handler {
	return &Handler{
		db:         db,
		config:     config,
		timeSource: deps.TimeSource,
		mailer:     deps.Mailer,
	}
}
