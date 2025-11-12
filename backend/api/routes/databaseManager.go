package routes

import "github.com/go-chi/chi/v5"

func databaseManagerRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== MySQL Management Routes ========

	// Sensitive / requires DSN → POST
	mux.Get("/mysql/databases", handlerRepo.DBManager.MySQLDB.ListDatabases)
	// mux.Get("/mysql/tables", handlerRepo.DB.ListTables)
	mux.Post("/mysql/create-database", handlerRepo.DBManager.MySQLDB.CreateMySQLDatabase)
	// mux.Post("/mysql/create-user", handlerRepo.DB.CreateUser)
	// mux.Patch("/mysql/grant", handlerRepo.DB.GrantPrivileges)

	return mux
}
