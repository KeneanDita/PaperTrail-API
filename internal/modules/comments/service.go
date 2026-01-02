package comments

import (
	"context"
	"database/sql"
	"time"
)

type Comment struct {
	ID        string    `json:"id"`
	PaperID   string    `json:"paper_id"`
	UserID    string    `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type Service interface {
	ListByPaper(ctx context.Context, paperID string) ([]Comment, error)
	Create(ctx context.Context, c Comment) (*Comment, error)
}

type service struct {
	db *sql.DB
}

func NewService(db *sql.DB) Service {
	return &service{db: db}
}

func (s *service) ListByPaper(ctx context.Context, paperID string) ([]Comment, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, paper_id, user_id, body, created_at FROM comments WHERE paper_id = $1 ORDER BY created_at ASC`, paperID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.PaperID, &c.UserID, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}

	return out, rows.Err()
}

func (s *service) Create(ctx context.Context, c Comment) (*Comment, error) {
	row := s.db.QueryRowContext(ctx, `INSERT INTO comments (paper_id, user_id, body) VALUES ($1, $2, $3) RETURNING id, created_at`, c.PaperID, c.UserID, c.Body)

	created := c
	if err := row.Scan(&created.ID, &created.CreatedAt); err != nil {
		return nil, err
	}

	return &created, nil
}
