package reviews

import (
	"context"
	"database/sql"
	"time"
)

type Review struct {
	ID         string    `json:"id"`
	PaperID    string    `json:"paper_id"`
	ReviewerID string    `json:"reviewer_id"`
	Rating     int       `json:"rating"`
	Comments   string    `json:"comments"`
	CreatedAt  time.Time `json:"created_at"`
}

type Service interface {
	ListByPaper(ctx context.Context, paperID string) ([]Review, error)
	Create(ctx context.Context, r Review) (*Review, error)
}

type service struct {
	db *sql.DB
}

func NewService(db *sql.DB) Service {
	return &service{db: db}
}

func (s *service) ListByPaper(ctx context.Context, paperID string) ([]Review, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, paper_id, reviewer_id, rating, comments, created_at FROM reviews WHERE paper_id = $1 ORDER BY created_at DESC`, paperID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Review
	for rows.Next() {
		var rev Review
		if err := rows.Scan(&rev.ID, &rev.PaperID, &rev.ReviewerID, &rev.Rating, &rev.Comments, &rev.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rev)
	}

	return out, rows.Err()
}

func (s *service) Create(ctx context.Context, r Review) (*Review, error) {
	row := s.db.QueryRowContext(ctx, `INSERT INTO reviews (paper_id, reviewer_id, rating, comments) VALUES ($1, $2, $3, $4) RETURNING id, created_at`, r.PaperID, r.ReviewerID, r.Rating, r.Comments)

	created := r
	if err := row.Scan(&created.ID, &created.CreatedAt); err != nil {
		return nil, err
	}

	return &created, nil
}
