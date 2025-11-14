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

func NewHandlerRepo(db *dbrepo.DBRepository, JWT models.JWTConfig, infoLog, errorLog *log.Logger, mysqlRootDSN  string) *HandlerRepo {
	return &HandlerRepo{
		Auth: newAuthHandler(db, JWT, infoLog, errorLog),
		DBManager: newDatabaseManagerHandler(db, infoLog, errorLog, mysqlRootDSN),
	}
}
