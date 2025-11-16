package routes

import "github.com/go-chi/chi/v5"

func projectHandlerRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== Project Handler Routes ========

	mux.Post("/upload-project-folder", handlerRepo.ProjectHandler.UploadProjectFolder)
	return mux
}
