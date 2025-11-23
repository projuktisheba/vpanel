package dbrepo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	// Import pgx stdlib driver
	_ "github.com/jackc/pgx/v5/stdlib"
)

// ============================== PostgreSQL Database Manager Repository ==============================
type PostgreSQLManagerRepo struct {
}

func NewPostgreSQLManagerRepo() *PostgreSQLManagerRepo {
	return &PostgreSQLManagerRepo{}
}

// CreatePostgreSQLDatabase creates a PostgreSQL database and optionally a user.
// Note: rootDSN should connect to 'postgres' or 'template1'.
func (pg *PostgreSQLManagerRepo) CreatePostgreSQLDatabase(rootDSN, dbName, username, password string) error {
	// Use "pgx" as the driver name
	db, err := sql.Open("pgx", rootDSN)
	if err != nil {
		return fmt.Errorf("failed to connect as root: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping Postgres: %v", err)
	}

	// 1. Create User if provided
	if username != "" {
		exists, err := pgUserExists(db, username)
		if err != nil {
			return fmt.Errorf("failed to check user: %v", err)
		}
		if !exists {
			// Parameterized queries generally don't work for identifiers/DDL
			query := fmt.Sprintf("CREATE USER \"%s\" WITH PASSWORD '%s';", username, password)
			if _, err := db.Exec(query); err != nil {
				return fmt.Errorf("failed to create user: %v", err)
			}
			log.Printf("User '%s' created successfully", username)
		}
	}

	// 2. Check if database exists
	exists, err := pgDatabaseExists(db, dbName)
	if err != nil {
		return fmt.Errorf("failed to check database: %v", err)
	}

	if !exists {
		// In Postgres, CREATE DATABASE cannot run in a transaction block.
		// pgx stdlib handles this correctly if not inside db.Begin()
		query := fmt.Sprintf("CREATE DATABASE \"%s\" ENCODING 'UTF8';", dbName)
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create database: %v", err)
		}
		log.Printf("Database '%s' created successfully", dbName)
	} else {
		return fmt.Errorf("database '%s' already exists", dbName)
	}

	// 3. Grant privileges
	if username != "" {
		// Grant connect privileges
		query := fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE \"%s\" TO \"%s\";", dbName, username)
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to grant privileges: %v", err)
		}

		// Change owner (optional, but recommended for full control)
		alter := fmt.Sprintf("ALTER DATABASE \"%s\" OWNER TO \"%s\";", dbName, username)
		if _, err := db.Exec(alter); err != nil {
			log.Printf("Warning: failed to set owner (non-critical): %v", err)
		}

		log.Printf("Granted privileges on '%s' to '%s'", dbName, username)
	}

	return nil
}

// ExecuteSQLFile executes a .SQL file against the target database.
// Accepts the custom PGSQLFile type for the file path.
func (pg *PostgreSQLManagerRepo) ExecuteSQLFile(rootDSN, dbName string, filePath string) error {
	// 1. Switch connection to the target database
	// pgx connection strings are robust; appending "/foo" usually overrides previous values.
	// Ideally, use pgx.ParseConfig for full robustness, but this string manip works for stdlib.
	targetDSN := fmt.Sprintf("%s/%s", rootDSN, dbName)

	db, err := sql.Open("pgx", targetDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to target db '%s': %v", dbName, err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping target db '%s': %v", dbName, err)
	}

	// 2. Read SQL file
	pathStr := string(filePath)
	sqlContent, err := os.ReadFile(pathStr)
	if err != nil {
		return fmt.Errorf("failed to read SQL file '%s': %v", pathStr, err)
	}

	// 3. Execute
	// pgx supports multiple statements in one Exec call naturally
	if _, err := db.Exec(string(sqlContent)); err != nil {
		return fmt.Errorf("failed to execute SQL file: %v", err)
	}

	log.Printf("SQL file '%s' executed successfully on '%s'", pathStr, dbName)
	return nil
}

// DropPostgreSQLDatabase disconnects active users and drops the database.
func (pg *PostgreSQLManagerRepo) DropPostgreSQLDatabase(rootDSN, dbName, username string) error {
	db, err := sql.Open("pgx", rootDSN)
	if err != nil {
		return fmt.Errorf("failed to connect as root: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping Postgres: %v", err)
	}

	// 1. Check existence
	exists, err := pgDatabaseExists(db, dbName)
	if err != nil {
		return fmt.Errorf("failed to check database: %v", err)
	}
	if !exists {
		return fmt.Errorf("database '%s' does not exist", dbName)
	}

	// 2. Kill active connections
	// Postgres pgx driver might maintain its own pool; we ensure all external connections are killed.
	killQuery := `
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = $1
		AND pid <> pg_backend_pid();
	`
	_, err = db.Exec(killQuery, dbName)
	if err != nil {
		return fmt.Errorf("failed to kill active connections: %v", err)
	}

	// 3. Drop Database
	dropQuery := fmt.Sprintf("DROP DATABASE \"%s\";", dbName)
	if _, err := db.Exec(dropQuery); err != nil {
		return fmt.Errorf("failed to drop database: %v", err)
	}
	log.Printf("Database '%s' deleted successfully", dbName)

	// 4. Optional: Drop User
	if username != "" {
		// Note: This fails if the user owns objects in other databases
		dropUser := fmt.Sprintf("DROP USER IF EXISTS \"%s\";", username)
		if _, err := db.Exec(dropUser); err != nil {
			log.Printf("Warning: Could not drop user '%s' (may own other objects): %v", username, err)
		}
	}

	return nil
}

// ResetPostgreSQLDatabase truncates all tables in the 'public' schema.
func (pg *PostgreSQLManagerRepo) ResetPostgreSQLDatabase(dsn string, dbName string) error {
	// Ensure we are connected to the specific DB
	targetDSN := dsn
	if !strings.Contains(dsn, "dbname=") {
		targetDSN = fmt.Sprintf("%s/%s", dsn, dbName)
	}

	db, err := sql.Open("pgx", targetDSN)
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}
	defer db.Close()

	// 1. Discover tables
	query := `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public';
	`
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return err
		}
		tables = append(tables, t)
	}

	if len(tables) == 0 {
		log.Println("No tables to reset.")
		return nil
	}

	// 2. Truncate with Cascade
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, t := range tables {
		// RESTART IDENTITY resets sequences to 1
		// CASCADE deletes dependent rows in other tables
		q := fmt.Sprintf("TRUNCATE TABLE \"%s\" RESTART IDENTITY CASCADE;", t)
		if _, err := tx.Exec(q); err != nil {
			return fmt.Errorf("failed to truncate '%s': %v", t, err)
		}
	}

	return tx.Commit()
}

// GetPostgreSQLDatabaseStats returns size (MB) and table count.
func (pg *PostgreSQLManagerRepo) GetPostgreSQLDatabaseStats(ctx context.Context, dsn, dbName string) (float64, int, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return 0, 0, err
	}
	defer db.Close()

	// -------------------------
	// 1. Get database size
	// -------------------------
	var sizeBytes int64
	err = db.QueryRowContext(ctx, 
		"SELECT pg_database_size($1)", 
		dbName,
	).Scan(&sizeBytes)

	if err != nil {
		return 0, 0, fmt.Errorf("failed to get db size: %w", err)
	}

	// -------------------------
	// 2. Get correct table count
	// -------------------------
	var tableCount int

	// This ALWAYS works and does NOT depend on pg_tables, search_path, or user visibility.
	err = db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM pg_class
		WHERE relkind = 'r'
		AND relnamespace = 'public'::regnamespace
	`).Scan(&tableCount)

	if err != nil {
		return float64(sizeBytes) / (1024 * 1024), 0,
			fmt.Errorf("failed to count tables: %w", err)
	}

	return float64(sizeBytes) / (1024 * 1024), tableCount, nil
}


// CreatePostgreSQLUser creates a user and grants privileges on multiple DBs.
func (pg *PostgreSQLManagerRepo) CreatePostgreSQLUser(ctx context.Context, rootDSN, username, password string, dbNames []string) error {
	db, err := sql.Open("pgx", rootDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	// 1. Create User
	exists, err := pgUserExists(db, username)
	if err != nil {
		return err
	}
	if !exists {
		query := fmt.Sprintf("CREATE USER \"%s\" WITH PASSWORD '%s';", username, password)
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("create user failed: %v", err)
		}
	}

	// 2. Grant Privileges
	for _, name := range dbNames {
		grant := fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE \"%s\" TO \"%s\";", name, username)
		if _, err := db.ExecContext(ctx, grant); err != nil {
			return fmt.Errorf("grant failed on %s: %v", name, err)
		}
	}

	return nil
}

// ============================== Internal Helpers ==============================

func pgDatabaseExists(db *sql.DB, dbName string) (bool, error) {
	var exists int
	// pgx supports $1 syntax
	err := db.QueryRow("SELECT 1 FROM pg_database WHERE datname = $1", dbName).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return exists == 1, err
}

func pgUserExists(db *sql.DB, username string) (bool, error) {
	var exists int
	err := db.QueryRow("SELECT 1 FROM pg_roles WHERE rolname = $1", username).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return exists == 1, err
}
