package jobs

import (
	"context"
	"time"

	"on-my-interview/server/internal/storage/repository"
)

type Scheduler struct {
	service *Service
	ticker  *time.Ticker
	params  repository.CreateJobParams
	done    chan struct{}
}

func NewScheduler(service *Service, interval time.Duration, params repository.CreateJobParams) *Scheduler {
	return &Scheduler{
		service: service,
		ticker:  time.NewTicker(interval),
		params:  params,
		done:    make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				s.ticker.Stop()
				close(s.done)
				return
			case <-s.ticker.C:
				_, _ = s.service.CreateJob(ctx, s.params)
			}
		}
	}()
}

func (s *Scheduler) Done() <-chan struct{} {
	return s.done
}
