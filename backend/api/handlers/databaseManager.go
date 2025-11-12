package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/projuktisheba/vpanel/backend/internal/dbrepo"
	"github.com/projuktisheba/vpanel/backend/internal/models"
	"github.com/projuktisheba/vpanel/backend/utils"
)

type DatabaseManagerHandler struct {
	MySQLDB *MySQLDB
}
type MySQLDB struct{
	rootDSN string
	DB        *dbrepo.DBRepository
	infoLog   *log.Logger
	errorLog  *log.Logger
}
func newMySQLManagerRepo(db *dbrepo.DBRepository, infoLog, errorLog *log.Logger) *MySQLDB{
	return &MySQLDB{
		rootDSN: "root:G*vA#N+aQk1&k@7j=a8!3xJzs98$*QKo@tcp(127.0.0.1:3306)/?multiStatements=true", // Must have admin privileges
		DB:        db,
		infoLog:   infoLog,
		errorLog:  errorLog,
	}
}
func NewDatabaseManagerHandler(db *dbrepo.DBRepository, infoLog, errorLog *log.Logger) *DatabaseManagerHandler {
	return &DatabaseManagerHandler{
		MySQLDB: newMySQLManagerRepo(db, infoLog, errorLog),
	}
}


func (h *MySQLDB) CreateMySQLDatabase(w http.ResponseWriter, r *http.Request) {
	type payload struct {
		DatabaseName string `json:"database_name"`
		Username     string `json:"database_user"`
	}

	var req payload
	if err := utils.ReadJSON(w, r, &req); err != nil {
		h.errorLog.Println("ERROR_01_Signin: invalid JSON:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request payload: %w", err))
		return
	}

	h.infoLog.Println("Payload: ", req)

	//TODO: implement database creation algorithm
	if err := h.DB.MySQL.CreateMySQLDatabase(h.rootDSN, req.DatabaseName, req.Username, "apppass123"); err != nil {
		utils.ServerError(w, err)
		return
	}

	// Build response
	var resp models.Response
	resp.Error = true
	resp.Message = req.DatabaseName + " created successfully"
	utils.WriteJSON(w, http.StatusOK, resp)
}

// ListTables handles the request to fetch all MySQL tables with metadata.
//
// Request Body (JSON):
// {
//   "dsn": "root:password@tcp(127.0.0.1:3306)/your_database"
// }
//
// Response:
// {
//   "error": false,
//   "message": "Tables fetched successfully",
//   "count": 3,
//   "tables": [
//     { "table_name": "users", "engine": "InnoDB", ... }
//   ]
// }
//
// Expected behavior:
// - Reads DSN from JSON request.
// - Validates input.
// - Connects to MySQL and fetches all tables with metadata.
// - Returns structured JSON response.
func (h *MySQLDB) ListTables(w http.ResponseWriter, r *http.Request) {
	// Define request structure
	type requestPayload struct {
		TableName string `json:"table_name"`
	}

	var req requestPayload

	// Parse incoming JSON
	if err := utils.ReadJSON(w, r, &req); err != nil {
		h.errorLog.Println("ERROR_01_ListTables: invalid JSON:", err)
		utils.BadRequest(w, fmt.Errorf("invalid request payload: %w", err))
		return
	}

	// Trim and validate DSN
	req.TableName = strings.TrimSpace(req.TableName)
	if req.TableName == "" {
		utils.BadRequest(w, errors.New("DSN is required"))
		return
	}

	// Fetch all tables and metadata
	tables, err := h.DB.MySQL.ListTablesWithMeta(h.rootDSN, req.TableName)
	if err != nil {
		h.errorLog.Println("ERROR_02_ListTables: failed to fetch tables:", err)
		utils.BadRequest(w, fmt.Errorf("failed to list tables: %w", err))
		return
	}

	// Build success response payload
	resp := struct {
		Error   bool        `json:"error"`   // Indicates whether an error occurred
		Message string      `json:"message"` // Response message
		Count   int         `json:"count"`   // Number of tables returned
		Tables  interface{} `json:"tables"`  // List of tables with metadata
	}{
		Error:   false,
		Message: "Tables fetched successfully",
		Count:   len(tables),
		Tables:  tables,
	}

	// Send JSON response
	utils.WriteJSON(w, http.StatusOK, resp)
}

// ListDatabases handles HTTP requests to fetch all MySQL databases.
//
// Request Body (JSON):
// {
//   "dsn": "root:password@tcp(127.0.0.1:3306)/"
// }
//
// Response (JSON):
// {
//   "error": false,
//   "message": "Databases fetched successfully",
//   "count": 3,
//   "databases": ["app_db", "test_db", "analytics"]
// }
//
// Expected behavior:
// - Reads MySQL DSN from JSON body.
// - Validates input.
// - Lists all non-system databases using dbrepo.MySQLManagerRepo.
// - Returns structured JSON response with consistent format.
func (h *MySQLDB) ListDatabases(w http.ResponseWriter, r *http.Request) {


	// Fetch databases
	databases, err := h.DB.MySQL.ListDatabasesWithMeta(h.rootDSN)
	if err != nil {
		h.errorLog.Println("ERROR_02_ListDatabases: failed to fetch database list:", err)
		utils.BadRequest(w, fmt.Errorf("failed to list databases: %w", err))
		return
	}

	// Prepare successful response
	resp := struct {
		Error      bool     `json:"error"`      // Indicates if an error occurred
		Message    string   `json:"message"`    // Response message
		Count      int      `json:"count"`      // Number of databases
		Databases  []models.DatabaseMeta `json:"databases"`  // List of database names
	}{
		Error:     false,
		Message:   "Databases fetched successfully",
		Count:     len(databases),
		Databases: databases,
	}

	// Send response
	utils.WriteJSON(w, http.StatusOK, resp)
}
