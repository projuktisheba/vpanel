package handlers

import (
	"log"

	"github.com/projuktisheba/vpanel/backend/internal/dbrepo"
	"github.com/projuktisheba/vpanel/backend/internal/models"
)

type HandlerRepo struct {
	Auth           AuthHandler
	DBManager      DatabaseManagerHandler
	Project ProjectHandler
	WordPress WordPressHandler
	PHP PHPHandler
	DomainHandler  DomainHandler
}

func NewHandlerRepo(db *dbrepo.DBRepository, JWT models.JWTConfig, infoLog, errorLog *log.Logger, mysqlRootDSN string) *HandlerRepo {
	return &HandlerRepo{
		Auth:           newAuthHandler(db, JWT, infoLog, errorLog),
		DBManager:      newDatabaseManagerHandler(db, infoLog, errorLog, mysqlRootDSN),
		Project: newProjectHandler(db, infoLog, errorLog),
		WordPress: newWordPressHandler(db, infoLog, errorLog),
		PHP: newPHPHandler(db, infoLog, errorLog),
		DomainHandler:  newDomainHandler(db, infoLog, errorLog),
	}
}
