package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"on-my-interview/server/internal/http/handlers"
)

type Dependencies struct {
	JobService   handlers.JobService
	QueryService handlers.QueryService
}

func NewRouter(deps ...Dependencies) http.Handler {
	router := chi.NewRouter()
	router.Get("/health", handlers.Health)
	var apiDeps handlers.Dependencies
	if len(deps) > 0 {
		apiDeps = handlers.Dependencies{
			JobService:   deps[0].JobService,
			QueryService: deps[0].QueryService,
		}
	}
	api := handlers.NewAPI(apiDeps)
	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/crawl/jobs", func(r chi.Router) {
			r.Post("/", api.CreateJob)
			r.Get("/", api.ListJobs)
			r.Get("/{jobID}", api.GetJob)
		})
		r.Route("/posts", func(r chi.Router) {
			r.Get("/", api.ListPosts)
			r.Get("/{platform}/{sourcePostID}", api.GetPost)
			r.Post("/{platform}/{sourcePostID}/reparse", api.ReparsePost)
		})
		r.Get("/questions", api.ListQuestions)
		r.Get("/companies", api.ListCompanies)
		r.Get("/usage/windows", api.ListUsageWindows)
	})
	return router
}
