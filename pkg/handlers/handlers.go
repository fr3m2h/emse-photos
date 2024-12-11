package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"photos/pkg/config"
)

type Config config.Config

func RespondWithMessage(w http.ResponseWriter, error string, status int) {
	if status >= 500 {
		log.Printf("5xx error: %s", error)
		http.Error(w, "Internal server error", status)
		return
	}
	http.Error(w, error, status)
}

func renderTemplate(w http.ResponseWriter, t *template.Template, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	err := t.ExecuteTemplate(w, name, data)
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("error executing template: %v", err), http.StatusInternalServerError)
		return
	}
}
