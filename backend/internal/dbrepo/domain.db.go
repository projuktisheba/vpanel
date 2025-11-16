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
type DomainRepo struct {
	db *pgxpool.Pool
}

func NewDomainRepo(db *pgxpool.Pool) *DomainRepo {
	return &DomainRepo{db: db}
}

// CreateDomain inserts a new domain record
func (r *DomainRepo) CreateDomain(ctx context.Context, d *models.Domain) error {
	query := `
		INSERT INTO domains (domain, ssl_update_date, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query, d.Domain, d.SSLUpdateDate)

	err := row.Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique violation
			if pgErr.ConstraintName == "domains_domain_key" {
				return errors.New("this domain already exists")
			}
		}

		return err
	}

	return nil
}

// UpdateDomain updates domain fields
func (r *DomainRepo) UpdateDomain(ctx context.Context, d *models.Domain) error {
	query := `
		UPDATE domains
		SET domain = $1,
		    ssl_update_date = $2,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING updated_at
	`

	row := r.db.QueryRow(ctx, query, d.Domain, d.SSLUpdateDate, d.ID)

	err := row.Scan(&d.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError

		// Unique violation
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "domains_domain_key" {
				return errors.New("another record already uses this domain")
			}
		}

		if err == pgx.ErrNoRows {
			return errors.New("domain not found")
		}

		return err
	}

	return nil
}

// UpdateDomainName updates only the domain field of a domain record
func (r *DomainRepo) UpdateDomainName(ctx context.Context, id int64, newDomain string) (time.Time, error) {
	query := `
		UPDATE domains
		SET domain = $1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING updated_at
	`

	var updatedAt time.Time

	row := r.db.QueryRow(ctx, query, newDomain, id)

	err := row.Scan(&updatedAt)
	if err != nil {
		var pgErr *pgconn.PgError

		// Unique violation
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "domains_domain_key" {
				return time.Time{}, errors.New("another record already uses this domain")
			}
		}

		if err == pgx.ErrNoRows {
			return time.Time{}, errors.New("domain not found")
		}

		return time.Time{}, err
	}

	return updatedAt, nil
}

// DeleteDomain deletes a domain by ID
func (r *DomainRepo) DeleteDomain(ctx context.Context, id int64) error {
	query := `DELETE FROM domains WHERE id = $1`

	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return errors.New("domain not found")
	}

	return nil
}

// ListDomains returns all domains
func (r *DomainRepo) ListDomains(ctx context.Context) ([]*models.Domain, error) {
	query := `
		SELECT id, domain, ssl_update_date, created_at, updated_at
		FROM domains
		ORDER BY id DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.Domain

	for rows.Next() {
		var d models.Domain

		err := rows.Scan(
			&d.ID,
			&d.Domain,
			&d.SSLUpdateDate,
			&d.CreatedAt,
			&d.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, &d)
	}

	return items, nil
}
