package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"photos/pkg/handlers"
	"time"
)

func MaxBodySize(size int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r2 := *r
			r2.Body = http.MaxBytesReader(w, r.Body, size)
			next.ServeHTTP(w, &r2)
		})
	}
}

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
			ctx := context.WithValue(r.Context(), cfg.Security.Session.CookieName, sessionToken)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// Assumes that is used after AuthRestricted
func AdminRestricted(cfg handlers.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionToken := r.Context().Value(cfg.Security.Session.CookieName)
			userInfo, err := cfg.DB.GetUserWithSession(r.Context(), sessionToken.(string))
			if err != nil {
				handlers.RespondWithMessage(w, fmt.Sprintf("DB Failure: %v", err), http.StatusInternalServerError)
				return
			}
			if !userInfo.IsAdmin {
				handlers.RespondWithMessage(w, "Sorry you're not an admin", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func redirectToLanding(w http.ResponseWriter, r *http.Request, cfg handlers.Config) {
	http.SetCookie(w, &http.Cookie{
		Name:   cfg.Security.Session.CookieName,
		MaxAge: -1,
	})
	http.Redirect(w, r, cfg.Routes.Landing, http.StatusFound)
}
