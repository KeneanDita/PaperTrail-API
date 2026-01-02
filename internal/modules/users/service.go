package users

import (
	"context"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Service interface {
	Get(ctx context.Context, id string) (*User, error)
	List(ctx context.Context) ([]User, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Get(ctx context.Context, id string) (*User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) List(ctx context.Context) ([]User, error) {
	return s.repo.List(ctx)
}
