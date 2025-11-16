package dbrepo

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBRepository contains all individual repositories
type DBRepository struct {
	UserRepo    *UserRepo
	DBRegistry *DatabaseRegistryRepo
	MySQL    *MySQLManagerRepo
	Domain *DomainRepo
	ProjectRepo *ProjectRepo
}

// NewDBRepository initializes all repositories with a shared connection pool
func NewDBRepository(db *pgxpool.Pool) *DBRepository {
	return &DBRepository{
		UserRepo:    NewUserRepo(db),
		DBRegistry: newDatabaseRegistryRepo(db),
		MySQL:    NewMySQLManagerRepo(),
		Domain: NewDomainRepo(db),
		ProjectRepo: NewProjectRepo(db),
	}
}
