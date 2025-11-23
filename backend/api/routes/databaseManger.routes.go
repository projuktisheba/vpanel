package routes

import "github.com/go-chi/chi/v5"

func databaseRegistryRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== MySQL Management Routes ========
	// Sensitive / requires DSN → POST
	mux.Get("/mysql/databases", handlerRepo.MySQLManager.ListMySQLDatabases)
	mux.Post("/mysql/create-database", handlerRepo.MySQLManager.CreateMySQLDatabase)
	mux.Post("/mysql/import-database", handlerRepo.MySQLManager.ImportMySQLDatabase)
	mux.Delete("/mysql/delete-database", handlerRepo.MySQLManager.DeleteMySQLDatabase)
	mux.Delete("/mysql/reset-database", handlerRepo.MySQLManager.ResetMySQLDatabase)
	mux.Get("/mysql/users", handlerRepo.MySQLManager.ListMySQLUsers)
	mux.Post("/mysql/create-user", handlerRepo.MySQLManager.CreateMySQLUser)
	// mux.Patch("/mysql/grant", handlerRepo.DB.GrantPrivileges)

	// ======== PostgreSQL Management Routes ========
	// Sensitive / requires DSN → POST
	mux.Get("/postgresql/databases", handlerRepo.PostgreSQLManager.ListPostgreSQLDatabases)
	mux.Post("/postgresql/create-database", handlerRepo.PostgreSQLManager.CreatePostgreSQLDatabase)
	mux.Post("/postgresql/import-database", handlerRepo.PostgreSQLManager.ImportPostgreSQLDatabase)
	mux.Delete("/postgresql/delete-database", handlerRepo.PostgreSQLManager.DeletePostgreSQLDatabase)
	mux.Delete("/postgresql/reset-database", handlerRepo.PostgreSQLManager.ResetPostgreSQLDatabase)
	mux.Get("/postgresql/users", handlerRepo.PostgreSQLManager.ListPostgreSQLUsers)
	mux.Post("/postgresql/create-user", handlerRepo.PostgreSQLManager.CreatePostgreSQLUser)
	// mux.Patch("/postgresql/grant", handlerRepo.DB.GrantPrivileges)

	return mux
}
