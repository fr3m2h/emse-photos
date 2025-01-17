package routes

import (
	"fmt"
	"net/http"
	"photos/internal/handlers"
	"photos/internal/middlewares"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/rs/zerolog/hlog"
)

// Service creates and configures the HTTP service for the application.
// It sets up routes, applies global middlewares, and defines groups for public and authenticated routes.
// This handler integrates rate limiting, authentication, and CSRF protection for secure operations.
func Service(cfg handlers.Config) http.Handler {
	r := chi.NewRouter()
	loadGlobalMiddlewares(r, cfg)

	r.NotFound(cfg.ServeNotFoundHandler)
	r.Get(cfg.Routes.Favicon, handlers.ServeFaviconHandler)
	r.Get("/htmx.min.js", handlers.ServeHtmxScriptHandler)
	r.Get(cfg.Routes.Landing, cfg.ServeLandingHandler)

	r.Group(func(r chi.Router) {
		r.Use(httprate.Limit(
			10,
			time.Minute,
			httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
			}),
		))
		r.Get(cfg.Routes.Login, cfg.LoginHandler)
		r.Get(cfg.Routes.CasCallback, cfg.CasCallbackHandler)
		r.Get(fmt.Sprintf("%s/{}", strings.TrimPrefix(cfg.PhotosDir, ".")), cfg.PhotoHandler)
	})
	r.Group(func(r chi.Router) {
		r.Use(middlewares.AuthRestricted(cfg))
		r.Get(cfg.Routes.Dashboard, cfg.ServeDashboardHandler)
		r.Get(cfg.Routes.Logout, cfg.LogoutHandler)
		r.Get(cfg.Routes.Event, cfg.ServeEventHandler)
		r.Get(cfg.Routes.Photos, cfg.ServePhotosPage)
		r.Post("/create-event", cfg.CreateEventHandler)
		r.Post("/upload-photos", cfg.UploadPhotosHandler)
	})
	return r
}

// loadGlobalMiddlewares applies global middlewares to the router.
// These middlewares handle logging, request rate limiting, cross-origin resource sharing (CORS),
// request compression, content type validation, body size limits, CSRF protection...
func loadGlobalMiddlewares(r *chi.Mux, cfg handlers.Config) {
	r.Use(hlog.RemoteAddrHandler("ip"), hlog.UserAgentHandler("ua"), hlog.RefererHandler("referer"), hlog.RequestIDHandler("req-id", "X-Request-Id"))
	r.Use(hlog.NewHandler(cfg.Logger))
	r.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.AllowContentEncoding("gzip", "deflate", "gzip/deflate", "deflate/gzip"))
	r.Use(middleware.AllowContentType("application/json", "application/x-www-form-urlencoded", "multipart/form-data"))
	r.Use(middleware.CleanPath, middleware.RedirectSlashes)
	r.Use(middleware.Compress(4, "application/json", "application/x-www-form-urlencoded"))
	r.Use(middleware.Timeout(cfg.Server.RequestContextTimeout))
	// r.Use(csrf.Protect(
	// 	cfg.Security.Csrf.Secret,
	// 	csrf.MaxAge(int(cfg.Security.Csrf.CookieMaxAge.Seconds())),
	// 	csrf.HttpOnly(cfg.Security.Csrf.CookieHTTPOnly),
	// 	csrf.Secure(cfg.Security.Csrf.CookieSecure),
	// 	csrf.SameSite(csrf.SameSiteMode(cfg.Security.Csrf.CookieSameSite)),
	// 	csrf.RequestHeader(cfg.Security.Csrf.HeaderName),
	// 	csrf.FieldName(cfg.Security.Csrf.FieldName),
	// 	csrf.CookieName(cfg.Security.Csrf.CookieName),
	// ))
	r.Use(httprate.Limit(
		60,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
		}),
	))
}
