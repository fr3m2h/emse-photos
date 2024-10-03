package handlers

import (
	"net/http"
)

func (cfg Config) ServeLandingPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	err := cfg.Templates.ExecuteTemplate(w, "landing.html", nil)
	if err != nil {
		respondWithMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
