package routes

import (
	"github.com/go-chi/chi/v5"
)

func sslHandlerRoutes() *chi.Mux {
	mux := chi.NewRouter()
	
	// ---------------------------------------------
	// Route 1: Check if SSL exists for a domain
	// GET /ssl/check?domain=example.com
	// Returns: JSON { error, message, ssl_status }
	// ---------------------------------------------
	mux.Get("/check", handlerRepo.SSLHandler.CheckSSL)

	// ---------------------------------------------
	// Route 2: Check SSL and automatically issue if not present
	// GET /ssl/check-and-issue?domain=example.com
	// Returns: JSON { error, message, ssl_status }
	// ---------------------------------------------
	mux.Get("/check-and-issue", handlerRepo.SSLHandler.CheckAndIssueSSL)

	// ---------------------------------------------
	// Route 3: Force issue SSL for a domain (even if exists)
	// GET /ssl/issue?domain=example.com
	// Returns: JSON { error, message, ssl_status }
	// ---------------------------------------------
	mux.Get("/issue", handlerRepo.SSLHandler.IssueSSL)
	return mux
}
