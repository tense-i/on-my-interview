package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"on-my-interview/server/internal/storage/repository"
)

type API struct {
	deps Dependencies
}

func NewAPI(deps Dependencies) *API {
	return &API{deps: deps}
}

type createJobRequest struct {
	Platforms    []string `json:"platforms"`
	Keywords     []string `json:"keywords"`
	Pages        int      `json:"pages"`
	ForceReparse bool     `json:"force_reparse"`
}

type jobResponse struct {
	ID           int64                 `json:"id"`
	TriggerType  repository.JobTrigger `json:"trigger_type"`
	Status       repository.JobStatus  `json:"status"`
	Platforms    []string              `json:"platforms"`
	Keywords     []string              `json:"keywords"`
	Pages        int                   `json:"pages"`
	ForceReparse bool                  `json:"force_reparse"`
	StatsJSON    string                `json:"stats_json,omitempty"`
	ErrorMessage string                `json:"error_message,omitempty"`
	StartedAt    *time.Time            `json:"started_at,omitempty"`
	FinishedAt   *time.Time            `json:"finished_at,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

func (a *API) CreateJob(w http.ResponseWriter, r *http.Request) {
	if a.deps.JobService == nil {
		writeError(w, http.StatusServiceUnavailable, "job service is not configured")
		return
	}

	var req createJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	job, err := a.deps.JobService.CreateJob(r.Context(), repository.CreateJobParams{
		TriggerType:  repository.JobTriggerManual,
		Platforms:    req.Platforms,
		Keywords:     req.Keywords,
		Pages:        req.Pages,
		ForceReparse: req.ForceReparse,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, toJobResponse(job))
}

func (a *API) ListJobs(w http.ResponseWriter, r *http.Request) {
	if a.deps.QueryService == nil {
		writeError(w, http.StatusServiceUnavailable, "query service is not configured")
		return
	}
	jobs, err := a.deps.QueryService.ListJobs(r.Context(), repository.ListJobsFilter{
		Limit:  parseInt(r.URL.Query().Get("limit"), 50),
		Offset: parseInt(r.URL.Query().Get("offset"), 0),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	responses := make([]jobResponse, 0, len(jobs))
	for _, job := range jobs {
		responses = append(responses, toJobResponse(job))
	}
	writeJSON(w, http.StatusOK, responses)
}

func (a *API) GetJob(w http.ResponseWriter, r *http.Request) {
	if a.deps.QueryService == nil {
		writeError(w, http.StatusServiceUnavailable, "query service is not configured")
		return
	}
	jobID, err := strconv.ParseInt(chi.URLParam(r, "jobID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job id")
		return
	}
	job, err := a.deps.QueryService.GetJob(r.Context(), jobID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toJobResponse(job))
}

func toJobResponse(job repository.CrawlJob) jobResponse {
	return jobResponse{
		ID:           job.ID,
		TriggerType:  job.TriggerType,
		Status:       job.Status,
		Platforms:    append([]string(nil), job.Platforms...),
		Keywords:     append([]string(nil), job.Keywords...),
		Pages:        job.Pages,
		ForceReparse: job.ForceReparse,
		StatsJSON:    job.StatsJSON,
		ErrorMessage: job.ErrorMessage,
		StartedAt:    job.StartedAt,
		FinishedAt:   job.FinishedAt,
		CreatedAt:    job.CreatedAt,
		UpdatedAt:    job.UpdatedAt,
	}
}
