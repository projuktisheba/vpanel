package routes

import "github.com/go-chi/chi/v5"

func databaseRegistryRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== MySQL Management Routes ========

	// Sensitive / requires DSN → POST
	mux.Get("/mysql/databases", handlerRepo.DBManager.ListMySQLDatabases)
	mux.Post("/mysql/create-database", handlerRepo.DBManager.CreateMySQLDatabase)
	mux.Post("/mysql/import-database", handlerRepo.DBManager.ImportMySQLDatabase)
	mux.Delete("/mysql/delete-database", handlerRepo.DBManager.DeleteMySQLDatabase)
	mux.Get("/mysql/users", handlerRepo.DBManager.ListMySQLUsers)
	mux.Post("/mysql/create-user", handlerRepo.DBManager.CreateMySQLDatabase)
	// mux.Patch("/mysql/grant", handlerRepo.DB.GrantPrivileges)

	return mux
}
