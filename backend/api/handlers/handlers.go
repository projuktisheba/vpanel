package handlers

import (
	"log"

	"github.com/projuktisheba/vpanel/backend/internal/dbrepo"
	"github.com/projuktisheba/vpanel/backend/internal/models"
)

type HandlerRepo struct {
	Auth AuthHandler
	DBManager DatabaseManagerHandler
}

func NewHandlerRepo(db *dbrepo.DBRepository, JWT models.JWTConfig, infoLog, errorLog *log.Logger) *HandlerRepo {
	return &HandlerRepo{
		Auth: *NewAuthHandler(db, JWT, infoLog, errorLog),
		DBManager: *NewDatabaseManagerHandler(db, infoLog, errorLog),
	}
}
