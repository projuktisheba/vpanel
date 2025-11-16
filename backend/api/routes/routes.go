package routes

import (
	"log"
	"net"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/projuktisheba/vpanel/backend/api/handlers"
	"github.com/projuktisheba/vpanel/backend/api/middlewares"
	"github.com/projuktisheba/vpanel/backend/internal/dbrepo"
	"github.com/projuktisheba/vpanel/backend/internal/models"
	"github.com/projuktisheba/vpanel/backend/utils"
)

var handlerRepo *handlers.HandlerRepo


func Routes(env string, db *dbrepo.DBRepository, jwt models.JWTConfig, infoLogger, errorLogger *log.Logger, mysqlRootDSN string) http.Handler {
	mux := chi.NewRouter()

	// --- Global middlewares ---
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://vpanel.pssoft.xyz"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Branch-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	mux.Use(middlewares.Logger) // logger

	// --- Static file serving for images ---
	imageDir := filepath.Join(".", "data", "images")
	fs := http.StripPrefix("/api/v1/images/", http.FileServer(http.Dir(imageDir)))
	mux.Handle("/api/v1/images/*", fs)

	// --- Health check ---
	mux.Get("/api/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		ip := "unknown"
		if conn, err := net.Dial("udp", "1.1.1.1:80"); err == nil {
			defer conn.Close()
			ip = conn.LocalAddr().(*net.UDPAddr).IP.String()
		}
		resp := map[string]any{
			"status":    env,
			"server_ip": ip,
		}
		utils.WriteJSON(w, http.StatusOK, resp)
	})

	//get the handler repo
	handlerRepo = handlers.NewHandlerRepo(db, jwt, infoLogger, errorLogger, mysqlRootDSN)
	
	// Mount Auth routes
	mux.Mount("/api/v1/auth", authRoutes())
	
	// =========== Secure Routes ===========
	// Mount database registry routes
	mux.Mount("/api/v1/db", databaseRegistryRoutes())
	
	// Mount project handlers routes
	mux.Mount("/api/v1/project", projectHandlerRoutes())

	return mux
}
