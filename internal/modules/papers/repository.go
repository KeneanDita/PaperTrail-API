package papers

import (
	"context"
	"database/sql"
)

type Repository interface {
	Create(ctx context.Context, p Paper) (*Paper, error)
	Get(ctx context.Context, id string) (*Paper, error)
	List(ctx context.Context) ([]Paper, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, p Paper) (*Paper, error) {
	row := r.db.QueryRowContext(ctx, `INSERT INTO papers (title, abstract, author_id, pdf_url) VALUES ($1, $2, $3, $4) RETURNING id, created_at`, p.Title, p.Abstract, p.AuthorID, p.PdfURL)

	created := p
	if err := row.Scan(&created.ID, &created.CreatedAt); err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (*Paper, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, title, abstract, author_id, pdf_url, created_at FROM papers WHERE id = $1`, id)

	var p Paper
	if err := row.Scan(&p.ID, &p.Title, &p.Abstract, &p.AuthorID, &p.PdfURL, &p.CreatedAt); err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]Paper, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, title, abstract, author_id, pdf_url, created_at FROM papers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Paper
	for rows.Next() {
		var p Paper
		if err := rows.Scan(&p.ID, &p.Title, &p.Abstract, &p.AuthorID, &p.PdfURL, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}

	return out, rows.Err()
}
