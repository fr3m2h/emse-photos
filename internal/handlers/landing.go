package handlers

import (
	"fmt"
	"net/http"
	"net/url"
)

func (cfg Config) ServeLandingPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	params := url.Values{}
	casUrlWithCallback := ""
	if cfg.DevMode.Enabled {
		params.Add("service", fmt.Sprintf("%s%s", cfg.BaseUrls.Dev.Service, cfg.Routes.CasCallback))
		casUrlWithCallback = fmt.Sprintf("%s?%s", cfg.BaseUrls.Dev.Cas, params.Encode())
	} else {
		params.Add("service", fmt.Sprintf("%s%s", cfg.BaseUrls.Prod.Service, cfg.Routes.CasCallback))
		casUrlWithCallback = fmt.Sprintf("%s?%s", cfg.BaseUrls.Prod.Cas, params.Encode())
	}

	err := cfg.Templates.ExecuteTemplate(w, "landing.html", struct{ CAS_URL_WITH_CALLBACK string }{CAS_URL_WITH_CALLBACK: casUrlWithCallback})
	if err != nil {
		respondWithMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
