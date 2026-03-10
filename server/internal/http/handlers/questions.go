package handlers

import (
	"net/http"

	"on-my-interview/server/internal/storage/repository"
)

func (a *API) ListQuestions(w http.ResponseWriter, r *http.Request) {
	if a.deps.QueryService == nil {
		writeError(w, http.StatusServiceUnavailable, "query service is not configured")
		return
	}
	questions, err := a.deps.QueryService.ListQuestions(r.Context(), repository.QuestionFilter{
		Platform: r.URL.Query().Get("platform"),
		Company:  r.URL.Query().Get("company"),
		Tag:      r.URL.Query().Get("tag"),
		Query:    r.URL.Query().Get("query"),
		Limit:    parseInt(r.URL.Query().Get("limit"), 50),
		Offset:   parseInt(r.URL.Query().Get("offset"), 0),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	responses := make([]questionResponse, 0, len(questions))
	for _, question := range questions {
		responses = append(responses, toQuestionResponse(question))
	}
	writeJSON(w, http.StatusOK, responses)
}
