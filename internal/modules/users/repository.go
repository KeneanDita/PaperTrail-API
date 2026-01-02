package users

import (
	"context"
	"database/sql"
)

type Repository interface {
	Create(ctx context.Context, email string, role string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	List(ctx context.Context) ([]User, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, email string, role string) (*User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`INSERT INTO users (email, role) VALUES ($1, COALESCE(NULLIF($2, ''), DEFAULT)) RETURNING public_id, email, role, created_at`,
		email,
		role,
	)

	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (*User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT public_id, email, role, created_at FROM users WHERE public_id = $1`, id)

	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT public_id, email, role, created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}

	return out, rows.Err()
}
