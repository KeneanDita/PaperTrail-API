package papers

import (
	"context"
	"time"

	"papertrail/internal/storage"
)

type Paper struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Abstract  string    `json:"abstract"`
	AuthorID  string    `json:"author_id"`
	PdfURL    string    `json:"pdf_url"`
	CreatedAt time.Time `json:"created_at"`
}

type Service interface {
	Create(ctx context.Context, p Paper) (*Paper, error)
	Get(ctx context.Context, id string) (*Paper, error)
	List(ctx context.Context) ([]Paper, error)
}

type service struct {
	repo    Repository
	storage *storage.SupabaseClient
}

func NewService(repo Repository, storage *storage.SupabaseClient) Service {
	return &service{repo: repo, storage: storage}
}

func (s *service) Create(ctx context.Context, p Paper) (*Paper, error) {
	return s.repo.Create(ctx, p)
}

func (s *service) Get(ctx context.Context, id string) (*Paper, error) {
	return s.repo.Get(ctx, id)
}

func (s *service) List(ctx context.Context) ([]Paper, error) {
	return s.repo.List(ctx)
}
