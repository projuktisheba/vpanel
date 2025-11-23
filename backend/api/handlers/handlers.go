package handlers

import (
	"log"

	"github.com/projuktisheba/vpanel/backend/internal/dbrepo"
	"github.com/projuktisheba/vpanel/backend/internal/models"
)

type HandlerRepo struct {
	Auth          AuthHandler
	MySQLManager     MySQLManagerHandler
	PostgreSQLManager     PostgreSQLManagerHandler
	WordPress     WordPressHandler
	PHP           PHPHandler
	DomainHandler DomainHandler
	SSLHandler    SSLHandler
}

func NewHandlerRepo(host string, db *dbrepo.DBRepository, JWT models.JWTConfig, infoLog, errorLog *log.Logger, mysqlRootDSN string, postgresqlRootDSN string) *HandlerRepo {
	return &HandlerRepo{
		Auth:          newAuthHandler(db, JWT, infoLog, errorLog),
		MySQLManager:     newMySQLManagerHandler(db, infoLog, errorLog, mysqlRootDSN),
		PostgreSQLManager:     newPostgreSQLManagerHandler(db, infoLog, errorLog, postgresqlRootDSN),
		WordPress:     newWordPressHandler(db, infoLog, errorLog),
		PHP:           newPHPHandler(db, infoLog, errorLog),
		DomainHandler: newDomainHandler(host, db, infoLog, errorLog),
		SSLHandler:    newSSLHandler(infoLog, errorLog),
	}
}
