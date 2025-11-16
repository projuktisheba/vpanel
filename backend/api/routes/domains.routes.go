package routes

import "github.com/go-chi/chi/v5"

func domainHandlerRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== Domain Handler Routes ========

	// Create a new domain
	mux.Post("/", handlerRepo.DomainHandler.CreateDomain)

	// Update entire domain record (domain name + SSL update date)
	mux.Put("/", handlerRepo.DomainHandler.UpdateDomain) //query parameter : domain_id

	// Update only the domain name
	mux.Put("//name", handlerRepo.DomainHandler.UpdateDomainName) //query parameter : domain_id

	// Delete a domain by ID
	mux.Delete("/", handlerRepo.DomainHandler.DeleteDomain) //query parameter : domain_id

	// List all domains
	mux.Get("/list", handlerRepo.DomainHandler.ListDomains)

	return mux
}