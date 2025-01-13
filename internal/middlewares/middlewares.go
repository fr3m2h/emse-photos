package middlewares

import (
	"context"
	"net/http"
	"photos/internal/db/query"
	"photos/internal/handlers"
	"time"
)

// MaxBodySize creates a middleware that limits the size of the request body.
// Requests exceeding the specified size will result in the connection being closed.
// This middleware is useful for preventing excessive memory consumption or abuse.
func MaxBodySize(size int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r2 := *r
			r2.Body = http.MaxBytesReader(w, r.Body, size)
			next.ServeHTTP(w, &r2)
		})
	}
}

// AuthRestricted creates a middleware that restricts access to authenticated users only.
// It verifies the presence and validity of a session cookie. If the session token is missing, invalid,
// or expired, the middleware redirects the user to the landing page. Valid session tokens are added
// to the request context for subsequent use.
func AuthRestricted(cfg handlers.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cfg.Security.Session.CookieName)
			if err != nil {
				redirectToLanding(w, r, cfg)
				return
			}
			var data map[string]string
			err = cfg.Security.Session.SecureCookie.Decode(cfg.Security.Session.CookieName, cookie.Value, &data)
			if err != nil {
				redirectToLanding(w, r, cfg)
				return
			}
			sessionToken, ok := data[cfg.Security.Session.CookieName]
			if !ok || sessionToken == "" {
				redirectToLanding(w, r, cfg)
				return
			}
			session, err := cfg.DB.GetSessionWithToken(r.Context(), sessionToken)
			if err != nil {
				redirectToLanding(w, r, cfg)
				return
			}
			if session.CreationDate.Add(cfg.Security.Session.CookieMaxAge).Before(time.Now()) {
				redirectToLanding(w, r, cfg)
				return
			}
			userInfo, err := cfg.DB.GetUserWithSession(r.Context(), sessionToken)
			if err != nil {
				redirectToLanding(w, r, cfg)
				return
			}
			ctx := context.WithValue(r.Context(), cfg.Security.Session.CookieName, sessionToken)
			ctx = context.WithValue(ctx, "userInfo", userInfo)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// AdminRestricted creates a middleware that restricts access to administrators only.
// If the user is not an administrator the request is rejected.
// AuthRestricted must be applied before this middleware to ensure the session is authenticated.
func AdminRestricted(cfg handlers.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userInfo := r.Context().Value(cfg.Security.Session.CookieName).(query.User)
			if !userInfo.IsAdmin {
				handlers.RespondWithMessage(w, "Sorry you're not an admin", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// redirectToLanding clears the session cookie and redirects the user to the landing page.
// This function is used when authentication fails, ensuring the session is invalidated
// and the user is directed to the default entry point.
func redirectToLanding(w http.ResponseWriter, r *http.Request, cfg handlers.Config) {
	http.SetCookie(w, &http.Cookie{
		Name:   cfg.Security.Session.CookieName,
		MaxAge: -1,
	})
	http.Redirect(w, r, cfg.Routes.Landing, http.StatusFound)
}
