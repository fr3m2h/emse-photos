package handlers

import (
	"net/http"
)

func (cfg Config) ServeDashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	err := cfg.Templates.ExecuteTemplate(w, "dashboard.html", nil)
	if err != nil {
		RespondWithMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
