package routes

import "github.com/go-chi/chi/v5"

func projectHandlerRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== Project Routes ========
	mux.Post("/", handlerRepo.ProjectHandler.CreateProject)                // Create new project
	mux.Put("/", handlerRepo.ProjectHandler.UpdateProject)            // Update all project fields
	mux.Put("/status", handlerRepo.ProjectHandler.UpdateProjectStatus) // Update only project status, query parameter: project_id
	mux.Delete("/", handlerRepo.ProjectHandler.DeleteProject)         // Delete project, query parameter: project_id
	mux.Get("/list", handlerRepo.ProjectHandler.ListProjects)                  // List all projects

	// Optional: Upload project folder (if using UploadProjectFolder style)
	mux.Post("/upload-project-folder", handlerRepo.ProjectHandler.UploadProjectFile)

	return mux
}
