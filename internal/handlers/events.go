package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"photos/internal/db/query"
	"strconv"
	"time"

	"github.com/gorilla/csrf"
)

func (cfg Config) ServeEventHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventID, err := strconv.Atoi(r.FormValue("event_id"))
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("Could not parse event_id param: %s", err), http.StatusInternalServerError)
		return
	}

	events, err := cfg.DB.DB.GetEvents(ctx)
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("DB Failure: %s", err), http.StatusInternalServerError)
		return
	}

	csrfToken := csrf.Token(r)
	userInfo := ctx.Value("userInfo").(query.User)
	// Filter the main event and its child events
	var mainEvent query.Event
	childEvents := make([]query.Event, 0)
	eventExists := false

	for _, e := range events {
		if e.EventID == uint32(eventID) {
			mainEvent = e
			eventExists = true
		}
		if e.ParentEventID.Valid && e.ParentEventID.Int32 == int32(eventID) {
			childEvents = append(childEvents, e)
		}
	}

	// If the main event doesn't exist, respond with an error
	if !eventExists {
		RespondWithMessage(w, "event_id does not correspond to any existing event", http.StatusBadRequest)
		return
	}

	photos, err := cfg.DB.DB.GetPhotosByEventID(ctx, mainEvent.EventID)
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("DB Failure: %s", err), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	defaultDate := now.Format("2006-01-02T15:04") // Proper datetime-local format
	// Prepare the data for the template
	data := map[string]interface{}{
		"Event":       mainEvent,
		"ChildEvents": childEvents,
		"Photos":      photos,
		"UserInfo":    userInfo,
		"CSRF_TOKEN":  csrfToken,
		"DefaultDate": defaultDate,
		"ParentID":    eventID,
	}

	w.Header().Set("Content-Type", "text/html")
	err = cfg.Templates.ExecuteTemplate(w, "event.html", data)
	if err != nil {
		RespondWithMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (cfg *Config) CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Parse form values
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	eventName := r.FormValue("event_name")
	eventDescription := r.FormValue("event_description")
	eventDate := r.FormValue("event_date")
	eventParentID := r.FormValue("event_parentID")

	isEventParentIDNotNil := false
	var eventParentIDConverted int
	if eventParentID != "" {
		eventParentIDConverted, err = strconv.Atoi(eventParentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		isEventParentIDNotNil = true
	}

	if eventName == "" || eventDescription == "" || eventDate == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Convert eventDate to time.Time
	parsedEventDate, err := time.Parse("2006-01-02T15:04", eventDate)
	if err != nil {
		http.Error(w, "Invalid event date format", http.StatusBadRequest)
		return
	}

	err = cfg.DB.DB.CreateEvent(ctx, query.CreateEventParams{
		Name:        eventName,
		Description: eventDescription,
		EventDate:   parsedEventDate,
		ParentEventID: sql.NullInt32{
			Valid: isEventParentIDNotNil,
			Int32: int32(eventParentIDConverted),
		},
	})
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("DB Failure: %v", err), http.StatusInternalServerError)
		return
	}
	if isEventParentIDNotNil {
		http.Redirect(w, r, fmt.Sprintf("%s?event_id=%d", cfg.Routes.Event, eventParentIDConverted), http.StatusSeeOther)

	} else {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}
