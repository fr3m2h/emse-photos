package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"photos/internal/db/query"
	"strconv"
	"time"
)

func (cfg Config) ServePhotosPage(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(r.URL.Query().Get("event_id"))
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("Could not parse event_id param: %s", err), http.StatusInternalServerError)
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		offset = 0
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 10
	}
	if limit >= 20 {
		limit = 20
	}

	photos, err := cfg.DB.DB.GetPhotosByEventIDWithPagination(context.Background(), query.GetPhotosByEventIDWithPaginationParams{
		EventID: uint32(eventID),
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		RespondWithMessage(w, fmt.Sprintf("DB Failure: %s", err), http.StatusInternalServerError)
		return
	}

	if len(photos) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	data := map[string]interface{}{
		"Photos":     photos,
		"EventID":    eventID,
		"NextOffset": offset + limit, // Calculate the next offset
		"Limit":      limit,          // Keep the same limit
	}

	err = cfg.Templates.ExecuteTemplate(w, "photos.html", data)
	if err != nil {
		RespondWithMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (cfg Config) PhotoHandler(w http.ResponseWriter, r *http.Request) {
	// Get the "path" query parameter
	photoPath := r.URL.Path[len("/photos_dir/"):]
	if photoPath == "" {
		http.Error(w, "Missing 'path' query parameter", http.StatusBadRequest)
		return
	}
	// Build the full path to the photo
	fullPath := filepath.Join(cfg.PhotosDir, photoPath)

	// Check if the file exists and is not a directory
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) || info.IsDir() {
		RespondWithMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Serve the file
	http.ServeFile(w, r, fullPath)
}
func (cfg Config) UploadPhotosHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseMultipartForm(cfg.Server.MaxBodySize); err != nil { // 100 MB limit
		http.Error(w, "Failed to parse form data "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get the event ID
	eventIDStr := r.FormValue("event_id")
	if eventIDStr == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	// Validate the event ID
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil || eventID <= 0 {
		http.Error(w, "Invalid Event ID", http.StatusBadRequest)
		return
	}

	// Get the uploaded files
	files := r.MultipartForm.File["photos"]
	if len(files) == 0 {
		http.Error(w, "No photos uploaded", http.StatusBadRequest)
		return
	}

	// Start a transaction
	tx, err := cfg.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start database transaction", http.StatusInternalServerError)
		return
	}

	// Rollback on failure
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}()

	// Process each uploaded file
	for _, fileHeader := range files {
		// Open the uploaded file
		file, err := fileHeader.Open()
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to open uploaded file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Generate a unique file name
		fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileHeader.Filename)
		filePath := filepath.Join(cfg.PhotosDir, fileName)

		// Save the file to the configured directory
		outFile, err := os.Create(filePath)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, file)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Failed to write file to disk", http.StatusInternalServerError)
			return
		}

		err = cfg.DB.DB.CreatePhoto(ctx, query.CreatePhotoParams{
			PathToPhoto: filePath,
			EventID:     uint32(eventID),
		})
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit database transaction", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s?event_id=%d", cfg.Routes.Event, eventID), http.StatusSeeOther)
}
