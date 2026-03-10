package jobs

import "context"

type Worker struct {
	service *Service
}

func NewWorker(service *Service) *Worker {
	return &Worker{service: service}
}

func (w *Worker) RunJob(ctx context.Context, jobID int64) error {
	return w.service.RunJob(ctx, jobID)
}
