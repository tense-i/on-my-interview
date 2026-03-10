package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"on-my-interview/server/internal/config"
	"on-my-interview/server/internal/crawler"
	"on-my-interview/server/internal/crawler/nowcoder"
	httpapi "on-my-interview/server/internal/http"
	"on-my-interview/server/internal/jobs"
	openai "on-my-interview/server/internal/llm/openai"
	"on-my-interview/server/internal/storage/mysql"
	"on-my-interview/server/internal/storage/repository"
)

type App struct {
	cfg        config.Config
	httpServer *http.Server
	db         *sql.DB
	jobs       *jobs.Service
	scheduler  *jobs.Scheduler
}

func New(cfg config.Config) (*App, error) {
	if cfg.MySQLDSN == "" {
		return nil, fmt.Errorf("INTERVIEW_CRAWLER_MYSQL_DSN is required")
	}

	db, err := mysql.Open(cfg.MySQLDSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	if db == nil {
		return nil, fmt.Errorf("mysql db is nil")
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	if err := mysql.ApplyMigrations(context.Background(), db); err != nil {
		return nil, fmt.Errorf("apply migrations: %w", err)
	}

	repo := mysql.NewRepository(db)
	crawlerRegistry := crawler.NewRegistry(nowcoder.NewClient(cfg.NowCoderBaseURL, nil))
	extractor := openai.NewExtractor(openai.Config{
		BaseURL: cfg.LLMBaseURL,
		APIKey:  cfg.LLMAPIKey,
		Model:   cfg.LLMModel,
	})
	jobService := jobs.NewService(repo, crawlerRegistry, extractor, cfg.LLMModel)

	server := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: httpapi.NewRouter(httpapi.Dependencies{
			JobService:   jobService,
			QueryService: repo,
		}),
	}

	var scheduler *jobs.Scheduler
	if cfg.SchedulerEnabled && len(cfg.SchedulerPlatforms) > 0 && len(cfg.SchedulerKeywords) > 0 {
		scheduler = jobs.NewScheduler(jobService, cfg.SchedulerInterval, repository.CreateJobParams{
			TriggerType: repository.JobTriggerScheduler,
			Platforms:   cfg.SchedulerPlatforms,
			Keywords:    cfg.SchedulerKeywords,
			Pages:       cfg.SchedulerPages,
		})
	}

	return &App{
		cfg:        cfg,
		httpServer: server,
		db:         db,
		jobs:       jobService,
		scheduler:  scheduler,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.jobs.Start(ctx)
	if a.scheduler != nil {
		a.scheduler.Start(ctx)
	}

	errCh := make(chan error, 1)

	go func() {
		errCh <- a.httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return a.httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
