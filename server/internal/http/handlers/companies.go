package handlers

import "net/http"

func (a *API) ListCompanies(w http.ResponseWriter, r *http.Request) {
	if a.deps.QueryService == nil {
		writeError(w, http.StatusServiceUnavailable, "query service is not configured")
		return
	}
	companies, err := a.deps.QueryService.ListCompanies(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	responses := make([]companyResponse, 0, len(companies))
	for _, company := range companies {
		responses = append(responses, toCompanyResponse(company))
	}
	writeJSON(w, http.StatusOK, responses)
}
