package handlers

import (
	"fmt"
	"net/http"
	"photos/internal/db/query"
	"time"

	"github.com/gorilla/csrf"
)

func (cfg Config) ServeDashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userInfo := ctx.Value("userInfo").(query.User)
	events, err := cfg.DB.DB.GetEvents(ctx)
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("DB Failure: %s", err), http.StatusInternalServerError)
		return
	}
	csrfToken := csrf.Token(r)
	now := time.Now()
	defaultDate := now.Format("2006-01-02T15:04") // Proper datetime-local format
	w.Header().Set("Content-Type", "text/html")
	err = cfg.Templates.ExecuteTemplate(w, "dashboard.html", map[string]interface{}{"Events": events, "UserInfo": userInfo, "CSRF_TOKEN": csrfToken, "DefaultDate": defaultDate})
	if err != nil {
		RespondWithMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
