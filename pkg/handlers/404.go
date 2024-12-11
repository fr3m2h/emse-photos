package handlers

import "net/http"

func (cfg Config) ServeNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusNotFound)

	err := cfg.Templates.ExecuteTemplate(w, "404.html", nil)
	if err != nil {
		RespondWithMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
