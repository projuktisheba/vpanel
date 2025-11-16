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

// ============================== User Repository ==============================
type ProjectRepo struct {
	db *pgxpool.Pool
}

func NewProjectRepo(db *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func (r *ProjectRepo) CreateProject(ctx context.Context, p *models.Project) error {
	query := `
        INSERT INTO projects
        (project_name, domain_id, status, project_framework, root_directory, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING id, created_at, updated_at
    `
	row := r.db.QueryRow(ctx, query,
		p.ProjectName, p.DomainID, p.Status, p.ProjectFramework, p.RootDirectory,
	)

	if err := row.Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign key violation
			return errors.New("invalid domain_id")
		}
		return err
	}

	return nil
}

func (r *ProjectRepo) UpdateProject(ctx context.Context, p *models.Project) error {
	query := `
        UPDATE projects
        SET project_name = $1,
            domain_id = $2,
            status = $3,
            project_framework = $4,
            root_directory = $5,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $6
        RETURNING updated_at
    `

	row := r.db.QueryRow(ctx, query,
		p.ProjectName, p.DomainID, p.Status, p.ProjectFramework, p.RootDirectory, p.ID,
	)

	if err := row.Scan(&p.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return errors.New("invalid domain_id")
		}
		if err == pgx.ErrNoRows {
			return errors.New("project not found")
		}
		return err
	}

	return nil
}

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

func (r *ProjectRepo) ListProjects(ctx context.Context) ([]*models.Project, error) {
    rows, err := r.db.Query(ctx, `
        SELECT id, project_name, domain_id, status, project_framework, root_directory, created_at, updated_at
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
            &p.ID, &p.ProjectName, &p.DomainID, &p.Status,
            &p.ProjectFramework, &p.RootDirectory, &p.CreatedAt, &p.UpdatedAt,
        ); err != nil {
            return nil, err
        }
        projects = append(projects, &p)
    }

    return projects, nil
}

func (r *ProjectRepo) ListProjectsByFramework(ctx context.Context, framework string) ([]*models.Project, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, project_name, domain_id, database_id, status, project_framework, root_directory, created_at, updated_at
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
			&p.ID, &p.ProjectName, &p.DomainID, &p.DatabaseID, &p.Status,
			&p.ProjectFramework, &p.RootDirectory, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, &p)
	}

	return projects, nil
}

