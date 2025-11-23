package dbrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projuktisheba/vpanel/backend/internal/models"
)

// DatabaseRegistryRepo provides methods to manage user and database registries.
// It supports creating, updating, soft-deleting, and checking deletion status for both users and databases.
// All operations use pgxpool.Pool for PostgreSQL database interaction and respect context for timeouts/cancellation.
type DatabaseRegistryRepo struct {
	db *pgxpool.Pool
}

// newDatabaseRegistryRepo creates a new instance of DatabaseRegistryRepo.
// Params:
// - db: pgxpool.Pool instance connected to the PostgreSQL database.
// Returns a pointer to DatabaseRegistryRepo.
func newDatabaseRegistryRepo(db *pgxpool.Pool) *DatabaseRegistryRepo {
	return &DatabaseRegistryRepo{db: db}
}

// InsertUserRegistry inserts a new user into db_users table.
// Params:
// - ctx: context for request cancellation/timeouts
// - u: pointer to DBUser model containing Username and Password (encrypted)
// Returns nil if successful, or an error if insertion fails or username already exists.
func (r *DatabaseRegistryRepo) InsertDBUser(ctx context.Context, u *models.DBUser) error {
	query := `
		INSERT INTO db_users (username, password, user_type, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query, u.Username, u.Password, u.UserType)
	err := row.Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			if pgErr.ConstraintName == "idx_db_users_username" {
				return errors.New("Username already exists")
			}
		}
		return err
	}

	return nil
}

// UpdateRegistry updates an existing user's username or password.
// Params:
// - ctx: context for request cancellation/timeouts
// - u: pointer to DBUser model with ID set; Username or Password may be updated
// Returns nil if successful, or an error if update fails.
func (r *DatabaseRegistryRepo) UpdateUserRegistry(ctx context.Context, u *models.DBUser) error {
	query := `
		UPDATE db_users
		SET username = $1,
		    password = $2,
		    user_type = $3,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING updated_at
	`
	row := r.db.QueryRow(ctx, query, u.Username, u.Password, u.UserType, u.ID)
	return row.Scan(&u.UpdatedAt)
}

// DeleteUserUserRegistry performs a soft delete of a user by setting deleted_at timestamp.
// Params:
// - ctx: context
// - userID: ID of the user to soft-delete
// Returns nil if successful, or an error if user does not exist or already deleted.
func (r *DatabaseRegistryRepo) DeleteUserFromUserRegistry(ctx context.Context, userID int64) error {
	query := `
		UPDATE db_users
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`
	cmdTag, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("user not found or already deleted")
	}
	return nil
}

// GetUserByUsername retrieves a DB user by username.
// Returns the user struct if found and not soft-deleted.
// Params:
// - ctx: context
// - username: username to search
// Returns:
// - *models.DBUser, error
func (r *DatabaseRegistryRepo) GetUserByUsername(ctx context.Context, username string) (models.DBUser, error) {
	query := `
		SELECT id, username, password, user_type, created_at, updated_at, deleted_at
		FROM db_users
		WHERE username = $1
	`

	var user models.DBUser
	var deletedAt *string

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.UserType,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return user, fmt.Errorf("user '%s' not found", username)
		}
		return user, err
	}

	// Check soft delete
	if deletedAt != nil {
		return user, fmt.Errorf("user '%s' is deleted", username)
	}

	return user, nil
}

// GetAllUsers retrieves all users that are not soft-deleted.
// Returns a slice of DBUser records.
// Params:
// - ctx: context
// Returns:
// - ([]models.DBUser, error)
func (r *DatabaseRegistryRepo) GetAllUsers(ctx context.Context, userType string) ([]*models.DBUser, error) {
	UserTypeFilter := ""
	if strings.TrimSpace(userType) != "" {
		UserTypeFilter = fmt.Sprintf("AND user_type='%s'", userType)
	}
	query := fmt.Sprintf(`
		SELECT id, username, password, user_type, created_at, updated_at
		FROM db_users
		WHERE deleted_at IS NULL %s
		ORDER BY id ASC
	`, UserTypeFilter)

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %v", err)
	}
	defer rows.Close()

	var users []*models.DBUser

	for rows.Next() {
		var u models.DBUser
		err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Password,
			&u.UserType,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return users, nil
}

// InsertDatabaseRegistry inserts a new database into databases table.
// Params:
// - ctx: context for request cancellation/timeouts
// - d: pointer to Database model containing DBName and UserID
// Returns nil if successful, or an error if insertion fails.
func (r *DatabaseRegistryRepo) InsertDatabaseRegistry(ctx context.Context, d *models.Database) error {
	query := `
		INSERT INTO databases (db_name, db_type, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`
	row := r.db.QueryRow(ctx, query, d.DBName, d.DBType, d.UserID)
	return row.Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
}

// UpdateDatabaseRegistry updates the database name or user of an existing record.
// Params:
// - ctx: context for request cancellation/timeouts
// - d: pointer to Database model with ID set; DBName or UserID may be updated
// Returns nil if successful, or an error if update fails.
func (r *DatabaseRegistryRepo) UpdateDatabaseRegistry(ctx context.Context, d *models.Database) error {
	query := `
		UPDATE databases
		SET db_name = COALESCE(NULLIF($1,''), db_name),
			db_type = COALESCE(NULLIF($2,''), db_type),
			user_id = COALESCE(NULLIF($3,0), user_id),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING updated_at
	`
	row := r.db.QueryRow(ctx, query, d.DBName, d.DBType, d.UserID, d.ID)
	return row.Scan(&d.UpdatedAt)
}

// DeleteDatabase performs a soft delete of a database by setting deleted_at timestamp.
// Params:
// - ctx: context
// - dbID: ID of the database to soft-delete
// Returns nil if successful, or an error if database does not exist or already deleted.
func (r *DatabaseRegistryRepo) DeleteDatabase(ctx context.Context, dbID int64) error {
	query := `
		DELETE FROM databases
		WHERE id = $1
	`
	cmdTag, err := r.db.Exec(ctx, query, dbID)
	if err != nil {
		return fmt.Errorf("failed to delete database: %v", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("database not found")
	}

	return nil
}

// GetDatabaseByName retrieves a database record by its db_name.
// Returns the database struct if found and not soft-deleted.
// Params:
// - ctx: context
// - dbName: the name of the database to search
// Returns:
// - *models.Database, error
func (r *DatabaseRegistryRepo) GetDatabaseByName(ctx context.Context, dbName string) (models.Database, error) {
	query := `
        SELECT
            d.id,
            d.db_name,
            d.db_type,
            d.user_id,
            d.deleted_at,
            COALESCE(u.username, '') AS username,
            COALESCE(u.password, '') AS password,
			d.created_at,
			d.updated_at
        FROM databases d
        LEFT JOIN db_users u ON d.user_id = u.id
        WHERE d.db_name = $1
        LIMIT 1
    `

	var d models.Database
	d.User = &models.DBUser{} // initialize to prevent nil pointer dereference
	var deletedAt *time.Time

	err := r.db.QueryRow(ctx, query, dbName).Scan(
		&d.ID,
		&d.DBName,
		&d.DBType,
		&d.UserID,
		&deletedAt,
		&d.User.Username,
		&d.User.Password,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return d, fmt.Errorf("database '%s' not found", dbName)
		}
		return d, err
	}

	return d, nil
}

// GetAllDatabase returns all saved databases from the registry,
// including username and password using LEFT JOIN so entries with no user still appear.
func (r *DatabaseRegistryRepo) GetAllDatabase(ctx context.Context, databaseType string) ([]*models.Database, error) {
	filter := ""
	if strings.TrimSpace(databaseType) != "" {
		filter = fmt.Sprintf("WHERE db_type='%s'", databaseType)
	}
	query := fmt.Sprintf(`
        SELECT
            d.id,
            d.db_name,
            d.db_type,
			d.user_id,
			d.deleted_at,
            COALESCE(u.username, '') AS username,
            COALESCE(u.password, '') AS password,
			d.created_at,
			d.updated_at
        FROM databases d
        LEFT JOIN db_users u ON d.user_id = u.id
		%s
        ORDER BY d.id ASC
    `, filter)

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []*models.Database

	for rows.Next() {
		var d models.Database
		var u models.DBUser
		var deletedAt *time.Time
		err := rows.Scan(
			&d.ID,
			&d.DBName,
			&d.DBType,
			&d.UserID,
			&deletedAt,
			&u.Username,
			&u.Password,
			&d.CreatedAt,
			&d.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if deletedAt == nil {
			d.User = &u
			databases = append(databases, &d)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return databases, nil
}
