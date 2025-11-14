package handlers

import (
	"fmt"
	"log"
	"net/http"
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
			utils.ServerError(w, fmt.Errorf("failed to create database: %w", err))
			return
		}
	}

	// Add database to registry
	registryDB.DBName = req.DatabaseName
	registryDB.UserID = registryUser.ID
	if err := h.DB.DBRegistry.InsertMySqlDatabaseRegistry(r.Context(), &registryDB); err != nil {
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

func (h *DatabaseManagerHandler) ListMySQLUsers(w http.ResponseWriter, r *http.Request) {

	// Fetch users
	users, err := h.DB.DBRegistry.GetAllMySqlUsers(r.Context())
	if err != nil {
		h.errorLog.Println("ERROR_01_ListMySQLUsers: failed to fetch users:", err)
		utils.BadRequest(w, fmt.Errorf("failed to list users: %w", err))
		return
	}

	// Prepare successful response
	resp := struct {
		Error   bool             `json:"error"`   // Indicates if an error occurred
		Message string           `json:"message"` // Response message
		Count   int              `json:"count"`   // Number of databases
		DBUsers []*models.DBUser `json:"dbUsers"` // List of database names
	}{
		Error:   false,
		Message: "Users fetched successfully",
		Count:   len(users),
		DBUsers: users,
	}

	// Send response
	utils.WriteJSON(w, http.StatusOK, resp)
}
