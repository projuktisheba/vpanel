package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/projuktisheba/vpanel/backend/internal/dbrepo"
	"github.com/projuktisheba/vpanel/backend/internal/models"
	"github.com/projuktisheba/vpanel/backend/internal/utils"
)

type PostgreSQLManagerHandler struct {
	postgresqlRootDSN string
	DB                *dbrepo.DBRepository
	infoLog           *log.Logger
	errorLog          *log.Logger
}

func newPostgreSQLManagerHandler(db *dbrepo.DBRepository, infoLog, errorLog *log.Logger, postgresqlRootDSN string) PostgreSQLManagerHandler {
	return PostgreSQLManagerHandler{
		postgresqlRootDSN: postgresqlRootDSN, // Must have admin privileges
		DB:                db,
		infoLog:           infoLog,
		errorLog:          errorLog,
	}
}

func (h *PostgreSQLManagerHandler) CreatePostgreSQLDatabase(w http.ResponseWriter, r *http.Request) {

	type payload struct {
		DatabaseName string `json:"database_name"`
		Username     string `json:"database_user"`
	}

	var req payload
	err := utils.ReadJSON(w, r, &req)
	if err != nil {
		h.errorLog.Println("ERROR_01_CreateDB: invalid JSON:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request payload: %w", err))
		return
	}

	h.infoLog.Println("Payload: ", req)

	var registryUser models.DBUser
	var registryDB models.Database

	// ==================== Step 1: Handle user ====================
	if req.Username != "" {
		// Check if user exists in PostgreSQL
		registryUser, err = h.DB.DBRegistry.GetUserByUsername(r.Context(), req.Username)
		if err != nil {
			utils.BadRequest(w, fmt.Errorf("Database user not found"))
			return
		}
	}

	// ==================== Step 2: Handle database ====================
	registryDB, err = h.DB.DBRegistry.GetDatabaseByName(r.Context(), req.DatabaseName)

	if err != nil {
		// Create the database, assign user if provided
		if err := h.DB.PostgreSQL.CreatePostgreSQLDatabase(h.postgresqlRootDSN, req.DatabaseName, registryUser.Username, registryUser.Password); err != nil {
			utils.ServerError(w, err)
			return
		}
	}

	// Add database to registry
	registryDB.DBName = req.DatabaseName
	registryDB.DBType = "postgresql"
	registryDB.UserID = registryUser.ID
	if err := h.DB.DBRegistry.InsertDatabaseRegistry(r.Context(), &registryDB); err != nil {
		if strings.Contains(err.Error(), "databases_db_name_key") {
			utils.BadRequest(w, fmt.Errorf("Database %s already exist", registryDB.DBName))
			return
		}
		utils.ServerError(w, fmt.Errorf("failed to insert database into registry: %w", err))
		return
	}

	// ==================== Step 3: Build response ====================
	var resp models.Response
	resp.Error = false
	resp.Message = fmt.Sprintf("Database '%s' created successfully", req.DatabaseName)
	if registryUser.Username != "" {
		resp.Message += fmt.Sprintf(" with user '%s'", registryUser.Username)
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// ImportPostgreSQLDatabase handles uploading, saving, and optionally executing a .SQL file
// for an existing PostgreSQL database.
//
// This function expects a multipart/form-data POST request with the following fields:
//   - "dbName": the name of the target database (required, must exist in registry)
//   - "sqlFile": the uploaded .sql file to import
//
// Steps:
//  1. Parse the multipart form and validate input.
//  2. Verify that the database exists in the registry.
//  3. Validate that the uploaded file has a ".sql" extension.
//  4. Save the uploaded file to "$USER/projuktisheba/template/database/<dbName>_postgresql.sql".
//  5. Execute the SQL statements in the file against the database.
//  6. Return a JSON response indicating success or failure.
func (h *PostgreSQLManagerHandler) ImportPostgreSQLDatabase(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (limit to 50MB)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		h.errorLog.Println("ERROR_01_ImportDB: failed to parse form:", err)
		utils.BadRequest(w, fmt.Errorf("invalid form data: %w", err))
		return
	}

	// Read dbName
	dbName := r.FormValue("dbName")
	if dbName == "" {
		utils.BadRequest(w, fmt.Errorf("database name is required"))
		return
	}

	h.infoLog.Println("Importing SQL for database:", dbName)

	// Check if database exists in registry
	registryDB, err := h.DB.DBRegistry.GetDatabaseByName(r.Context(), dbName)
	if err != nil || registryDB.ID == 0 {
		utils.BadRequest(w, fmt.Errorf("database %s not found in registry", dbName))
		return
	}

	// Read uploaded SQL file
	file, header, err := r.FormFile("sqlFile")
	if err != nil {
		utils.BadRequest(w, fmt.Errorf("SQL file is required: %w", err))
		return
	}
	defer file.Close()

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".sql") {
		utils.BadRequest(w, fmt.Errorf("invalid file type: only .sql files are allowed"))
		return
	}

	//get temp directory
	dirPath := utils.GetTempDirectory()
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		h.errorLog.Println("ERROR_02_ImportDB: failed to create directory:", err)
		utils.ServerError(w, fmt.Errorf("failed to create directory: %w", err))
		return
	}

	savePath := filepath.Join(dirPath, fmt.Sprintf("%s.sql", registryDB.DBName))
	outFile, err := os.Create(savePath)
	if err != nil {
		h.errorLog.Println("ERROR_03_ImportDB: failed to create file:", err)
		utils.ServerError(w, fmt.Errorf("failed to save file: %w", err))
		return
	}
	defer outFile.Close()

	// Copy uploaded file content
	if _, err := io.Copy(outFile, file); err != nil {
		h.errorLog.Println("ERROR_04_ImportDB: failed to write file:", err)
		utils.ServerError(w, fmt.Errorf("failed to write file: %w", err))
		return
	}

	defer func() {
		//remove file finally
		os.Remove(savePath)
	}()

	// ==================== Execute SQL statements ====================
	err = h.DB.PostgreSQL.ExecuteSQLFile(h.postgresqlRootDSN, dbName, savePath)
	if err != nil {
		h.errorLog.Println("ERROR_05_ImportDB: Failed to execute SQL file:", err)
		utils.ServerError(w, fmt.Errorf("Failed to execute SQL file: %w", err))
		return
	}

	// ==================== Build response ====================

	var resp models.Response
	resp.Error = false
	resp.Message = "Database updated successfully"

	utils.WriteJSON(w, http.StatusOK, resp)
}

// DeletePostgreSQLDatabase permanently drops a PostgreSQL database
// and updates the database registry (soft delete).
func (h *PostgreSQLManagerHandler) DeletePostgreSQLDatabase(w http.ResponseWriter, r *http.Request) {
	dbName := strings.TrimSpace(r.URL.Query().Get("db_name"))
	if dbName == "" {
		utils.BadRequest(w, fmt.Errorf("invalid request payload: database_name is required"))
		return
	}

	h.infoLog.Println("Delete request for DB:", dbName)

	// ----------------------------------------
	// 1 Check if database exists in registry
	// ----------------------------------------
	registryDB, err := h.DB.DBRegistry.GetDatabaseByName(r.Context(), dbName)
	if err != nil {
		utils.BadRequest(w, fmt.Errorf("database '%s' does not exist", dbName))
		return
	}
	fmt.Println(registryDB)
	// ----------------------------------------
	// 2 Drop the database from PostgreSQL
	// ----------------------------------------
	err = h.DB.PostgreSQL.DropPostgreSQLDatabase(h.postgresqlRootDSN, dbName, registryDB.User.Username)
	if err != nil {
		utils.ServerError(w, fmt.Errorf("failed to drop database: %w", err))
		return
	}

	// ----------------------------------------
	// 3 delete from registry
	// ----------------------------------------
	err = h.DB.DBRegistry.DeleteDatabase(r.Context(), registryDB.ID)
	if err != nil {
		utils.ServerError(w, fmt.Errorf("failed to update registry: %w", err))
		return
	}
	// ----------------------------------------
	// 3 delete backup sql file
	// ----------------------------------------
	// Create path to save SQL file
	homeDir, _ := os.UserHomeDir()
	sqlPath := filepath.Join(homeDir, "projuktisheba", "templates", "databases", "postgresql", fmt.Sprintf("%s.sql", registryDB.DBName))
	// Soft delete (ignore errors)
	_ = exec.Command("sudo", []string{"rm", sqlPath}...)

	// ----------------------------------------
	// 4 Response
	// ----------------------------------------
	var resp models.Response
	resp.Error = false
	resp.Message = fmt.Sprintf("Database '%s' deleted successfully", dbName)

	utils.WriteJSON(w, http.StatusOK, resp)
}

// ResetPostgreSQLDatabase connects to the specified PostgreSQL database and truncates all tables,
// deleting all data and resetting auto-increment keys to 1.
// It reads db_name from the query parameter list
func (h *PostgreSQLManagerHandler) ResetPostgreSQLDatabase(w http.ResponseWriter, r *http.Request) {
	dbName := strings.TrimSpace(r.URL.Query().Get("db_name"))
	if dbName == "" {
		utils.BadRequest(w, fmt.Errorf("invalid request payload: database_name is required"))
		return
	}

	// ----------------------------------------
	// 1 Check if database exists in registry
	// ----------------------------------------
	database, err := h.DB.DBRegistry.GetDatabaseByName(r.Context(), dbName)
	if err != nil {
		utils.BadRequest(w, fmt.Errorf("database '%s' does not exist", dbName))
		return
	}

	if database.DBType != "postgresql" {
		utils.BadRequest(w, fmt.Errorf("database %s must be of type PostgreSQL", database.DBName))
		return
	}

	// ----------------------------------------
	// 2 Drop the database from PostgreSQL
	// ----------------------------------------
	dsn := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", database.User.Username, database.User.Password, "5432", dbName)

	err = h.DB.PostgreSQL.ResetPostgreSQLDatabase(dsn, dbName)
	if err != nil {
		utils.ServerError(w, fmt.Errorf("failed to reset database: %w", err))
		return
	}

	// ----------------------------------------
	// 4 Response
	// ----------------------------------------
	var resp models.Response
	resp.Error = false
	resp.Message = fmt.Sprintf("Database '%s' cleared successfully", dbName)

	utils.WriteJSON(w, http.StatusOK, resp)
}

// ListPostgreSQLDatabases handles HTTP requests to fetch all PostgreSQL databases.
// Response (JSON):
//
//	{
//	  "error": false,
//	  "message": "Databases fetched successfully",
//	  "count": 3,
//	  "databases": [database1, database2 ... objects]
//	}
//
// - Returns structured JSON response with consistent format.
func (h *PostgreSQLManagerHandler) ListPostgreSQLDatabases(w http.ResponseWriter, r *http.Request) {

	// Fetch databases
	databases, err := h.DB.DBRegistry.GetAllDatabase(r.Context(), "postgresql")
	if err != nil {
		h.errorLog.Println("ERROR_01_ListPostgreSQLUsers: failed to fetch database list:", err)
		utils.BadRequest(w, fmt.Errorf("failed to list databases: %w", err))
		return
	}

	// Fetch database stats
	for _, d := range databases {
		dsn := fmt.Sprintf("postgres://%s:%s@127.0.0.1:5432/%s", d.User.Username, d.User.Password, d.DBName)
		d.DatabaseSizeMB, d.TableCount, _ = h.DB.PostgreSQL.GetPostgreSQLDatabaseStats(r.Context(), dsn, d.DBName)
	}

	// Prepare successful response
	resp := struct {
		Error     bool               `json:"error"`     // Indicates if an error occurred
		Message   string             `json:"message"`   // Response message
		Count     int                `json:"count"`     // Number of databases
		Databases []*models.Database `json:"databases"` // List of database names
	}{
		Error:     false,
		Message:   "Databases fetched successfully",
		Count:     len(databases),
		Databases: databases,
	}

	// Send response
	utils.WriteJSON(w, http.StatusOK, resp)
}

func (h *PostgreSQLManagerHandler) CreatePostgreSQLUser(w http.ResponseWriter, r *http.Request) {
	// Parse JSON body
	var payload models.DBUser
	if err := utils.ReadJSON(w, r, &payload); err != nil {
		h.errorLog.Println("ERROR_01_CreatePostgreSQLUser: failed to parse request:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Validate input
	if payload.Username == "" || payload.Password == "" {
		utils.BadRequest(w, fmt.Errorf("dbName, user, and password are required"))
		return
	}
	//specify user type
	payload.UserType = "postgresql"

	// Create the PostgreSQL user
	err := h.DB.PostgreSQL.CreatePostgreSQLUser(r.Context(), h.postgresqlRootDSN, payload.Username, payload.Password, []string{})
	if err != nil {
		h.errorLog.Println("ERROR_02_CreatePostgreSQLUser: failed to create postgresql user: ", err)
		utils.BadRequest(w, fmt.Errorf("failed to create PostgreSQL user: %w", err))
		return
	}

	// Call DBRegistry to create the PostgreSQL user
	err = h.DB.DBRegistry.InsertDBUser(r.Context(), &payload)
	if err != nil {
		if strings.Contains(err.Error(), "db_users_username_key") {
			utils.BadRequest(w, fmt.Errorf("User %s already exist", payload.Username))
			return
		}
		h.errorLog.Println("ERROR_03_CreatePostgreSQLUser: failed to insert into user registry:", err)
		utils.BadRequest(w, fmt.Errorf("failed to insert PostgreSQL user into user registry: %w", err))
		return
	}

	// Prepare successful response
	resp := struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		User    string `json:"user"`
		DBName  string `json:"dbName"`
	}{
		Error:   false,
		Message: "PostgreSQL user created successfully",
		User:    payload.Username,
	}

	// Send response
	utils.WriteJSON(w, http.StatusOK, resp)
}

// ListPostgreSQLUsers handles HTTP requests to fetch all PostgreSQL users.
// It retrieves users from the database registry, optionally decrypts their passwords,
// and returns a structured JSON response.
// Only users that are not soft-deleted are returned.
func (h *PostgreSQLManagerHandler) ListPostgreSQLUsers(w http.ResponseWriter, r *http.Request) {
	// Fetch all users from the DBRegistry repository
	users, err := h.DB.DBRegistry.GetAllUsers(r.Context(), "postgresql")
	if err != nil {
		h.errorLog.Println("ERROR_01_ListPostgreSQLUsers: failed to fetch users:", err)
		utils.BadRequest(w, fmt.Errorf("failed to list users: %w", err))
		return
	}

	// Loop through users to decrypt passwords (for admin display)
	// for _, u := range users {
	// 	pass, err := utils.DecryptAES(u.Password)
	// 	if err != nil {
	// 		h.errorLog.Println("ERROR_02_ListPostgreSQLUsers: failed to decrypt password:", err)
	// 		u.Password = "" // Hide password if decryption fails
	// 	} else {
	// 		u.Password = pass
	// 	}
	// }

	// Prepare response payload
	resp := struct {
		Error   bool             `json:"error"`   // Indicates if there was an error
		Message string           `json:"message"` // Human-readable message
		Count   int              `json:"count"`   // Number of users returned
		Users   []*models.DBUser `json:"users"`   // List of user records
	}{
		Error:   false,
		Message: "Users fetched successfully",
		Count:   len(users),
		Users:   users,
	}

	// Send JSON response to client
	utils.WriteJSON(w, http.StatusOK, resp)
}
