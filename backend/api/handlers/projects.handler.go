package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
