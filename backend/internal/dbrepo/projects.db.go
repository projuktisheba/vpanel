package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projuktisheba/vpanel/backend/internal/models"
)

// ============================== Project Repository ==============================
type ProjectRepo struct {
	db *pgxpool.Pool
}

func NewProjectRepo(db *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{db: db}
}
// CreateProject inserts a new project
func (r *ProjectRepo) CreateProject(ctx context.Context, p *models.Project) error {
	query := `
        INSERT INTO projects
        (project_name, domain_name, db_name, project_framework, template_path, project_directory, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING id, created_at, updated_at
    `
	row := r.db.QueryRow(ctx, query,
		p.ProjectName,
		p.DomainName,
		p.DBName,
		p.ProjectFramework,
		p.TemplatePath,
		p.ProjectDirectory,
		p.Status,
	)

	if err := row.Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError

		// Foreign key violation for domain_name
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return errors.New("invalid domain_name: domain does not exist")
		}
		return err
	}
	return nil
}

// UpdateProject updates a project by ID
func (r *ProjectRepo) UpdateProject(ctx context.Context, p *models.Project) error {
	query := `
        UPDATE projects
        SET project_name = $1,
            domain_name = $2,
            db_name = $3,
            project_framework = $4,
            template_path = $5,
            project_directory = $6,
            status = $7,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $8
        RETURNING updated_at
    `

	row := r.db.QueryRow(ctx, query,
		p.ProjectName,
		p.DomainName,
		p.DBName,
		p.ProjectFramework,
        p.TemplatePath,
        p.ProjectDirectory,
		p.Status,
		p.ID,
	)

	if err := row.Scan(&p.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError

		// Invalid domain name
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return errors.New("invalid domain_name")
		}

		// No project found
		if err == pgx.ErrNoRows {
			return errors.New("project not found")
		}
		return err
	}

	return nil
}

// UpdateProjectStatus updates only the status of a project
func (r *ProjectRepo) UpdateProjectStatus(ctx context.Context, id int64, status string) (time.Time, error) {
	query := `
        UPDATE projects
        SET status = $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
        RETURNING updated_at
    `

	var updatedAt time.Time
	row := r.db.QueryRow(ctx, query, status, id)

	if err := row.Scan(&updatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return time.Time{}, errors.New("project not found")
		}
		return time.Time{}, err
	}

	return updatedAt, nil
}

// DeleteProject deletes a project by ID
func (r *ProjectRepo) DeleteProject(ctx context.Context, id int64) error {
	cmd, err := r.db.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("project not found")
	}
	return nil
}

// ListProjects returns all projects
func (r *ProjectRepo) ListProjects(ctx context.Context) ([]*models.Project, error) {
	rows, err := r.db.Query(ctx, `
        SELECT 
            id, 
            project_name, 
            domain_name, 
            db_name, 
            project_framework,
            template_path,
            project_directory,
            status, 
            created_at, 
            updated_at
        FROM projects
        ORDER BY id DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*models.Project

	for rows.Next() {
		var p models.Project
		if err := rows.Scan(
			&p.ID,
			&p.ProjectName,
			&p.DomainName,
			&p.DBName,
			&p.ProjectFramework,
			&p.TemplatePath,
			&p.ProjectDirectory,
			&p.Status,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, &p)
	}

	return projects, nil
}

// ListProjectsByFramework returns projects filtered by framework
func (r *ProjectRepo) ListProjectsByFramework(ctx context.Context, framework string) ([]*models.Project, error) {
	rows, err := r.db.Query(ctx, `
		SELECT 
			id, 
			project_name, 
			domain_name, 
			db_name, 
			project_framework,
			template_path,
			project_directory,
			status, 
			created_at, 
			updated_at
		FROM projects
		WHERE project_framework = $1
		ORDER BY id DESC
	`, framework)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*models.Project

	for rows.Next() {
		var p models.Project

		if err := rows.Scan(
			&p.ID,
			&p.ProjectName,
			&p.DomainName,
			&p.DBName,
			&p.ProjectFramework,
			&p.TemplatePath,
			&p.ProjectDirectory,
			&p.Status,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}

		projects = append(projects, &p)
	}

	return projects, nil
}
