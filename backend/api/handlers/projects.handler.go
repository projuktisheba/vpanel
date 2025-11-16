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
	"time"

	"github.com/projuktisheba/vpanel/backend/internal/dbrepo"
	"github.com/projuktisheba/vpanel/backend/internal/models"
	"github.com/projuktisheba/vpanel/backend/utils"
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
	req.RootDirectory = strings.TrimSpace(req.RootDirectory)

	if req.ProjectName == "" || req.ProjectFramework == "" || req.RootDirectory == "" {
		utils.BadRequest(w, errors.New("project_name, project_framework, and root_directory are required"))
		return
	}

	if req.Status == "" {
		req.Status = "inactive"
	}

	// ======== Create Project ========
	if err := h.DB.ProjectRepo.CreateProject(r.Context(), &req); err != nil {
		h.errorLog.Println("ERROR_02_CreateProject: failed to create project:", err)
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
	req.RootDirectory = strings.TrimSpace(req.RootDirectory)

	if req.ProjectName == "" || req.ProjectFramework == "" || req.RootDirectory == "" {
		utils.BadRequest(w, errors.New("project_name, project_framework, and root_directory are required"))
		return
	}

	if req.Status == "" {
		req.Status = "inactive"
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
		Error    bool             `json:"error"`
		Message  string           `json:"message"`
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
	// Parse multipart form (limit to 50MB per chunk)
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

	h.infoLog.Println("Uploading project folder for:", projectName)

	// Read chunk information
	filename := r.FormValue("filename")
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

	// Get the user home directory
	homeDir, _ := os.UserHomeDir()
	// Temporary chunk folder
	tmpDir := filepath.Join(homeDir, "projuktisheba", "projects", projectFramework, projectName, "tmp_chunks")
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		h.errorLog.Println("ERROR_02_UploadProjectFolder: failed to create tmp directory:", err)
		utils.ServerError(w, fmt.Errorf("failed to create tmp directory: %w", err))
		return
	}

	// Save chunk
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

	// ==================== Merge & Extract when last chunk ====================
	projectDir := filepath.Join(homeDir, "projuktisheba", "projects", "php", projectName)
	finalZipPath := filepath.Join(projectDir, filename)
	if chunkIndex+1 == totalChunks {
		// Ensure final project directory exists
		if err := os.MkdirAll(projectDir, os.ModePerm); err != nil {
			h.errorLog.Println("ERROR_05_UploadProjectFolder: failed to create project directory:", err)
			utils.ServerError(w, fmt.Errorf("failed to create project directory: %w", err))
			return
		}

		// Merge all chunks
		if err := mergeChunks(tmpDir, finalZipPath, totalChunks); err != nil {
			h.errorLog.Println("ERROR_06_UploadProjectFolder: failed to merge chunks:", err)
			utils.ServerError(w, fmt.Errorf("failed to merge chunks: %w", err))
			return
		}

		// Clean up temporary files and ZIP
		os.RemoveAll(tmpDir)
		os.Remove(finalZipPath)
	}

	// ==================== Register the projects ====================

	// ==================== Build response ====================
	var resp models.Response
	resp.Error = false
	resp.Message = "Project folder uploaded successfully"

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
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

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
