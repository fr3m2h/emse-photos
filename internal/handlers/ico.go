package handlers

import "net/http"

func ServeFaviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "assets/img/favicon.ico")
}
