package dbrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

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
		log.Printf("Database '%s' already exists", dbName)
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

// DropMySQLDatabase deletes a MySQL database if it exists.
// Params:
// - dsn: connection string for a MySQL user with privileges to drop databases (e.g., "user:password@tcp(127.0.0.1:3306)/")
// - dbName: name of the database to delete
// Returns nil if the database was successfully deleted or didn't exist, otherwise returns an error.
func (mysql *MySQLManagerRepo) DropMySQLDatabase(user, password, host, dbName string) error {
	// 1. Construct DSN (Data Source Name) for connecting to MySQL
	// Note: Do NOT specify a database here because you cannot drop a database
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/", user, password, host)

	// 2. Open a connection to MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	// Ensure the database connection is closed when function exits
	defer db.Close()

	// 3. Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL: %v", err)
	}

	// 4. Prepare the SQL query to drop the database if it exists
	query := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName)

	// 5. Execute the query
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop database: %v", err)
	}

	// 6. Inform the user
	fmt.Printf("Database %s deleted successfully.\n", dbName)
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
func (mysql *MySQLManagerRepo) GetMySQLDatabaseStats(ctx context.Context,dsn, dbName string) (float64, int, error) {
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
