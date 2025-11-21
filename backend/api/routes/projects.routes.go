package routes

import (
	"github.com/go-chi/chi/v5"
)

func projectHandlerRoutes() *chi.Mux {
	mux := chi.NewRouter()

	// ======== Project Routes ========
	// Optional: Upload project folder in .zip format
	mux.Post("/upload-project-folder", handlerRepo.Project.UploadProjectFile)

	// ======== PHP Project Routes ========
	//Initiate a php project
	// request body: {domainName, dbName}, response: {error, message, summary}
	mux.Post("/php/init", handlerRepo.PHP.InitProject)

	// Upload project folder to the project directory
	// request body: {projectName, projectID, filename, chunkIndex, totalChunks}, response: {error, message}
	mux.Post("/php/upload-project-file", handlerRepo.PHP.UploadProjectFile)

	// Deploy the project(php-fpm setup, dependency installation, nginx server block setup)
	mux.Post("/php/deploy", handlerRepo.PHP.DeploySite)

	// List all projects
	mux.Get("/php/list", handlerRepo.PHP.ListProjects)

	// ======== Wordpress Project Routes ========
	// req body {domainName, dbName}
	mux.Post("/wordpress/deploy", handlerRepo.WordPress.DeploySite)

	// query parameter: project_id
	mux.Post("/wordpress/get-status", handlerRepo.WordPress.GetSiteStatus)
	
	// query parameter: project_id
	mux.Post("/wordpress/suspend", handlerRepo.WordPress.SuspendSite)

	// query parameter: project_id
	mux.Post("/wordpress/restart", handlerRepo.WordPress.RestartSite)

	// query parameter: project_id
	mux.Post("/wordpress/delete", handlerRepo.WordPress.DeleteSite)
	return mux
}
