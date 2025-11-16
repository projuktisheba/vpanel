package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/projuktisheba/vpanel/backend/internal/dbrepo"
	"github.com/projuktisheba/vpanel/backend/internal/models"
	"github.com/projuktisheba/vpanel/backend/utils"
)

type DatabaseManagerHandler struct {
	mysqlRootDSN string
	DB           *dbrepo.DBRepository
	infoLog      *log.Logger
	errorLog     *log.Logger
}

func newDatabaseManagerHandler(db *dbrepo.DBRepository, infoLog, errorLog *log.Logger, mysqlRootDSN string) DatabaseManagerHandler {
	return DatabaseManagerHandler{
		mysqlRootDSN: mysqlRootDSN, // Must have admin privileges
		DB:           db,
		infoLog:      infoLog,
		errorLog:     errorLog,
	}
}

func (h *DatabaseManagerHandler) CreateMySQLDatabase(w http.ResponseWriter, r *http.Request) {

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
		// Check if user exists in MySQL
		registryUser, err = h.DB.DBRegistry.GetMySqlUserByUsername(r.Context(), req.Username)
		if err != nil {
			utils.BadRequest(w, fmt.Errorf("Database user not found"))
			return
		}
	}

	// ==================== Step 2: Handle database ====================
	registryDB, err = h.DB.DBRegistry.GetMySQLDatabaseByName(r.Context(), req.DatabaseName)

	if err != nil {
		// Create the database, assign user if provided
		if err := h.DB.MySQL.CreateMySQLDatabase(h.mysqlRootDSN, req.DatabaseName, registryUser.Username, registryUser.Password); err != nil {
			utils.ServerError(w, err)
			return
		}
	}

	// Add database to registry
	registryDB.DBName = req.DatabaseName
	registryDB.UserID = registryUser.ID
	if err := h.DB.DBRegistry.InsertMySqlDatabaseRegistry(r.Context(), &registryDB); err != nil {
		if strings.Contains(err.Error(), "mysql_databases_db_name_key") {
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

// ImportMySQLDatabase handles uploading, saving, and optionally executing a .SQL file
// for an existing MySQL database.
//
// This function expects a multipart/form-data POST request with the following fields:
//   - "dbName": the name of the target database (required, must exist in registry)
//   - "sqlFile": the uploaded .sql file to import
//
// Steps:
//  1. Parse the multipart form and validate input.
//  2. Verify that the database exists in the registry.
//  3. Validate that the uploaded file has a ".sql" extension.
//  4. Save the uploaded file to "$USER/projuktisheba/template/database/<dbName>_mysql.sql".
//  5. Execute the SQL statements in the file against the database.
//  6. Return a JSON response indicating success or failure.
func (h *DatabaseManagerHandler) ImportMySQLDatabase(w http.ResponseWriter, r *http.Request) {
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
	registryDB, err := h.DB.DBRegistry.GetMySQLDatabaseByName(r.Context(), dbName)
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

	// Create path to save SQL file
	homeDir, _ := os.UserHomeDir()
	dirPath := filepath.Join(homeDir, "projuktisheba", "templates", "databases")
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		h.errorLog.Println("ERROR_02_ImportDB: failed to create directory:", err)
		utils.ServerError(w, fmt.Errorf("failed to create directory: %w", err))
		return
	}

	savePath := filepath.Join(dirPath, fmt.Sprintf("%s_mysql.sql", registryDB.DBName))
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

	// ==================== Execute SQL statements ====================
	err = h.DB.MySQL.ExecuteSQLFile(h.mysqlRootDSN, dbName, savePath)
	if err != nil {
		h.errorLog.Println("ERROR_05_ImportDB: Failed to execute SQL file:", err)
		utils.ServerError(w, fmt.Errorf("Failed to execute SQL file: %w", err))
	}

	// ==================== Build response ====================
	
	var resp models.Response
	resp.Error = false
	resp.Message = "Database updated successfully"

	utils.WriteJSON(w, http.StatusOK, resp)
}

// DeleteMySQLDatabase permanently drops a MySQL database
// and updates the database registry (soft delete).
func (h *DatabaseManagerHandler) DeleteMySQLDatabase(w http.ResponseWriter, r *http.Request) {
	dbName := strings.TrimSpace(r.URL.Query().Get("db_name"))
	if dbName == "" {
		utils.BadRequest(w, fmt.Errorf("invalid request payload: database_name is required"))
		return
	}

	h.infoLog.Println("Delete request for DB:", dbName)

	// ----------------------------------------
	// 1 Check if database exists in registry
	// ----------------------------------------
	registryDB, err := h.DB.DBRegistry.GetMySQLDatabaseByName(r.Context(), dbName)
	if err != nil {
		utils.BadRequest(w, fmt.Errorf("database '%s' does not exist", dbName))
		return
	}

	// ----------------------------------------
	// 2 Drop the database from MySQL
	// ----------------------------------------
	err = h.DB.MySQL.DropMySQLDatabase(registryDB.User.Username, registryDB.User.Password, "127.0.0.1:3306", dbName)
	if err != nil {
		utils.ServerError(w, fmt.Errorf("failed to drop database: %w", err))
		return
	}

	// ----------------------------------------
	// 3 Soft delete from registry
	// ----------------------------------------
	err = h.DB.DBRegistry.DeleteDatabase(r.Context(), registryDB.ID)
	if err != nil {
		utils.ServerError(w, fmt.Errorf("failed to update registry: %w", err))
		return
	}

	// ----------------------------------------
	// 4️⃣ Response
	// ----------------------------------------
	var resp models.Response
	resp.Error = false
	resp.Message = fmt.Sprintf("Database '%s' deleted successfully", dbName)

	utils.WriteJSON(w, http.StatusOK, resp)
}

// ListMySQLDatabases handles HTTP requests to fetch all MySQL databases.
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
func (h *DatabaseManagerHandler) ListMySQLDatabases(w http.ResponseWriter, r *http.Request) {

	// Fetch databases
	databases, err := h.DB.DBRegistry.GetAllMySQLDatabase(r.Context())
	if err != nil {
		h.errorLog.Println("ERROR_01_ListMySQLUsers: failed to fetch database list:", err)
		utils.BadRequest(w, fmt.Errorf("failed to list databases: %w", err))
		return
	}

	// Fetch database stats
	for _, d := range databases {
		d.DatabaseSizeMB, d.TableCount, _ = h.DB.MySQL.GetMySQLDatabaseStats(r.Context(), h.mysqlRootDSN, d.DBName)
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
func (h *DatabaseManagerHandler) CreateMySQLUser(w http.ResponseWriter, r *http.Request) {
	// Parse JSON body
	var payload models.DBUser
	if err := utils.ReadJSON(w, r, &payload); err != nil {
		h.errorLog.Println("ERROR_01_CreateMySQLUser: failed to parse request:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Validate input
	if payload.Username == "" || payload.Password == "" {
		utils.BadRequest(w, fmt.Errorf("dbName, user, and password are required"))
		return
	}

	// Create the MySQL user
	err := h.DB.MySQL.CreateMySQLUser(r.Context(), h.mysqlRootDSN, payload.Username, payload.Password, []string{})
	if err != nil {
		h.errorLog.Println("ERROR_02_CreateMySQLUser: failed to create mysql user: ", err)
		utils.BadRequest(w, fmt.Errorf("failed to create MySQL user: %w", err))
		return
	}

	// Call DBRegistry to create the MySQL user
	err = h.DB.DBRegistry.InsertMySqlUser(r.Context(), payload)
	if err != nil {
		if strings.Contains(err.Error(), "mysql_db_users_username_key") {
			utils.BadRequest(w, fmt.Errorf("User %s already exist", payload.Username))
			return
		}
		h.errorLog.Println("ERROR_03_CreateMySQLUser: failed to insert into user registry:", err)
		utils.BadRequest(w, fmt.Errorf("failed to insert MySQL user into user registry: %w", err))
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
		Message: "MySQL user created successfully",
		User:    payload.Username,
	}

	// Send response
	utils.WriteJSON(w, http.StatusOK, resp)
}

// ListMySQLUsers handles HTTP requests to fetch all MySQL users.
// It retrieves users from the database registry, optionally decrypts their passwords,
// and returns a structured JSON response.
// Only users that are not soft-deleted are returned.
func (h *DatabaseManagerHandler) ListMySQLUsers(w http.ResponseWriter, r *http.Request) {
	// Fetch all users from the DBRegistry repository
	users, err := h.DB.DBRegistry.ListMySqlUsers(r.Context())
	if err != nil {
		h.errorLog.Println("ERROR_01_ListMySQLUsers: failed to fetch users:", err)
		utils.BadRequest(w, fmt.Errorf("failed to list users: %w", err))
		return
	}

	// Loop through users to decrypt passwords (for admin display)
	// for _, u := range users {
	// 	pass, err := utils.DecryptAES(u.Password)
	// 	if err != nil {
	// 		h.errorLog.Println("ERROR_02_ListMySQLUsers: failed to decrypt password:", err)
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
