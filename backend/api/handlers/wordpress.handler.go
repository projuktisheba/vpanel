package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/projuktisheba/vpanel/backend/internal/dbrepo"
	"github.com/projuktisheba/vpanel/backend/internal/deploy"
	"github.com/projuktisheba/vpanel/backend/internal/models"
	"github.com/projuktisheba/vpanel/backend/internal/utils"
)

type WordPressHandler struct {
	DB       *dbrepo.DBRepository
	infoLog  *log.Logger
	errorLog *log.Logger
}

func newWordPressHandler(db *dbrepo.DBRepository, infoLog, errorLog *log.Logger) WordPressHandler {
	return WordPressHandler{
		DB:       db,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

func (h *WordPressHandler) DeploySite(w http.ResponseWriter, r *http.Request) {
	var req models.Project
	if err := utils.ReadJSON(w, r, &req); err != nil {
		h.errorLog.Println("ERROR_01_DeploySite: invalid JSON:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request payload: %w", err))
		return
	}

	// ======== Trim & Validate ========
	//domain name
	req.DomainName = strings.TrimSpace(req.DomainName)
	if req.DomainName == "" {
		utils.BadRequest(w, errors.New("domainName is missing"))
		return
	}
	//database name
	req.DBName = strings.TrimSpace(req.DBName)
	if req.DBName == "" {
		utils.BadRequest(w, errors.New("dbName is missing"))
		return
	}
	//project status
	if req.Status == "" {
		req.Status = models.ProjectStatusInit
	}

	projectDir := utils.GetWordpressProjectDirectory()
	projectUniqueName := utils.GetWordpressProjectName(req.DomainName)

	//update object
	req.ProjectName = projectUniqueName
	req.TemplatePath = ""
	req.ProjectFramework = "Wordpress"
	req.ProjectDirectory = projectDir

	// ======== Create Project ========
	// step: Insert a record to the projects table
	if err := h.DB.ProjectRepo.CreateProject(r.Context(), &req); err != nil {
		h.errorLog.Println("ERROR_02_DeploySite: failed to create project:", err)
		utils.ServerError(w, fmt.Errorf("failed to create project: %w", err))
		return
	}

	// step:2 Call PHP builder function
	if err := deploy.DeployWordPress(req.DomainName, req.ProjectDirectory); err != nil {
		h.errorLog.Println("ERROR_03_DeploySite: failed to create project:", err)

		//silently update project status to Error
		req.Status = models.ProjectStatusError
		h.DB.ProjectRepo.UpdateProjectStatus(r.Context(), req.ID, req.Status)

		//send the response
		utils.ServerError(w, fmt.Errorf("failed to create project: %w", err))
		return
	}

	// step:3 Update the project status
	req.Status = models.ProjectStatusRunning
	if _, err := h.DB.ProjectRepo.UpdateProjectStatus(r.Context(), req.ID, req.Status); err != nil {
		h.errorLog.Println("ERROR_04_DeploySite: failed to update project status:", err)
		utils.ServerError(w, fmt.Errorf("failed to update project status: %w", err))
		return
	}

	// ======== Build Response ========
	databaseDetails, _ := h.DB.DBRegistry.GetDatabaseByName(r.Context(), req.DBName)
	req.DatabaseInfo = &databaseDetails
	resp := struct {
		Error   bool           `json:"error"`
		Message string         `json:"message"`
		Summary models.Project `json:"summary"`
	}{
		Error:   false,
		Message: "Project created successfully",
		Summary: req,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *WordPressHandler) UpdateProjectStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("project_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid project ID"))
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := utils.ReadJSON(w, r, &req); err != nil {
		h.errorLog.Println("ERROR_01_UpdateProjectStatus: invalid JSON:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request payload: %w", err))
		return
	}

	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		utils.BadRequest(w, errors.New("status is required"))
		return
	}

	updatedAt, err := h.DB.ProjectRepo.UpdateProjectStatus(r.Context(), id, req.Status)
	if err != nil {
		h.errorLog.Println("ERROR_02_UpdateProjectStatus: failed to update status:", err)
		utils.ServerError(w, fmt.Errorf("failed to update project status: %w", err))
		return
	}

	resp := struct {
		Error     bool   `json:"error"`
		Message   string `json:"message"`
		ID        int64  `json:"id"`
		Status    string `json:"status"`
		UpdatedAt string `json:"updated_at"`
	}{
		Error:     false,
		Message:   "Project status updated successfully",
		ID:        id,
		Status:    req.Status,
		UpdatedAt: updatedAt.Format(time.RFC3339),
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *WordPressHandler) GetSiteStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("project_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid project ID"))
		return
	}
	//Get the project info
	project, err := h.DB.ProjectRepo.GetProjectByID(r.Context(), int64(id))
	if err != nil {
		h.errorLog.Println("ERROR_01_GetSiteStatus: failed to get project information:", err)
		utils.ServerError(w, fmt.Errorf("Site information not found"))
		return
	}

	var resp struct {
		Error      bool   `json:"error"`
		Message    string `json:"message"`
		SiteStatus string `json:"siteStatus"`
	}
	resp.Error = false
	resp.Message = "Site status: " + project.Status
	resp.SiteStatus = project.Status

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *WordPressHandler) SuspendSite(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("project_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid project ID"))
		return
	}
	//Get the project info
	var project *models.Project
	project, err = h.DB.ProjectRepo.GetProjectByID(r.Context(), int64(id))
	if err != nil {
		h.errorLog.Println("ERROR_01_SuspendSite: failed to suspend project:", err)
		utils.ServerError(w, fmt.Errorf("Site information not found"))
		return
	}

	//Only Wordpress site can be suspended
	if project.ProjectFramework != "Wordpress" {
		utils.ServerError(w, fmt.Errorf("Only Wordpress site can be suspended"))
		return
	}
	//Only running site can be suspended
	if project.Status != models.ProjectStatusRunning {
		utils.ServerError(w, fmt.Errorf("Only running site can be suspended"))
		return
	}

	//suspend wordpress site
	if err := deploy.SuspendWordpressSite(project.ProjectName); err != nil {
		h.errorLog.Println("ERROR_02_SuspendSite: failed to suspend project:", err)
		utils.ServerError(w, fmt.Errorf("failed to suspend project: %w", err))
		return
	}

	//update status to suspended
	if _, err := h.DB.ProjectRepo.UpdateProjectStatus(r.Context(), id, models.ProjectStatusSuspended); err != nil {
		h.errorLog.Println("ERROR_01_SuspendSite: failed to suspend project:", err)
		utils.ServerError(w, fmt.Errorf("failed to suspend project: %w", err))
		return
	}

	resp := models.Response{
		Error:   false,
		Message: "Project suspended successfully",
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *WordPressHandler) RestartSite(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("project_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid project ID"))
		return
	}
	//Get the project info
	var project *models.Project
	project, err = h.DB.ProjectRepo.GetProjectByID(r.Context(), int64(id))
	if err != nil {
		h.errorLog.Println("ERROR_01_RestartSite: failed to restart project:", err)
		utils.ServerError(w, fmt.Errorf("failed to restart project: %w", err))
		return
	}
	//Only Wordpress site can be restarted
	if project.ProjectFramework != "Wordpress" {
		utils.ServerError(w, fmt.Errorf("Only Wordpress site can be restarted"))
		return
	}
	//Only suspended site can be restarted
	if project.Status != models.ProjectStatusSuspended {
		utils.ServerError(w, fmt.Errorf("Only suspended project can be restarted"))
		return
	}

	//restart wordpress site
	if err := deploy.RestartWordpressSite(project.ProjectName); err != nil {
		h.errorLog.Println("ERROR_02_RestartSite: failed to restart project:", err)
		utils.ServerError(w, fmt.Errorf("failed to restart project: %w", err))
		return
	}

	//update status to suspended
	if _, err := h.DB.ProjectRepo.UpdateProjectStatus(r.Context(), id, models.ProjectStatusRunning); err != nil {
		h.errorLog.Println("ERROR_01_RestartSite: failed to restart project:", err)
		utils.ServerError(w, fmt.Errorf("failed to restart project: %w", err))
		return
	}

	resp := models.Response{
		Error:   false,
		Message: "Project restarted successfully",
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *WordPressHandler) DeleteSite(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("project_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid project ID"))
		return
	}
	//Get the project info
	var project *models.Project
	project, err = h.DB.ProjectRepo.GetProjectByID(r.Context(), int64(id))
	if err != nil {
		h.errorLog.Println("ERROR_01_DeleteProject: failed to delete project:", err)
		utils.ServerError(w, fmt.Errorf("failed to delete project: %w", err))
		return
	}
	//Only Wordpress site can be deleted
	if project.ProjectFramework != "Wordpress" {
		utils.ServerError(w, fmt.Errorf("Only Wordpress site can be deleted"))
		return
	}
	//delete the project files and users
	if err := deploy.DeleteWordpressSite(project.ProjectName,project.DomainName, project.ProjectDirectory); err != nil {
		h.errorLog.Println("ERROR_02_DeleteProject: failed to delete project:", err)
		utils.ServerError(w, fmt.Errorf("failed to delete project: %w", err))
		return
	}

	if err := h.DB.ProjectRepo.DeleteProject(r.Context(), id); err != nil {
		h.errorLog.Println("ERROR_01_DeleteProject: failed to delete project:", err)
		utils.ServerError(w, fmt.Errorf("failed to delete project: %w", err))
		return
	}

	resp := models.Response{
		Error:   false,
		Message: "Project deleted successfully",
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
func (h *WordPressHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	// Get optional query param
	framework := strings.TrimSpace(r.URL.Query().Get("framework"))

	var projects []*models.Project
	var err error

	if framework != "" {
		projects, err = h.DB.ProjectRepo.ListProjectsByFramework(r.Context(), framework)
	} else {
		projects, err = h.DB.ProjectRepo.ListProjects(r.Context())
	}

	if err != nil {
		h.errorLog.Println("ERROR_01_ListProjects: failed to fetch projects:", err)
		utils.ServerError(w, fmt.Errorf("failed to fetch projects: %w", err))
		return
	}

	resp := struct {
		Error    bool              `json:"error"`
		Message  string            `json:"message"`
		Projects []*models.Project `json:"projects"`
	}{
		Error:    false,
		Message:  "Projects fetched successfully",
		Projects: projects,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
