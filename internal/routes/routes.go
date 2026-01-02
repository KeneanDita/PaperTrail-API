package routes

import (
	"database/sql"
	"net/http"
	"time"

	"papertrail/internal/config"
	"papertrail/internal/middleware"
	"papertrail/internal/modules/comments"
	"papertrail/internal/modules/papers"
	"papertrail/internal/modules/reviews"
	"papertrail/internal/modules/users"
	"papertrail/internal/storage"
	"papertrail/internal/utils"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(cfg *config.Config, db *sql.DB, supa *storage.SupabaseClient) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID, chimw.RealIP, chimw.Recoverer, middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"uptime":  time.Now().Unix(),
			"env":     cfg.Environment,
			"version": "v0.1.0",
		})
	})

	auth := middleware.Auth(cfg.JWTSecret)

	userRepo := users.NewPostgresRepository(db)
	userService := users.NewService(userRepo)
	userHandler := users.NewHandler(userService)

	paperRepo := papers.NewPostgresRepository(db)
	paperService := papers.NewService(paperRepo, supa)
	paperHandler := papers.NewHandler(paperService)

	reviewService := reviews.NewService(db)
	reviewHandler := reviews.NewHandler(reviewService)

	commentService := comments.NewService(db)
	commentHandler := comments.NewHandler(commentService)

	r.Route("/api", func(api chi.Router) {
		// Users are public for now to allow easy bootstrapping via Postman.
		userHandler.RegisterRoutes(api)

		api.Group(func(priv chi.Router) {
			priv.Use(auth)
			paperHandler.RegisterRoutes(priv)
			reviewHandler.RegisterRoutes(priv)
			commentHandler.RegisterRoutes(priv)
		})
	})

	return r
}
