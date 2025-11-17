package handlers

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/projuktisheba/vpanel/backend/internal/dbrepo"
	"github.com/projuktisheba/vpanel/backend/internal/models"
	"github.com/projuktisheba/vpanel/backend/internal/utils"
)

type ProjectHandler struct {
	DB       *dbrepo.DBRepository
	infoLog  *log.Logger
	errorLog *log.Logger
}

func newProjectHandler(db *dbrepo.DBRepository, infoLog, errorLog *log.Logger) ProjectHandler {
	return ProjectHandler{
		DB:       db,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req models.Project
	if err := utils.ReadJSON(w, r, &req); err != nil {
		h.errorLog.Println("ERROR_01_CreateProject: invalid JSON:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request payload: %w", err))
		return
	}

	// ======== Trim & Validate ========
	req.ProjectName = strings.TrimSpace(req.ProjectName)
	req.ProjectFramework = strings.TrimSpace(req.ProjectFramework)

	if req.ProjectName == "" || req.ProjectFramework == "" || req.DomainName == "" {
		utils.BadRequest(w, errors.New("projectName, projectFramework and domainName are required"))
		return
	}

	if req.Status == "" {
		req.Status = "inactive"
	}

	//get file extension
	originalFilename := strings.TrimSpace(r.URL.Query().Get("filename"))
	extension := filepath.Ext(originalFilename)
	if extension == "" {
		extension = ".zip" // assume zipped project by default
	}

	templatePath := utils.GetTemplatePath(req.ProjectFramework, req.ProjectName, extension)
	projectDir := utils.GetProjectDirectory(req.DomainName)

	//update object
	req.TemplatePath = templatePath
	req.ProjectDirectory = projectDir

	// ======== Create Project ========
	//step:1 extract the zip file
	if err := extractZip(templatePath, projectDir); err != nil {
		h.errorLog.Println("ERROR_02_CreateProject: failed to extract zip files:", err)
		utils.ServerError(w, fmt.Errorf("failed to create project: %w", err))
		return
	}

	// step:2 Insert a record to the projects table
	if err := h.DB.ProjectRepo.CreateProject(r.Context(), &req); err != nil {
		h.errorLog.Println("ERROR_03_CreateProject: failed to create project:", err)
		utils.ServerError(w, fmt.Errorf("failed to create project: %w", err))
		return
	}

	// step:3 Call PHP builder function
	if err := DeployPHPProject(req.ProjectDirectory, req.ProjectFramework, req.DomainName); err != nil {
		h.errorLog.Println("ERROR_03_CreateProject: failed to create project:", err)
		utils.ServerError(w, fmt.Errorf("failed to create project: %w", err))
		return
	}

	// ======== Build Response ========
	resp := struct {
		Error   bool            `json:"error"`
		Message string          `json:"message"`
		Project *models.Project `json:"project"`
	}{
		Error:   false,
		Message: "Project created successfully",
		Project: &req,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	var req models.Project

	if err := utils.ReadJSON(w, r, &req); err != nil {
		h.errorLog.Println("ERROR_01_UpdateProject: invalid JSON:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request payload: %w", err))
		return
	}

	// Trim spaces
	req.ProjectName = strings.TrimSpace(req.ProjectName)
	req.ProjectFramework = strings.TrimSpace(req.ProjectFramework)

	if req.ProjectName == "" {
		utils.BadRequest(w, errors.New("projectName is required"))
		return
	}

	if req.ProjectFramework == "" {
		utils.BadRequest(w, errors.New("projectFramework is required"))
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}

	if err := h.DB.ProjectRepo.UpdateProject(r.Context(), &req); err != nil {
		h.errorLog.Println("ERROR_02_UpdateProject: failed to update project:", err)
		utils.ServerError(w, fmt.Errorf("failed to update project: %w", err))
		return
	}

	resp := struct {
		Error   bool            `json:"error"`
		Message string          `json:"message"`
		Project *models.Project `json:"project"`
	}{
		Error:   false,
		Message: "Project updated successfully",
		Project: &req,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *ProjectHandler) UpdateProjectStatus(w http.ResponseWriter, r *http.Request) {
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
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("project_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(w, errors.New("invalid project ID"))
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
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
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

// UploadPHPProjectFile handles uploading a zipped folder in chunks,
// saving it under projuktisheba/projects/php/<project_name>,
// and extracting it.
//
// This function expects a multipart/form-data POST request with the following fields:
//   - "projectName": the name of the target project (required)
//   - "filename": the uploaded ZIP file name
//   - "chunk": one chunk of the ZIP file
//   - "chunkIndex": index of the current chunk (0-based)
//   - "totalChunks": total number of chunks
//
// Steps:
//  1. Parse the multipart form and validate input.
//  2. Save each received chunk to a temporary folder.
//  3. When the last chunk is received, merge chunks into a single ZIP file.
//  4. Extract the ZIP file to projuktisheba/projects/php/<project_name>.
//  5. Remove temporary chunk files and optionally the ZIP file.
//  6. Return a JSON response indicating success or failure.
func (h *ProjectHandler) UploadProjectFile(w http.ResponseWriter, r *http.Request) {
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

	// Read projectFramework
	projectFramework := r.FormValue("projectFramework")
	if projectFramework == "" {
		utils.BadRequest(w, fmt.Errorf("projectFramework is required"))
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

	baseName := strings.TrimSuffix(originalFilename, extension)

	h.infoLog.Printf("Uploading project folder for %s (%s), file: %s, ext: %s\n",
		projectName, projectFramework, baseName, extension)

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
	projectDir := utils.GetTemplateDirectory(projectFramework, projectName)

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
		if err := mergeChunks(tmpDir, finalZipPath, totalChunks); err != nil {
			h.errorLog.Println("ERROR_06_UploadProjectFolder: failed to merge chunks:", err)
			utils.ServerError(w, fmt.Errorf("failed to merge chunks: %w", err))
			return
		}

		// Remove temporary chunks
		os.RemoveAll(tmpDir)
	}

	// ==================== Build Response ====================
	var resp models.Response
	resp.Error = false

	if chunkIndex+1 < totalChunks {
		resp.Message = fmt.Sprintf("Chunk %d of %d uploaded successfully", chunkIndex+1, totalChunks)
	} else {
		resp.Message = "All chunks uploaded successfully. Project folder created."
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// mergeChunks merges all chunk files into a single destination file
func mergeChunks(tmpDir, destPath string, totalChunks int) error {
	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(tmpDir, fmt.Sprintf("chunk_%d", i))
		chunk, err := os.Open(chunkPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(dest, chunk); err != nil {
			chunk.Close()
			return err
		}
		chunk.Close()
	}
	return nil
}

// extractZip extracts a ZIP file to the specified destination directory
// If files already exist, they will be overwritten
func extractZip(zipPath, destDir string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0777); err != nil { // full read/write/exec permissions
		switch {
		case os.IsPermission(err):
			return fmt.Errorf("permission denied while creating directory: %s", destDir)
		case errors.Is(err, syscall.ENOSPC):
			return fmt.Errorf("insufficient storage space to create directory")
		case errors.Is(err, syscall.ENAMETOOLONG):
			return fmt.Errorf("path too long: %s", destDir)
		case errors.Is(err, syscall.EINVAL):
			return fmt.Errorf("invalid directory name: %s", destDir)
		default:
			return fmt.Errorf("unexpected error creating directory %s: %v", destDir, err)
		}
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, 0777); err != nil {
				return err
			}
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(fpath), 0777); err != nil {
			return err
		}

		// Create/truncate file (overwrites if exists)
		outFile, err := os.Create(fpath)
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			outFile.Close()
			rc.Close()
			return err
		}

		outFile.Close()
		rc.Close()
	}

	return nil
}
