package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"on-my-interview/server/internal/storage/repository"
)

func (a *API) ListPosts(w http.ResponseWriter, r *http.Request) {
	if a.deps.QueryService == nil {
		writeError(w, http.StatusServiceUnavailable, "query service is not configured")
		return
	}
	posts, err := a.deps.QueryService.ListPosts(r.Context(), repository.PostFilter{
		Platform: r.URL.Query().Get("platform"),
		Company:  r.URL.Query().Get("company"),
		Tag:      r.URL.Query().Get("tag"),
		Limit:    parseInt(r.URL.Query().Get("limit"), 50),
		Offset:   parseInt(r.URL.Query().Get("offset"), 0),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	responses := make([]postDetailResponse, 0, len(posts))
	for _, post := range posts {
		responses = append(responses, toPostDetailResponse(post))
	}
	writeJSON(w, http.StatusOK, responses)
}

func (a *API) GetPost(w http.ResponseWriter, r *http.Request) {
	if a.deps.QueryService == nil {
		writeError(w, http.StatusServiceUnavailable, "query service is not configured")
		return
	}
	post, err := a.deps.QueryService.GetPostBySource(r.Context(), chi.URLParam(r, "platform"), chi.URLParam(r, "sourcePostID"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toPostDetailResponse(post))
}

func (a *API) ReparsePost(w http.ResponseWriter, r *http.Request) {
	if a.deps.JobService == nil {
		writeError(w, http.StatusServiceUnavailable, "job service is not configured")
		return
	}
	err := a.deps.JobService.ReparsePost(r.Context(), repository.ReparsePostParams{
		Platform:     chi.URLParam(r, "platform"),
		SourcePostID: chi.URLParam(r, "sourcePostID"),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}
