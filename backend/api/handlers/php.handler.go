package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/projuktisheba/vpanel/backend/internal/dbrepo"
	"github.com/projuktisheba/vpanel/backend/internal/deploy"
	"github.com/projuktisheba/vpanel/backend/internal/models"
	"github.com/projuktisheba/vpanel/backend/internal/pkg/user"
	"github.com/projuktisheba/vpanel/backend/internal/utils"
)

type PHPHandler struct {
	DB       *dbrepo.DBRepository
	infoLog  *log.Logger
	errorLog *log.Logger
}

func newPHPHandler(db *dbrepo.DBRepository, infoLog, errorLog *log.Logger) PHPHandler {
	return PHPHandler{
		DB:       db,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

func (h *PHPHandler) InitProject(w http.ResponseWriter, r *http.Request) {
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

	// Generate new project object
	var projectData models.Project

	projectData.ProjectName = utils.GetPHPProjectName(req.DomainName)
	projectData.DomainName = req.DomainName
	projectData.DBName = req.DBName
	projectData.ProjectFramework = "PHP"
	projectData.TemplatePath = ""
	projectData.ProjectDirectory = utils.GetPHPProjectDirectory(req.DomainName)
	projectData.Status = models.StatusInit

	// ======== Create Project ========
	// step 1: Insert a record to the projects table
	if err := h.DB.ProjectRepo.CreateProject(r.Context(), &projectData); err != nil {
		h.errorLog.Println("ERROR_02_DeploySite: failed to create project:", err)

		var pgErr *pgconn.PgError

		// Foreign key violation for domain_name
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			utils.BadRequest(w, fmt.Errorf("domain does not exist"))
			return
		}
		// Unique key violation for domain_name
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			utils.BadRequest(w, fmt.Errorf("domain already exist"))
			return
		}

		utils.BadRequest(w, fmt.Errorf("failed to create project: %w", err))
		return
	}
	// ======== Build Response ========

	resp := struct {
		Error   bool           `json:"error"`
		Message string         `json:"message"`
		Summary models.Project `json:"summary"`
	}{
		Error:   false,
		Message: "Project created successfully",
		Summary: projectData,
	}

	utils.WriteJSON(w, http.StatusOK, resp)

}
func (h *PHPHandler) UploadProjectFile(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (limit 10MB per chunk)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.errorLog.Println("ERROR_01_UploadProjectFolder: failed to parse form:", err)
		utils.BadRequest(w, fmt.Errorf("invalid form data: %w", err))
		return
	}

	// Read projectName
	projectName := r.FormValue("projectName")
	if projectName == "" {
		utils.BadRequest(w, fmt.Errorf("projectName is required"))
		return
	}
	// Read projectName
	projectIDStr := r.FormValue("projectID")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil || projectID == 0 {
		utils.BadRequest(w, fmt.Errorf("Invalid project id"))
		return
	}

	// Read filename and extract extension
	originalFilename := r.FormValue("filename")
	if originalFilename == "" {
		utils.BadRequest(w, fmt.Errorf("filename is required"))
		return
	}

	extension := filepath.Ext(originalFilename)
	if extension == "" {
		extension = ".zip" // assume zipped project by default
	}

	// Read chunk information
	chunkIndexStr := r.FormValue("chunkIndex")
	totalChunksStr := r.FormValue("totalChunks")

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		utils.BadRequest(w, fmt.Errorf("invalid chunkIndex: %w", err))
		return
	}
	totalChunks, err := strconv.Atoi(totalChunksStr)
	if err != nil {
		utils.BadRequest(w, fmt.Errorf("invalid totalChunks: %w", err))
		return
	}

	// Read uploaded chunk
	file, _, err := r.FormFile("chunk")
	if err != nil {
		utils.BadRequest(w, fmt.Errorf("chunk file is required: %w", err))
		return
	}
	defer file.Close()

	// Directory to save the final project after rebuild
	projectDir := utils.GetPHPProjectDirectory(projectName)

	// Temporary directory for chunks
	tmpDir := filepath.Join(projectDir, "tmp_chunks")
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		h.errorLog.Println("ERROR_02_UploadProjectFolder: failed to create tmp directory:", err)
		utils.ServerError(w, fmt.Errorf("failed to create tmp directory: %w", err))
		return
	}

	// Save the chunk file
	chunkPath := filepath.Join(tmpDir, fmt.Sprintf("chunk_%d", chunkIndex))
	outFile, err := os.Create(chunkPath)
	if err != nil {
		h.errorLog.Println("ERROR_03_UploadProjectFolder: failed to create chunk file:", err)
		utils.ServerError(w, fmt.Errorf("failed to save chunk: %w", err))
		return
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		h.errorLog.Println("ERROR_04_UploadProjectFolder: failed to write chunk file:", err)
		utils.ServerError(w, fmt.Errorf("failed to write chunk: %w", err))
		return
	}

	// ==================== Merge on last chunk ====================
	// Final zip path = projectDir/projectName + extension
	finalZipPath := filepath.Join(projectDir, projectName+extension)

	if chunkIndex+1 == totalChunks {
		// Ensure project directory exists
		if err := os.MkdirAll(projectDir, os.ModePerm); err != nil {
			h.errorLog.Println("ERROR_05_UploadProjectFolder: failed to create project directory:", err)
			utils.ServerError(w, fmt.Errorf("failed to create project directory: %w", err))
			return
		}

		// Merge chunks into final ZIP
		if err := utils.MergeChunks(tmpDir, finalZipPath, totalChunks); err != nil {
			h.errorLog.Println("ERROR_06_UploadProjectFolder: failed to merge chunks:", err)
			utils.ServerError(w, fmt.Errorf("failed to merge chunks: %w", err))
			return
		}
		// extract zip file to the project directory
		if err := utils.ExtractZip(finalZipPath, projectDir); err != nil {
			h.errorLog.Println("ERROR_07_UploadProjectFolder: failed to extract final zip file:", err)
			utils.ServerError(w, fmt.Errorf("failed to extract final zip file: %w", err))
			return
		}
		// Remove temporary chunks
		os.RemoveAll(tmpDir)
		os.RemoveAll(finalZipPath)

		// set project status to file uploaded
		if _, err := h.DB.ProjectRepo.UpdateProjectStatus(r.Context(), int64(projectID), models.StatusFileUploaded); err != nil {
			h.errorLog.Println("ERROR_08_UploadProjectFolder: failed to update status:", err)
			utils.ServerError(w, fmt.Errorf("failed to update status: %w", err))
			return
		}
	}

	// ==================== Build Response ====================
	var resp models.Response
	resp.Error = false

	if chunkIndex+1 == totalChunks {
		resp.Message = "All chunks uploaded successfully. Project folder created."
	} else {
		resp.Message = fmt.Sprintf("Chunk %d of %d uploaded successfully", chunkIndex+1, totalChunks)
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
func (h *PHPHandler) DeploySite(w http.ResponseWriter, r *http.Request) {
	// Read projectID
	projectIDStr := r.FormValue("projectID")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil || projectID == 0 {
		utils.BadRequest(w, fmt.Errorf("Invalid project id"))
		return
	}

	domainName := strings.TrimSpace(r.FormValue("projectName"))
	if domainName == "" {
		utils.BadRequest(w, fmt.Errorf("domainName is required"))
		return
	}

	// Project Directory
	projectDir := utils.GetPHPProjectDirectory(domainName)

	// ======== Create Project ========
	// step 1: run the php deployer script with arguments [domainName]
	if err := deploy.DeployPHPSite(domainName, user.GetCurrentUser().Name, projectDir); err != nil {
		h.errorLog.Println("ERROR_01_DeploySite: failed to deploy site:", err)
		utils.ServerError(w, fmt.Errorf("failed to to deploy site: %w", err))

		//update project status to error
		_, _ = h.DB.ProjectRepo.UpdateProjectStatus(r.Context(), int64(projectID), models.StatusError)
		return
	}

	// step 3: Update the project status to running
	if _, err := h.DB.ProjectRepo.UpdateProjectStatus(r.Context(), int64(projectID), models.StatusRunning); err != nil {
		h.errorLog.Println("ERROR_02_DeploySite: failed to update project status:", err)
		utils.ServerError(w, fmt.Errorf("failed to update project status: %w", err))
		return
	}

	// ======== Build Response ========
	projectInfo, _ := h.DB.ProjectRepo.GetProjectByID(r.Context(), int64(projectID))
	resp := struct {
		Error   bool            `json:"error"`
		Message string          `json:"message"`
		Summary *models.Project `json:"summary"`
	}{
		Error:   false,
		Message: "Project deployed successfully",
		Summary: projectInfo,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}
func (h *PHPHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
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
