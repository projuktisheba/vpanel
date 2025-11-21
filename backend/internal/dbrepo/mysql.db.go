package dbrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// ============================== MySQL Database Manager Repository ==============================
type MySQLManagerRepo struct {
}

func NewMySQLManagerRepo() *MySQLManagerRepo {
	return &MySQLManagerRepo{}
}

// CreateMySQLDatabase creates a MySQL database and optionally a user if not exist.
// Params:
// - rootDSN: connection string for root/admin user (e.g., "root:password@tcp(127.0.0.1:3306)/")
// - dbName: name of the database to create
// - username, password: if username != "", ensures that user exists (creates if not)
// Returns nil if everything is OK, or error otherwise.
func (mysql *MySQLManagerRepo) CreateMySQLDatabase(rootDSN, dbName, username, password string) error {
	// Connect as root
	db, err := sql.Open("mysql", rootDSN)
	if err != nil {
		return fmt.Errorf("failed to connect as root: %v", err)
	}
	defer db.Close()

	// Validate connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL: %v", err)
	}

	// Create user if username is provided
	if username != "" {
		exists, err := userExists(db, username)
		if err != nil {
			return fmt.Errorf("failed to check user: %v", err)
		}
		if !exists {
			log.Printf("User '%s' created successfully", username)
			return errors.New("User not found")
		}
	}

	// Check if database exists
	exists, err := databaseExists(db, dbName)
	if err != nil {
		return fmt.Errorf("failed to check database: %v", err)
	}

	if !exists {
		if _, err := db.Exec("CREATE DATABASE `" + dbName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"); err != nil {
			return fmt.Errorf("failed to create database: %v", err)
		}
		log.Printf("Database '%s' created successfully", dbName)
	} else {
		return fmt.Errorf("Database '%s' already exists", dbName)
	}

	// Grant privileges if user is specified
	if username != "" {
		grantQuery := fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%%'; FLUSH PRIVILEGES;", dbName, username)
		if _, err := db.Exec(grantQuery); err != nil {
			return fmt.Errorf("failed to grant privileges: %v", err)
		}
		log.Printf("Granted privileges on '%s' to '%s'", dbName, username)
	}

	return nil
}

// ExecuteSQLFile executes a .SQL file against the given MySQL database.
// It reads the file, splits statements by semicolon, and executes them sequentially.
//
// Parameters:
//   - dbName: the name of the target MySQL database.
//   - filePath: the path to the .SQL file to be executed.
//
// Returns:
//   - error: returns an error if reading the file, connecting to the database,
//     or executing any SQL statement fails.
func (mysql *MySQLManagerRepo) ExecuteSQLFile(rootDSN, dbName, filePath string) error {
	// -------------------- Step 1: Connect as root --------------------
	db, err := sql.Open("mysql", rootDSN)
	if err != nil {
		return fmt.Errorf("failed to connect as root: %v", err)
	}
	defer db.Close()

	// Validate connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL: %v", err)
	}

	// -------------------- Step 2: Read SQL file --------------------
	sqlContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file '%s': %v", filePath, err)
	}

	// -------------------- Step 3: Switch to target database --------------------
	if _, err := db.Exec(fmt.Sprintf("USE `%s`;", dbName)); err != nil {
		return fmt.Errorf("failed to switch to database '%s': %v", dbName, err)
	}

	// -------------------- Step 4: Execute SQL file as batch --------------------
	if _, err := db.Exec(string(sqlContent)); err != nil {
		return fmt.Errorf("failed to execute SQL file '%s': %v", filePath, err)
	}

	log.Printf("SQL file '%s' executed successfully for database '%s'", filePath, dbName)
	return nil
}

// DropMySQLDatabase deletes a MySQL database if it exists.
// Params:
// - dsn: connection string for a MySQL user with privileges to drop databases (e.g., "user:password@tcp(127.0.0.1:3306)/")
// - dbName: name of the database to delete
// Returns nil if the database was successfully deleted or didn't exist, otherwise returns an error.
func (mysql *MySQLManagerRepo) DropMySQLDatabase(rootDSN, dbName, username string) error {
	// Connect as root (must NOT specify a database)
	db, err := sql.Open("mysql", rootDSN)
	if err != nil {
		return fmt.Errorf("failed to connect as root: %v", err)
	}
	defer db.Close()

	// Validate connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL: %v", err)
	}

	// 1. Check if database exists
	exists, err := databaseExists(db, dbName)
	if err != nil {
		return fmt.Errorf("failed to check database: %v", err)
	}

	if !exists {
		return fmt.Errorf("database '%s' does not exist", dbName)
	}

	// 2. Kill active connections to the database
	killQuery := fmt.Sprintf(`
        SELECT CONCAT('KILL ', id, ';')
        FROM information_schema.processlist
        WHERE db = '%s';
    `, dbName)

	rows, err := db.Query(killQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var killCmd string
			rows.Scan(&killCmd)
			db.Exec(killCmd) // execute kill command
		}
	}

	// 3. Drop the database
	dropQuery := fmt.Sprintf("DROP DATABASE `%s`;", dbName)
	if _, err := db.Exec(dropQuery); err != nil {
		return fmt.Errorf("failed to drop database '%s': %v", dbName, err)
	}

	log.Printf("Database '%s' deleted successfully", dbName)

	// 4. (Optional) Remove user privileges or drop user
	if username != "" {
		revoke := fmt.Sprintf("REVOKE ALL PRIVILEGES, GRANT OPTION FROM '%s'@'%%';", username)
		db.Exec(revoke)

		// Optional: DROP USER entirely
		// dropUser := fmt.Sprintf("DROP USER IF EXISTS '%s'@'%%';", username)
		// db.Exec(dropUser)
	}

	return nil
}


// Helper: check if a database exists
func databaseExists(db *sql.DB, dbName string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = ?", dbName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetDatabaseStats connects to a MySQL database and retrieves:
// - Total database size in MB
// - Total number of tables
// Params:
// - dsn: MySQL DSN like "user:password@tcp(127.0.0.1:3306)/"
// - dbName: target database name
// Returns: sizeMB, tableCount, error
func (mysql *MySQLManagerRepo) GetMySQLDatabaseStats(ctx context.Context, dsn, dbName string) (float64, int, error) {
	// Connect to MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to connect: %w", err)
	}
	defer db.Close()

	// Query to get size and table count
	query := `
        SELECT 
            ROUND(IFNULL(SUM(data_length + index_length) / 1024 / 1024, 0), 2) AS size_mb,
            COUNT(*) AS table_count
        FROM information_schema.tables
        WHERE table_schema = ?
    `

	var sizeMB float64
	var tableCount int

	err = db.QueryRowContext(ctx, query, dbName).Scan(&sizeMB, &tableCount)
	if err != nil {
		return 0, 0, fmt.Errorf("query error: %w", err)
	}

	return sizeMB, tableCount, nil
}
func (mysql *MySQLManagerRepo) CreateMySQLUser(ctx context.Context, rootDSN, username, password string, dbNames []string) error {
	// Connect to MySQL as root
	db, err := sql.Open("mysql", rootDSN)
	if err != nil {
		return fmt.Errorf("failed to connect as root: %v", err)
	}
	defer db.Close()

	// Validate connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL: %v", err)
	}

	// Check if user exists
	exists, err := userExists(db, username)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %v", err)
	}

	if !exists {
		// Create user
		createQuery := fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s';", username, password)
		if _, err := db.ExecContext(ctx, createQuery); err != nil {
			return fmt.Errorf("failed to create user: %v", err)
		}
		log.Printf("User '%s' created successfully", username)
	} else {
		log.Printf("User '%s' already exists", username)
	}

	// Grant privileges on multiple databases
	for _, dbName := range dbNames {
		grantQuery := fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'%%';", dbName, username)
		if _, err := db.Exec(grantQuery); err != nil {
			return fmt.Errorf("failed to grant privileges on '%s': %v", dbName, err)
		}
		log.Printf("Granted privileges on '%s' to '%s'", dbName, username)
	}

	// Apply privileges
	if _, err := db.Exec("FLUSH PRIVILEGES;"); err != nil {
		return fmt.Errorf("failed to flush privileges: %v", err)
	}

	return nil
}

// Helper: check if a user exists
func userExists(db *sql.DB, username string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM mysql.user WHERE user = ?", username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
