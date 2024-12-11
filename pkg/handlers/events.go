package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Used after AdminRestricted
func (cfg Config) CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	params := url.Values{}
	casLoginUrlWithCallback := ""
	if cfg.DevMode.Enabled {
		params.Add("service", fmt.Sprintf("%s%s", cfg.BaseURLs.Dev.Service, cfg.Routes.CasCallback))
		casLoginUrlWithCallback = fmt.Sprintf("%s/login?%s", cfg.BaseURLs.Dev.Cas, params.Encode())
	} else {
		params.Add("service", fmt.Sprintf("%s%s", cfg.BaseURLs.Prod.Service, cfg.Routes.CasCallback))
		casLoginUrlWithCallback = fmt.Sprintf("%s/login?%s", cfg.BaseURLs.Prod.Cas, params.Encode())
	}

	cookie, err := r.Cookie(cfg.Security.Session.CookieName)
	if err != nil {
		http.Redirect(w, r, casLoginUrlWithCallback, http.StatusFound)
		return
	}
	var data map[string]string
	err = cfg.Security.Session.SecureCookie.Decode(cfg.Security.Session.CookieName, cookie.Value, &data)
	if err != nil {
		http.Redirect(w, r, casLoginUrlWithCallback, http.StatusFound)
		return
	}
	sessionToken, ok := data[cfg.Security.Session.CookieName]
	if !ok || sessionToken == "" {
		http.Redirect(w, r, casLoginUrlWithCallback, http.StatusFound)
		return
	}
	session, err := cfg.DB.GetSessionWithToken(r.Context(), sessionToken)
	if err != nil {
		log.Printf("DB Failure: %v", err)
		http.Redirect(w, r, casLoginUrlWithCallback, http.StatusFound)
		return
	}
	if session.CreationDate.Add(cfg.Security.Session.CookieMaxAge).Before(time.Now()) {
		http.Redirect(w, r, casLoginUrlWithCallback, http.StatusFound)
		return
	}
	http.Redirect(w, r, cfg.Routes.Dashboard, http.StatusFound)
}
