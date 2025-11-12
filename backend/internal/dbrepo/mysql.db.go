package dbrepo

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/projuktisheba/vpanel/backend/internal/models"
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
			createUserQuery := fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s';", username, password)
			if _, err := db.Exec(createUserQuery); err != nil {
				return fmt.Errorf("failed to create user '%s': %v", username, err)
			}
			log.Printf("User '%s' created successfully", username)
		} else {
			log.Printf("User '%s' already exists", username)
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

// ListDatabasesWithMeta lists all non-system databases on the MySQL server
// and provides detailed metadata for each database.
//
// Metadata includes:
//   - Database name
//   - Table count
//   - Total data size (MB)
//   - Total index size (MB)
//   - Last table update time
//   - Users with privileges on the database (optional)
//
// Params:
//   - dsn: MySQL connection DSN (e.g., "root:password@tcp(127.0.0.1:3306)/")
//
// Returns:
//   - []models.DatabaseMeta: list of databases with metadata
//   - error: if connection or query fails
func (mysql *MySQLManagerRepo) ListDatabasesWithMeta(dsn string) ([]models.DatabaseMeta, error) {
	// Connect to MySQL server using provided DSN
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	// Validate connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %v", err)
	}

	// Query all non-system databases
	dbQuery := `
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name NOT IN ('information_schema','mysql','performance_schema','sys')
		ORDER BY schema_name;
	`

	dbRows, err := db.Query(dbQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch database list: %v", err)
	}
	defer dbRows.Close()

	// Slice to hold all database metadata
	var databases []models.DatabaseMeta

	// Iterate over each database
	for dbRows.Next() {
		var dbName string
		if err := dbRows.Scan(&dbName); err != nil {
			return nil, err
		}

		// Initialize metadata struct for this database
		meta := models.DatabaseMeta{
			DBName: dbName,
		}

		// Fetch table count, total data size, total index size, last update time
		tableQuery := `
			SELECT 
				COUNT(*) AS table_count,
				IFNULL(SUM(DATA_LENGTH/1024/1024),0) AS data_size_mb,
				IFNULL(SUM(INDEX_LENGTH/1024/1024),0) AS index_size_mb,
				MAX(UPDATE_TIME) AS last_update
			FROM information_schema.tables
			WHERE table_schema = ?;
		`

		var lastUpdate sql.NullTime
		err := db.QueryRow(tableQuery, dbName).Scan(
			&meta.TableCount,
			&meta.DataSizeMB,
			&meta.IndexSizeMB,
			&lastUpdate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch table metadata for database %s: %v", dbName, err)
		}

		// Set last update if available
		if lastUpdate.Valid {
			meta.LastUpdate = &lastUpdate.Time
		}

		// Optional: fetch users who have privileges on this database
		userQuery := `
			SELECT DISTINCT GRANTEE
			FROM information_schema.SCHEMA_PRIVILEGES
			WHERE TABLE_SCHEMA = ?;
		`

		userRows, err := db.Query(userQuery, dbName)
		if err == nil {
			var users []string
			for userRows.Next() {
				var grantee string
				if err := userRows.Scan(&grantee); err == nil {
					users = append(users, grantee)
				}
			}
			meta.Users = users
			userRows.Close()
		}

		// Append this database metadata to the slice
		databases = append(databases, meta)
	}

	// Check for iteration errors
	if err = dbRows.Err(); err != nil {
		return nil, err
	}

	log.Printf("Found %d databases with metadata\n", len(databases))
	return databases, nil
}

// ListTablesWithMeta lists all tables in the given database along with metadata.
//
// Params:
//   - dsn: full MySQL connection string (e.g. "root:password@tcp(127.0.0.1:3306)/")
//   - dbName: name of the database whose tables you want to list.
//
// Returns:
//   - slice of models.TableMeta
//   - error if query or connection fails
func (mysql *MySQLManagerRepo) ListTablesWithMeta(dsn, dbName string) ([]models.TableMeta, error) {
	// Connect to MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %v", err)
	}

	// Query to fetch table metadata
	query := `
		SELECT 
			TABLE_NAME,
			ENGINE,
			ROW_FORMAT,
			TABLE_COLLATION,
			TABLE_ROWS,
			ROUND(DATA_LENGTH/1024/1024, 2) AS data_length_mb,
			ROUND(INDEX_LENGTH/1024/1024, 2) AS index_length_mb,
			CREATE_TIME,
			TABLE_COMMENT
		FROM information_schema.tables
		WHERE table_schema = ?
		ORDER BY TABLE_NAME;
	`

	rows, err := db.Query(query, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch table metadata: %v", err)
	}
	defer rows.Close()

	var tables []models.TableMeta
	for rows.Next() {
		var t models.TableMeta
		if err := rows.Scan(
			&t.TableName,
			&t.Engine,
			&t.RowFormat,
			&t.TableCollation,
			&t.TableRows,
			&t.DataLengthMB,
			&t.IndexLengthMB,
			&t.CreateTime,
			&t.TableComment,
		); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("Found %d tables in database '%s'\n", len(tables), dbName)
	return tables, nil
}
