package handlers

import (
	"net/http"

	"on-my-interview/server/internal/storage/repository"
)

func (a *API) ListUsageWindows(w http.ResponseWriter, r *http.Request) {
	if a.deps.QueryService == nil {
		writeError(w, http.StatusServiceUnavailable, "query service is not configured")
		return
	}

	from, err := parseRFC3339(r.URL.Query().Get("from"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid from")
		return
	}
	to, err := parseRFC3339(r.URL.Query().Get("to"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid to")
		return
	}

	result, err := a.deps.QueryService.ListUsageWindows(r.Context(), repository.UsageWindowFilter{
		Limit:  parseInt(r.URL.Query().Get("limit"), 50),
		Offset: parseInt(r.URL.Query().Get("offset"), 0),
		From:   from,
		To:     to,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toUsageWindowListResponse(result))
}
