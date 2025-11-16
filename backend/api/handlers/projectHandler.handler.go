package handlers

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

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

// UploadProjectFolder handles uploading a zipped folder in chunks,
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
func (h *ProjectHandler) UploadProjectFolder(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (limit to 500MB per chunk)
	if err := r.ParseMultipartForm(500 << 20); err != nil {
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
	tmpDir := filepath.Join(homeDir, "projuktisheba", "projects", "php", projectName, "tmp_chunks")
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

		if err := mergeChunks(tmpDir, finalZipPath, totalChunks); err != nil {
			h.errorLog.Println("ERROR_06_UploadProjectFolder: failed to merge chunks:", err)
			utils.ServerError(w, fmt.Errorf("failed to merge chunks: %w", err))
			return
		}

		if err := extractZip(finalZipPath, projectDir); err != nil {
			h.errorLog.Println("ERROR_07_UploadProjectFolder: failed to extract ZIP:", err)
			utils.ServerError(w, fmt.Errorf("failed to extract ZIP: %w", err))
			return
		}
		// Testing: Unzip final file
		err = extractZip(finalZipPath, projectDir)
		if err != nil {
			h.errorLog.Println("ERROR_08_UploadProjectFolder: failed to unzip final file:", err)
			utils.ServerError(w, fmt.Errorf("failed to unzip final file: %w", err))
			return
		}
		// Clean up temporary files
		os.RemoveAll(tmpDir)
		os.Remove(finalZipPath)
	}

	// Run project in production mode

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
