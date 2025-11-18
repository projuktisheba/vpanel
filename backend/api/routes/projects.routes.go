package routes

import "github.com/go-chi/chi/v5"

func projectHandlerRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== Project Routes ========
	mux.Post("/create", handlerRepo.Project.CreateProject)                // Create new project
	mux.Put("/update", handlerRepo.Project.UpdateProject)            // Update all project fields
	mux.Put("/status", handlerRepo.Project.UpdateProjectStatus) // Update only project status, query parameter: project_id
	mux.Delete("/remove", handlerRepo.Project.DeleteProject)         // Delete project, query parameter: project_id
	mux.Get("/list", handlerRepo.Project.ListProjects)                  // List all projects

	// Optional: Upload project folder (if using UploadProjectFolder style)
	mux.Post("/upload-project-folder", handlerRepo.Project.UploadProjectFile)


	// ======== Wordpress Project Routes ========
	mux.Post("/wordpress/deploy", handlerRepo.WordPress.DeploySite) 
	return mux
}
