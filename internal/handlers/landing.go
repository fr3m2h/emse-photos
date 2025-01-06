package handlers

import (
	"net/http"
)

func (cfg Config) ServeLandingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	err := cfg.Templates.ExecuteTemplate(w, "landing.html", struct{ LOGIN_ROUTE string }{LOGIN_ROUTE: cfg.Routes.Login})
	if err != nil {
		RespondWithMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
